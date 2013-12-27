package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"log"
	"os"
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

func init() {
	flag.BoolVar(&noLimitFlag, "a", false, "Specify to execute the solves query with no limit.")
	flag.BoolVar(&printPuzzleDataFlag, "p", false, "Specify that you want puzzle data printed out in the output.")
	flag.Float64Var(&cullCheaterPercentageFlag, "c", _PENALTY_PERCENTAGE_CUTOFF, "What percentage of solve time must be penalty for someone to be considered a cheater.")
	flag.IntVar(&minimumSolvesFlag, "s", _MINIMUM_SOLVES, "How many solves a user must have their scores considered.")
	flag.IntVar(&queryLimit, "n", _QUERY_LIMIT, "Number of solves to fetch from the database.")
	flag.BoolVar(&useMockData, "m", false, "Use mock data (useful if you don't have a real database to test with).")
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
	userRelativeDifficulty float32
	difficultyRating       int
	name                   string
	puzzle                 string
}

type puzzles []puzzle

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

	go getPuzzleDifficultyRatings(difficutlyRatingsChan)

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

	seenRows := make(map[string]bool)

	//First, process all user records in the DB to collect all solves by userName.
	for {

		row, _ := res.GetRow()

		if row == nil {
			break
		}

		i++

		rowHashValue := fmt.Sprintf("%v", row)

		if _, seen := seenRows[rowHashValue]; seen {
			skippedDuplicateSolves++
			continue
		} else {
			seenRows[rowHashValue] = true
		}

		userSolves, ok = solvesByUser[row.Str(0)]

		if !ok {
			userSolves = new(userSolvesCollection)
			userSolves.idPosition = make(map[int]int)
			solvesByUser[row.Str(0)] = userSolves
		}

		if !userSolves.addSolve(solve{row.Int(1), row.Int(2), row.Int(3)}) {
			skippedSolves++
		}

	}

	log.Println("Processed ", i, " solves by ", len(solvesByUser), " users.")
	log.Println("Skipped ", skippedSolves, " solves that cheated too much.")
	log.Println("Skipped ", skippedDuplicateSolves, " solves because they were duplicates of solves seen earlier.")

	//Now get the relative difficulty for each user's puzzles, and collect them.

	relativeDifficultiesByPuzzle := make(map[int][]float32)

	var skippedUsers int

	for _, collection := range solvesByUser {

		if !collection.valid() {
			skippedUsers++
			continue
		}

		for puzzleID, relativeDifficulty := range collection.relativeDifficulties() {
			relativeDifficultiesByPuzzle[puzzleID] = append(relativeDifficultiesByPuzzle[puzzleID], relativeDifficulty)
		}

		sort.Sort(bySolveTimeDsc(collection.solves))

		for i, puzzle := range collection.solves {
			collection.idPosition[puzzle.puzzleID] = i
		}

	}

	log.Println("Skipped ", skippedUsers, " users because they did not have enough solve times.")

	puzzles := make([]puzzle, len(relativeDifficultiesByPuzzle))

	var index int

	for puzzleID, difficulties := range relativeDifficultiesByPuzzle {
		var sum float32
		for _, difficulty := range difficulties {
			sum += difficulty
		}
		puzzles[index] = puzzle{id: puzzleID, userRelativeDifficulty: sum / float32(len(difficulties)), difficultyRating: -1}
		index++
	}

	//Sort the puzzles by relative user difficulty
	//We actually don't need the wrapper, since it will modify the underlying slice.
	sort.Sort(byUserRelativeDifficulty{puzzles})

	//Merge in the difficulty ratings from the server.
	difficultyRatings := <-difficutlyRatingsChan

	for i, puzzle := range puzzles {
		info, ok := difficultyRatings[puzzle.id]
		if ok {
			puzzle.difficultyRating = info.difficultyRating
			puzzle.name = info.name
			puzzle.puzzle = info.puzzle
		}
		//It's not a pointer so we have to copy it back.
		puzzles[i] = puzzle
	}

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
