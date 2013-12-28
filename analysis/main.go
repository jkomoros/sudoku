package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/skelterjohn/go.matrix"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"log"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
)

const _DB_CONFIG_FILENAME = "db_config.SECRET.json"
const _OUTPUT_FILENAME = "output.csv"
const _QUERY_LIMIT = 100
const _PENALTY_PERCENTAGE_CUTOFF = 0.10

//How many solves a user must have to have their relative scale included.
//A low value gives you far more very low or very high scores than you shoul get.
const _MINIMUM_SOLVES = 10

var noLimitFlag bool
var printPuzzleDataFlag bool
var cullCheaterPercentageFlag float64
var minimumSolvesFlag int
var useMockData bool
var queryLimit int
var verbose bool

func init() {
	flag.BoolVar(&noLimitFlag, "a", false, "Specify to execute the solves query with no limit.")
	flag.BoolVar(&printPuzzleDataFlag, "p", false, "Specify that you want puzzle data printed out in the output.")
	flag.Float64Var(&cullCheaterPercentageFlag, "c", _PENALTY_PERCENTAGE_CUTOFF, "What percentage of solve time must be penalty for someone to be considered a cheater.")
	flag.IntVar(&minimumSolvesFlag, "s", _MINIMUM_SOLVES, "How many solves a user must have their scores considered.")
	flag.IntVar(&queryLimit, "n", _QUERY_LIMIT, "Number of solves to fetch from the database.")
	flag.BoolVar(&useMockData, "m", false, "Use mock data (useful if you don't have a real database to test with).")
	flag.BoolVar(&verbose, "v", false, "Verbose mode.")

	//We're going to be doing some heavy-duty matrix multiplication, and the matrix package can take advantage of multiple cores.
	runtime.GOMAXPROCS(6)
}

type dbConfig struct {
	Url               string
	Username          string
	Password          string
	DbName            string
	SolvesTable       string
	SolvesID          string
	SolvesPuzzleID    string
	SolvesTotalTime   string
	SolvesPenaltyTime string
	SolvesUser        string
	PuzzlesTable      string
	PuzzlesID         string
	PuzzlesDifficulty string
	PuzzlesName       string
	PuzzlesPuzzle     string
}

var config dbConfig

type solve struct {
	puzzleID    int
	totalTime   int
	penaltyTime int
}

type userSolvesCollection struct {
	solves     []solve
	max        int
	min        int
	idPosition map[int]int
}

type puzzle struct {
	id                     int
	userRelativeDifficulty float64
	difficultyRating       int
	name                   string
	puzzle                 string
}

type puzzles []*puzzle

type byUserRelativeDifficulty struct {
	puzzles
}

func (self puzzles) Len() int {
	return len(self)
}

func (self puzzles) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self byUserRelativeDifficulty) Less(i, j int) bool {
	return self.puzzles[i].userRelativeDifficulty < self.puzzles[j].userRelativeDifficulty
}

type bySolveTimeDsc []solve

func (self bySolveTimeDsc) Len() int {
	return len(self)
}

func (self bySolveTimeDsc) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self bySolveTimeDsc) Less(i, j int) bool {
	//For the purposes of this algorithm, the "best" has to be lowest rank.
	return self[i].totalTime > self[j].totalTime
}

func (self *userSolvesCollection) addSolve(solve solve) bool {
	//Cull obviously incorrect solves.
	if solve.totalTime == 0 {
		return false
	}

	//Cull solves that leaned too heavily on hints.
	if float64(solve.penaltyTime)/float64(solve.totalTime) > cullCheaterPercentageFlag {
		return false
	}

	self.solves = append(self.solves, solve)
	if len(self.solves) == 1 {
		self.max = solve.totalTime
		self.min = solve.totalTime
	} else {
		if self.max < solve.totalTime {
			self.max = solve.totalTime
		}
		if self.min > solve.totalTime {
			self.min = solve.totalTime
		}
	}
	return true
}

//Whehter or not this should be included in calculation.
//Basically, whether the reltaiveDifficulties will all be valid.
//Normally this returns false if there is only one solve by the user, but could also
//happen when there are multiple solves but (crazily enough) they all have exactly the same solveTime.
//This DOES happen in the production dataset.
func (self *userSolvesCollection) valid() bool {
	if self.max == self.min {
		return false
	}

	if len(self.solves) < minimumSolvesFlag {
		return false
	}

	return true
}

