package main

import (
	"dokugen"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sajari/regression"
	"github.com/skelterjohn/go.matrix"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"log"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const _DB_CONFIG_FILENAME = "db_config.SECRET.json"
const _OUTPUT_FILENAME = "output.csv"
const _QUERY_LIMIT = 100
const _PENALTY_PERCENTAGE_CUTOFF = 0.01
const _MATRIX_DIFFERENCE_CUTOFF = 0.00001
const _MAX_MATRIX_POWER = 250

//Making this number much higher does not improve R2, even for very high numbers like 500.
const _NUMBER_OF_HUMAN_SOLVES = 10

const _NORMALIZED_LOWER_BOUND = 0.1
const _NORMALIZED_UPPER_BOUND = 0.9

//How many solves a user must have to have their relative scale included.
//A low value gives you far more very low or very high scores than you shoul get.
const _MINIMUM_SOLVES = 10

var noLimitFlag bool
var printPuzzleDataFlag bool
var cullCheaterPercentageFlag float64
var useMockData bool
var queryLimit int
var verbose bool
var minPuzzleCollections int
var calcWeights bool
var printPuzzleTechniques bool
var inputIsSolveData bool

func init() {
	flag.BoolVar(&noLimitFlag, "a", false, "Specify to execute the solves query with no limit.")
	flag.BoolVar(&printPuzzleDataFlag, "p", false, "Specify that you want puzzle data printed out in the output.")
	flag.Float64Var(&cullCheaterPercentageFlag, "c", _PENALTY_PERCENTAGE_CUTOFF, "What percentage of solve time must be penalty for someone to be considered a cheater.")
	flag.IntVar(&queryLimit, "n", _QUERY_LIMIT, "Number of solves to fetch from the database.")
	flag.BoolVar(&useMockData, "m", false, "Use mock data (useful if you don't have a real database to test with).")
	flag.BoolVar(&verbose, "v", false, "Verbose mode.")
	flag.IntVar(&minPuzzleCollections, "l", 10, "How many different user collections the puzzle must be included in for it to be included in the output.")
	//TODO: rationalize how you handle the three phases. It's seriously crazy the collection of flags and the spaghetti code.
	//For example, we should say that you can pass in a start phase and an input for that start phase, and then you saw the output phase you want to run to.
	flag.BoolVar(&calcWeights, "w", false, "Whether the output you want is to calculate technique weights")
	flag.BoolVar(&printPuzzleTechniques, "t", false, "If calculating weights, providing this value will output a CSV of linearized score and weight counts.")
	flag.BoolVar(&inputIsSolveData, "i", false, "If calculating weights, providing this switch will say the input CSV is solve data, not puzzle user difficulty.")

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

	if solve.puzzleID == 0 {
		//the production database has a zero-id'd puzzle in there for some reason.
		return false
	}

	//Cull solves that leaned too heavily on hints.
	if float64(solve.penaltyTime)/float64(solve.totalTime) > cullCheaterPercentageFlag {
		return false
	}

	self.solves = append(self.solves, solve)
	return true
}

func main() {

	/*

		There are three main phases:
			1) calculating real world difficulties for puzzles in the production database, and
			2) Calculating difficulties for solve techniques based on that data, which is based on two sub-phases:
				a) Run each puzzle through the human solver _NUMBER_OF_HUMAN_SOLVES times, and output how often we saw each step in the SolveDirections
				b) Take 2a), use it to configure a Multiple Linear Regression, and then return the coefficients of that.


		By default, we do #1 and not #2, outputing the difficulties as a CSV. If you pass -w, we'll do both phases (and skip outputting the intermediate CSV)

		By default, if we do phase #2 we take input from the first phase and feed it into the second phase. However, you can provide a CSV of phase 1 data instead as arg[0].

		If you pass -w -t then we'll do phase #2 and output the results of 2a and then stop.

		If you pass -w -i then we'll expect the provided CSV to be data from 2a and only run 2b on it.

		//TODO: rationalize the combinations of arguments you can provide.

	*/

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

	var puzzles []*puzzle
	var solveData [][]float64

	if flag.Arg(0) == "" {
		//Default: calculate relativeDifficulty like normal.

		if calcWeights {
			log.Println("No input CSV provided; calculating relative weights first.")
		}

		puzzles = calculateRelativeDifficulty()

	} else {
		if inputIsSolveData {
			log.Println("Loading solve data from CSV: ", flag.Arg(0))
		} else {
			log.Println("Loading relative difficulties data from CSV: ", flag.Arg(0))
		}
		inputFile, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal("Could not open the specified input CSV.")
		}
		defer file.Close()
		csvIn := csv.NewReader(inputFile)
		records, csvErr := csvIn.ReadAll()
		if csvErr != nil {
			log.Fatal("The provided CSV could not be parsed.")
		}
		if inputIsSolveData {
			//Input is data from phase 2a
			solveData = make([][]float64, len(records))
			for i, record := range records {
				if len(record) != len(sudoku.Techniques)+1 {
					log.Fatal("We didn't find as many columns as we expected in row: ", i)
				}
				solveData[i] = make([]float64, len(record))
				for j, item := range record {
					if theFloat, err := strconv.ParseFloat(item, 64); err == nil {
						solveData[i][j] = theFloat
					} else {
						log.Fatal(j, " column not a valid float64 in row ", i)
					}
				}
			}
		} else {
			//Input is data from phase 1
			//Read puzzles from provided CSV.
			puzzles = make([]*puzzle, len(records))
			for i, record := range records {
				thePuzzle := puzzle{}
				if len(record) < 4 {
					log.Fatal("Not enough records in row: ", i)
				}

				if theInt, err := strconv.Atoi(record[0]); err == nil {
					thePuzzle.id = theInt
				} else {
					log.Fatal("First column not a valid int in row ", i)
				}

				if theInt, err := strconv.Atoi(record[1]); err == nil {
					thePuzzle.difficultyRating = theInt
				} else {
					log.Fatal("Second column not a valid int in row ", i)
				}

				if theFloat, err := strconv.ParseFloat(record[2], 64); err == nil {
					thePuzzle.userRelativeDifficulty = theFloat
				} else {
					log.Fatal("Third column not a valid float64 in row ", i)
				}

				thePuzzle.name = record[3]

				if len(record) == 5 {
					thePuzzle.puzzle = record[4]
				} else {
					//TODO: if it doesn't, fix up the data ourselves with getPuzzleDifficultyData.
					log.Fatal("The CSV must include puzzle data. Export with -p.")
				}

				puzzles[i] = &thePuzzle
			}
		}

	}

	if calcWeights {
		//Okay, apparently we want to take all of that work and use it to calculate weights.

		//Are we going to do 2a and 2b, or just 2a, or just 2b?
		//!printPuzzleTechniques && !inputIsSolveData --> 2a + 2b
		//printPuzzleTechniques && !inputIsSolveData --> 2a and export
		//!printPuzzleTechniques && inputIsSolveData --> 2b
		//printPuzzleTechniques && inputIsSolveData --> invalid

		if len(solveData) == 0 {
			solveData = solvePuzzles(puzzles)
		}

		var stringified []string

		csvOut := csv.NewWriter(os.Stdout)
		if printPuzzleTechniques {
			if inputIsSolveData {
				log.Fatalln("Passing -t, -w, and -i together is not valid.")
			}
			//2a and export
			for _, dataPoint := range solveData {
				for _, variable := range dataPoint {
					stringified = append(stringified, strconv.FormatFloat(variable, 'f', -1, 64))
				}
				csvOut.Write(stringified)
			}
		} else {
			//Phase 2b
			result := calculateWeights(solveData)
			log.Println("Regression done. Results:")
			log.Println("N =", len(result.Data))
			log.Println("Variance Observed = ", result.VarianceObserved)
			log.Println("Variance Predicted = ", result.VariancePredicted)
			log.Println("R2 = ", result.Rsquared)
			log.Println("-------------------------")
			for i := 0; i < len(result.RegCoeff); i++ {

				var name string

				if i == 0 {
					name = "Constant"
				} else {
					name = result.Names.VariableNames[i-1]
				}
				csvOut.Write([]string{name, fmt.Sprintf("%g", result.GetRegCoeff(i))})
			}
		}
		csvOut.Flush()

	} else {
		//Apparently we just wanted to print out the relative difficulties, so do that.

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

}

func calculateRelativeDifficulty() []*puzzle {
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

	//Later we'll need to grab all of the userCollections that reference a given puzzle.
	//As we traverse through the userSolveCollections now we'll build that reference up.
	//This is a map of puzzles to a set of which collections include it.
	collectionByPuzzle := make(map[int]map[*userSolvesCollection]bool)

	for _, collection := range solvesByUser {

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

	//Cull puzzles where we don't have enough user-solves to have confidence in the rankings.
	if minPuzzleCollections != 1 {
		if verbose {
			log.Println("Starting to cull puzzles with fewer than", minPuzzleCollections, "userSolveCollections...")
		}
		counter := 0
		for id, collection := range collectionByPuzzle {
			if len(collection) < minPuzzleCollections {
				//Remove it
				delete(collectionByPuzzle, id)
				counter++
			}
		}
		if verbose {
			log.Println("Removed", counter, "puzzles because they had too few solves in their collection.")
		}
	}

	//Now, create the Markov Transition Matrix, according to algorithm MC4 of http://www.wisdom.weizmann.ac.il/~naor/PAPERS/rank_www10.html
	//The relevant part of the algorithm, from that source:
	/*

		If the current state is page P, then the next state is chosen as follows: first pick a page Q uniformly from the union of all pages ranked by the search engines. If t(Q) < t(P) for a majority of the lists t that ranked both P and Q, then go to Q, else stay in P.

	*/
	//We start by creating a stacked array of float64's that we'll pass to the matrix library.

	//This will be the dimensions of the matrix.
	numPuzzles := len(collectionByPuzzle)

	log.Println("Discovered", numPuzzles, "puzzles.")

	if numPuzzles == 0 {
		log.Fatalln("We filtered away all puzzles, so there's nothing more to do.")
		os.Exit(1)
	}

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
	for i := 0; i < _MAX_MATRIX_POWER; i++ {

		//Note: technically, this is incorrect--we should multiply by the ORIGINAL value of the matrix.
		//In practice, however, doing that takes multiple orders of magnitude more time, and the result is
		//basically indistinguishable.
		markovChain = matrix.ParallelProduct(markovChain, markovChain)

		//Are the rows converged enough for us to bail?
		difference := 0.0
		for i := 0; i < numPuzzles; i++ {
			difference += math.Abs(markovChain.Get(0, i) - markovChain.Get(1, i))
		}
		if verbose {
			log.Println("Finished matrix multiplication #", i+1, ", with a difference of", difference)
		}
		if difference < _MATRIX_DIFFERENCE_CUTOFF {
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
			thePuzzle.difficultyRating = info.difficultyRating
			thePuzzle.name = info.name
			thePuzzle.puzzle = info.puzzle
		}
		puzzles[i] = thePuzzle
	}

	//Lineralize the data, where min == 0.1 and max == 0.9
	min := math.MaxFloat64
	max := 0.0

	//Linearize data and figure out min and max so we can scale it to 0.1, 0.9 in next pass
	for i := 0; i < numPuzzles; i++ {
		//First, linearlize
		//Add 1 to make sure every input is at least 1, otherwise we'll get negative numbers which gum up later parts.
		//The larger the multiplicative constant, the closer to linear it gets. WAT?
		//TODO figure out what's going on with this constant.
		difficulty := math.Log(100000000*puzzles[i].userRelativeDifficulty + 1)

		if difficulty < min {
			min = difficulty
		}
		if difficulty > max {
			max = difficulty
		}

		puzzles[i].userRelativeDifficulty = difficulty
	}

	for i := 0; i < numPuzzles; i++ {
		difficulty := puzzles[i].userRelativeDifficulty

		//First, scale it to 0 to 1.0
		difficulty -= min
		difficulty /= (max - min)

		//Now, scale it to 0.1 to 0.9
		difficulty *= (_NORMALIZED_UPPER_BOUND - _NORMALIZED_UPPER_BOUND)
		difficulty += _NORMALIZED_LOWER_BOUND

		puzzles[i].userRelativeDifficulty = difficulty
	}

	//Sort the puzzles by relative user difficulty
	//We actually don't need the wrapper, since it will modify the underlying slice.
	sort.Sort(byUserRelativeDifficulty{puzzles})
	return puzzles
}

func solvePuzzles(puzzles []*puzzle) [][]float64 {

	//TODO: update this comment to reflect what we actually do in THIS function.
	/*
		The basic approach is to solve each puzzle many times with our human solver.
		Then, we summarize how often each technique was required for each puzzle
		(by averaging all of the solve runs together). Then we set up a multiple
		linear regression where the dependent var is the LOG of the userRelativelyDifficulty
		(to linearlize it somewhat) and the dependent vars are the number of times
		each technique was observed in the solve. Then we run the regression and
		return it.

		For more information on interpreting results from multiple linear regressions,
		see: http://onlinestatbook.com/2/regression/multiple_regression.html

	*/

	var result [][]float64

	//Generate a mapping of technique name to index.
	nameToIndex := make(map[string]int)
	for i, technique := range sudoku.Techniques {
		nameToIndex[technique.Name()] = i
	}

	for j, thePuzzle := range puzzles {

		if verbose {
			log.Println("Solving puzzle #", j)
		}

		grid := sudoku.NewGrid()
		grid.Load(convertPuzzleString(thePuzzle.puzzle))

		solveDirections := make([]sudoku.SolveDirections, _NUMBER_OF_HUMAN_SOLVES)

		sawNil := 0

		//Note: it appears that the number of solves hits a max R2 around 5 or so.
		for i := 0; i < _NUMBER_OF_HUMAN_SOLVES; i++ {

			solution := grid.HumanSolution()
			if solution == nil {
				sawNil++
			}
			solveDirections[i] = solution
		}

		if sawNil > 0 {
			log.Println("Puzzle #", thePuzzle.id, " was not able to be solved on ", sawNil, " of ", _NUMBER_OF_HUMAN_SOLVES, " runthroughs. Skipping.")
			continue
		}

		solveStats := make([]float64, len(sudoku.Techniques))

		//Accumulate number of times we've seen each technique across all solves.
		for _, directions := range solveDirections {
			for _, step := range directions {
				if index, ok := nameToIndex[step.Technique.Name()]; ok {
					solveStats[index] += 1.0
				} else {
					log.Fatal("For some reason we encountered a Technique that wasn't in hte list of Techniques")
				}
			}
		}

		//Convert each technique to an average by dividing by the number of different solves
		for i, _ := range solveStats {
			solveStats[i] /= _NUMBER_OF_HUMAN_SOLVES
		}
		result = append(result, solveStats)

	}

	return result

}

func removeZeroedColumns(stats [][]float64) [][]float64 {
	nonZeroColumns := make(map[int]bool)
	//Walk through all stats and keep track of which columns DO have non-zeros.
	for _, row := range stats {
		for i, col := range row {
			if col != 0.0 {
				nonZeroColumns[i] = true
			}
		}
	}
	indexesToKeep := make([]int, len(nonZeroColumns))
	i := 0
	for key, _ := range nonZeroColumns {
		indexesToKeep[i] = key
		i++
	}
	sort.Ints(indexesToKeep)

	var result [][]float64

	result = make([][]float64, len(stats))

	for i, row := range stats {
		result[i] = make([]float64, len(indexesToKeep))
		for j, index := range indexesToKeep {
			result[i][j] = row[index]
		}
	}

	return result

}

func calculateWeights(stats [][]float64) *regression.Regression {

	//TODO: update this description to describe what piece we actually do in THIS function.
	/*
		The basic approach is to solve each puzzle many times with our human solver.
		Then, we summarize how often each technique was required for each puzzle
		(by averaging all of the solve runs together). Then we set up a multiple
		linear regression where the dependent var is the LOG of the userRelativelyDifficulty
		(to linearlize it somewhat) and the dependent vars are the number of times
		each technique was observed in the solve. Then we run the regression and
		return it.

		For more information on interpreting results from multiple linear regressions,
		see: http://onlinestatbook.com/2/regression/multiple_regression.html

	*/

	//Set up the regression; I'll be adding data points as I go through each puzzle.
	var r regression.Regression

	r.SetObservedName("Real World Difficulty")
	for i, technique := range sudoku.Techniques {
		r.SetVarName(i, technique.Name())
	}

	for _, data := range stats {
		//TODO: remove columns that are all 0.
		//Note: I considered adding each solve for each puzzle as a separate datapoint. However, the R2 i was getting were
		//consistently much lower than this method.

		r.AddDataPoint(regression.DataPoint{Observed: data[0], Variables: data[1:]})
	}
	//Actually do the regression.
	r.RunLinearRegression()

	return &r
}

func convertPuzzleString(input string) string {
	//Puzzles stored in the database have a weird format. This function converts them into one that the sudoku library understands.

	//TODO: also handle odd things like user-provided marks and other things.

	var result string

	rows := strings.Split(input, ";")
	for _, row := range rows {
		cols := strings.Split(row, ",")
		for _, col := range cols {
			if strings.Contains(col, "!") {
				result += strings.TrimSuffix(col, "!")
			} else {
				result += "."
			}
		}
		result += "\n"
	}

	//We added an extra \n in the last runthrough, remove it.
	return strings.TrimSuffix(result, "\n")
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