func (self *userSolvesCollection) relativeDifficulties() map[int]float32 {
	//Returns a map of puzzle id to relative difficulty, normalized by our max and min.
	avgSolveTimes := make(map[int]float32)
	//Keep track of how many times we've seen each puzzle solved by this user so we can do correct averaging.
	avgSolveTimesCount := make(map[int]int)

	//First, collect the average solve time (in case the same user has solved more than once the same puzzle)

	for _, solve := range self.solves {
		currentAvgSolveTime := avgSolveTimes[solve.puzzleID]

		avgSolveTimes[solve.puzzleID] = (currentAvgSolveTime*float32(avgSolveTimesCount[solve.puzzleID]) + float32(solve.totalTime)) / float32(avgSolveTimesCount[solve.puzzleID]+1)

		avgSolveTimesCount[solve.puzzleID]++
	}

	//Now, relativize all of the scores.

	result := make(map[int]float32)

	for puzzleID, avgSolveTime := range avgSolveTimes {
		result[puzzleID] = (avgSolveTime - float32(self.min)) / float32(self.max-self.min)
	}

	return result
}

func main() {

	flag.Parse()

	//Load up the Database config.
	file, err := os.Open(_DB_CONFIG_FILENAME)
	if err != nil {
		log.Fatal("Could not find the config file at ", _DB_CONFIG_FILENAME, ". You should copy the SAMPLE one to that filename and configure.")
		os.Exit(1)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("There was an error parsing JSON from the config file: ", err)
		os.Exit(1)
	}

	difficutlyRatingsChan := make(chan map[int]puzzle)

	//Go fetch the difficulties for each puzzle; we'll need this data at the end.
	go getPuzzleDifficultyRatings(difficutlyRatingsChan)

	var db mysql.Conn

	//Should we use local mock data or actually hit the server?
	if useMockData {
		db = &mockConnection{}
	} else {
		db = mysql.New("tcp", "", config.Url, config.Username, config.Password, config.DbName)
	}

	if err := db.Connect(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var solvesQuery string

	if noLimitFlag {
		log.Println("Running without a limit for number of solves to retrieve.")
		solvesQuery = "select %s, %s, %s, %s from %s"
	} else {
		log.Println("Running with a limit of ", queryLimit, " for number of solves to retrieve.")
		solvesQuery = "select %s, %s, %s, %s from %s limit " + strconv.Itoa(queryLimit)
	}

	res, err := db.Start(solvesQuery, config.SolvesUser, config.SolvesPuzzleID, config.SolvesTotalTime, config.SolvesPenaltyTime, config.SolvesTable)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	solvesByUser := make(map[string]*userSolvesCollection)

	var userSolves *userSolvesCollection
	var ok bool
	var i int
	var skippedSolves int
	var skippedDuplicateSolves int

	//We want to skip rows we've already seen; we'll keep a set of used rows in here.
	//This is necessary because some solves in the production database appear to be dupes.
	seenRows := make(map[string]bool)

	//Conceptually all of the solves for a specific user show how hard the puzzles are from one user's perspective.
	//We'll use a markov chain-based analysis later to extract an aggregated rank that merges every user's perspective, to give a global perpsective.

	//First we will visit every solve record and collect them into a collection for each username.
	for {

		row, _ := res.GetRow()

		//Are we past the edge of the returned results?
		if row == nil {
			break
		}

		i++

		//Check to see if this was a row we've already seen
		rowHashValue := fmt.Sprintf("%v", row)
		if _, seen := seenRows[rowHashValue]; seen {
			skippedDuplicateSolves++
			continue
		} else {
			seenRows[rowHashValue] = true
		}

		//Is this the first time we've seen the user?
		userSolves, ok = solvesByUser[row.Str(0)]

		//This is the first time; initalize a new userSolvesCollection for them.
		if !ok {
			userSolves = new(userSolvesCollection)
			userSolves.idPosition = make(map[int]int)
			solvesByUser[row.Str(0)] = userSolves
		}

		//Add it to the collection. That method might decide to skip it.
		if !userSolves.addSolve(solve{row.Int(1), row.Int(2), row.Int(3)}) {
			skippedSolves++
		}

	}

	//Now that we've seen all of the solves, give some quick stats.
	log.Println("Processed", i, "solves by", len(solvesByUser), "users.")
	log.Println("Skipped", skippedSolves, "solves that cheated too much.")
	log.Println("Skipped", skippedDuplicateSolves, "solves because they were duplicates of solves seen earlier.")

	//Now get the relative difficulty for each user's puzzles, and collect them.

	relativeDifficultiesByPuzzle := make(map[int][]float32)

	//Later we'll need to grab all of the userCollections that reference a given puzzle.
	//As we traverse through the userSolveCollections now we'll build that reference up.
	//This is a map of puzzles to a set of which collections include it.
	collectionByPuzzle := make(map[int]map[*userSolvesCollection]bool)

	var skippedUsers int

	for _, collection := range solvesByUser {

		/*
			//TODO: consider removing this logic and all of skipped users totally.
				if !collection.valid() {
					skippedUsers++
					continue
				}
		*/

		//This next section is to be removed.
		for puzzleID, relativeDifficulty := range collection.relativeDifficulties() {
			relativeDifficultiesByPuzzle[puzzleID] = append(relativeDifficultiesByPuzzle[puzzleID], relativeDifficulty)
		}

		//Now that we have all of the solves for this user, we can sort them.
		//For the analysis we'll do later, a harder solve is ranked higher, and a higher rank is actually a LOW rank.
		sort.Sort(bySolveTimeDsc(collection.solves))

		for i, puzzle := range collection.solves {
			//Later in the analysis we'll need to know, given a collection and a puzzle id, what its rank within the collection is.
			collection.idPosition[puzzle.puzzleID] = i

			//Is this the first time we've seen this collection for this puzzle?
			collectionMap, ok := collectionByPuzzle[puzzle.puzzleID]
			//Yup, it is. Create a new one.
			if !ok {
				collectionMap = make(map[*userSolvesCollection]bool)
			}

			//Store that we are in the set of collections.
			collectionMap[collection] = true

			//Store that set back in the larger map.
			collectionByPuzzle[puzzle.puzzleID] = collectionMap
		}

	}

	log.Println("Skipped ", skippedUsers, " users because they did not have enough solve times.")

	//Now, create the Markov Transition Matrix, according to algorithm MC4 of http://www.wisdom.weizmann.ac.il/~naor/PAPERS/rank_www10.html
	//The relevant part of the algorithm, from that source:
	/*

		If the current state is page P, then the next state is chosen as follows: first pick a page Q uniformly from the union of all pages ranked by the search engines. If t(Q) < t(P) for a majority of the lists t that ranked both P and Q, then go to Q, else stay in P.

	*/
	//We start by creating a stacked array of float64's that we'll pass to the matrix library.

	//This will be the dimensions of the matrix.
	numPuzzles := len(collectionByPuzzle)

	log.Println("Discovered", numPuzzles, "puzzles.")

	//Create the stacked array we'll stuff values into.
	matrixData := make([][]float64, numPuzzles)
	for i := range matrixData {
		matrixData[i] = make([]float64, numPuzzles)
	}

	//Now we will associate each observed puzzleID with an index that it will be associated with in the matrix.
	puzzleIndex := make([]int, numPuzzles)
	counter := 0
	for key, _ := range collectionByPuzzle {
		puzzleIndex[counter] = key
		counter++
	}

	//Now we start to build up the matrix according to the MC4 algorithm.
	if verbose {
		log.Println("Starting to build up matrix...")
	}

	//For each cell in the matrix (pairwise comparison of puzzles).
	for i := 0; i < numPuzzles; i++ {
		for j := 0; j < numPuzzles; j++ {

			if i == j {
				//The special case; stay in the same state. We'll treat it specially.
				continue
			}

			//Convert the zero-index into the puzzle ID we're actually interested in.
			p := puzzleIndex[i]
			q := puzzleIndex[j]

			//Find the intersection of userSolveCollections that contain both p and q.

			//First, grab the two sets.
			pMap := collectionByPuzzle[p]
			qMap := collectionByPuzzle[q]

			//Build the intersection.
			var intersection []*userSolvesCollection
			for collection, _ := range pMap {
				if _, ok := qMap[collection]; ok {
					intersection = append(intersection, collection)
				}
			}

			//Next, calculate how many of the collections have q ranked better (lower!) than p.
			//This is fast thanks to the earlier processing.
			count := 0
			for _, collection := range intersection {
				if collection.idPosition[q] < collection.idPosition[p] {
					count++
				}
			}

			//Is it a majority? if so, transition. if not, leave at 0.
			//These are just a sentinel. Later we'll go through and normalize probabilities.
			if count > (len(intersection) / 2) {
				matrixData[i][j] = 1.0
			}

		}
	}

	if verbose {
		log.Println("Normalizing matrix...")
	}

	//Go through and normalize the probabilities in each row to sum to 1.
	//We treat the no-movement case specially; it's 1 - the sum of the probabilties of going to other cells.
	for i := 0; i < numPuzzles; i++ {
		//Count the number of rows that are 1.0.
		count := 0
		for j := 0; j < numPuzzles; j++ {
			if matrixData[i][j] > 0.0 {
				count++
			}
		}
		//Each unit of probability is this size:
		probability := 1.0 / float64(numPuzzles)

		//Stuff in the final normalized values, treating i,i the same.
		for j := 0; j < numPuzzles; j++ {
			if i == j {
				//The stay in the same space probability
				matrixData[i][j] = float64(numPuzzles-count) * probability
			} else if matrixData[i][j] > 0.0 {
				matrixData[i][j] = probability
			}
		}
	}

	//Create an actual matrix with the data.
	markovChain := matrix.MakeDenseMatrixStacked(matrixData)

	if verbose {
		log.Println("Beginning matrix multiplication...")
	}

	//We want to find the stable distribution, so we will raise the matrix to repeatedly high powers.
	//Over time the matrix will stabalize, at that point every row will look similar to each other.
	//We'll check for the matrix stabalizing before the end and break early if it does.
	for i := 0; i < 20; i++ {
		markovChain = matrix.ParallelProduct(markovChain, markovChain)

		//Are the rows converged enough for us to bail?
		difference := 0.0
		for i := 0; i < numPuzzles; i++ {
			difference += math.Abs(markovChain.Get(0, i) - markovChain.Get(1, i))
		}
		if verbose {
			log.Println("Finished matrix multiplication #", i+1, ", with a difference of", difference)
		}
		if difference < 0.0001 {
			log.Println("The markov chain converged after", i+1, "mulitplications.")
			break
		}
	}

	//Now we'll merge in information about the puzzles

	//Collect the results of the second query we kicked off at the beginning,
	//with meta-information about each puzzle.
	difficultyRatings := <-difficutlyRatingsChan

	//This will be our final output.
	puzzles := make([]*puzzle, numPuzzles)

	for i := 0; i < numPuzzles; i++ {
		thePuzzle := new(puzzle)
		thePuzzle.id = puzzleIndex[i]
		//TODO: rename this field, it no longer reflects what it really is.
		thePuzzle.userRelativeDifficulty = markovChain.Get(0, i)
		info, ok := difficultyRatings[thePuzzle.id]
		if ok {
			//TODO: rename this field; it's no longer accurate.
			thePuzzle.difficultyRating = info.difficultyRating
			thePuzzle.name = info.name
			thePuzzle.puzzle = info.puzzle
		}
		puzzles[i] = thePuzzle
	}

	//Sort the puzzles by relative user difficulty
	//We actually don't need the wrapper, since it will modify the underlying slice.
	sort.Sort(byUserRelativeDifficulty{puzzles})

	//Now print the results to stdout.

	csvOut := csv.NewWriter(os.Stdout)

	for _, puzzle := range puzzles {
		temp := []string{strconv.Itoa(puzzle.id), strconv.Itoa(puzzle.difficultyRating), fmt.Sprintf("%g", puzzle.userRelativeDifficulty), puzzle.name}
		if printPuzzleDataFlag {
			temp = append(temp, puzzle.puzzle)
		}
		csvOut.Write(temp)
	}

	csvOut.Flush()

}

func getPuzzleDifficultyRatings(result chan map[int]puzzle) {

	var db mysql.Conn

	if useMockData {
		db = &mockConnection{}
	} else {
		db = mysql.New("tcp", "", config.Url, config.Username, config.Password, config.DbName)
	}

	if err := db.Connect(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	res, err := db.Start("select %s, %s, %s, %s from %s", config.PuzzlesID, config.PuzzlesDifficulty, config.PuzzlesName, config.PuzzlesPuzzle, config.PuzzlesTable)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	puzzles := make(map[int]puzzle)

	for {

		row, _ := res.GetRow()

		if row == nil {
			break
		}

		puzzles[row.Int(0)] = puzzle{id: row.Int(0), difficultyRating: row.Int(1), name: row.Str(2), puzzle: row.Str(3)}
	}

	result <- puzzles

}
