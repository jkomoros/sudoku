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
const QUERY_LIMIT = 100

var noLimitFlag bool

func init() {
	flag.BoolVar(&noLimitFlag, "a", false, "Specify to execute the solves query with no limit.")
}

type dbConfig struct {
	Url, Username, Password, DbName, SolvesTable, SolvesID, SolvesPuzzleID, SolvesTotalTime, SolvesUser, PuzzlesTable, PuzzlesID, PuzzlesDifficulty string
}

type solve struct {
	puzzleID  int
	totalTime int
}

type userSolvesCollection struct {
	solves []solve
	max    int
	min    int
}

type puzzle struct {
	id                     int
	userRelativeDifficulty float32
	difficultyRating       int
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

func (self *userSolvesCollection) addSolve(solve solve) {
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
}

//Whehter or not this should be included in calculation.
//Basically, whether the reltaiveDifficulties will all be valid.
//Normally this returns false if there is only one solve by the user, but could also
//happen when there are multiple solves but (crazily enough) they all have exactly the same solveTime.
//This DOES happen in the production dataset.
func (self *userSolvesCollection) valid() bool {
	return self.max != self.min
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
	var config dbConfig
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("There was an error parsing JSON from the config file: ", err)
		os.Exit(1)
	}

	difficutlyRatingsChan := make(chan map[int]int)

	go getPuzzleDifficultyRatings(&config, difficutlyRatingsChan)

	db := mysql.New("tcp", "", config.Url, config.Username, config.Password, config.DbName)

	if err := db.Connect(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var solvesQuery string

	if noLimitFlag {
		fmt.Println("Running without a limit for number of solves to retrieve.")
		solvesQuery = "select %s, %s, %s from %s"
	} else {
		fmt.Println("Running with a limit of ", QUERY_LIMIT, " for number of solves to retrieve.")
		solvesQuery = "select %s, %s, %s from %s limit " + strconv.Itoa(QUERY_LIMIT)
	}

	res, err := db.Start(solvesQuery, config.SolvesUser, config.SolvesPuzzleID, config.SolvesTotalTime, config.SolvesTable)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	solvesByUser := make(map[string]*userSolvesCollection)

	var userSolves *userSolvesCollection
	var ok bool
	var i int

	//First, process all user records in the DB to collect all solves by userName.
	for {

		row, _ := res.GetRow()

		if row == nil {
			break
		}

		userSolves, ok = solvesByUser[row.Str(0)]

		if !ok {
			userSolves = new(userSolvesCollection)
			solvesByUser[row.Str(0)] = userSolves
		}

		userSolves.addSolve(solve{row.Int(1), row.Int(2)})
		i++
	}

	fmt.Println("Processed ", i, " solves by ", len(solvesByUser), " users.")

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

	}

	fmt.Println("Skipped ", skippedUsers, " users because they had only had a single unique solve time.")

	puzzles := make([]puzzle, len(relativeDifficultiesByPuzzle))

	var index int

	for puzzleID, difficulties := range relativeDifficultiesByPuzzle {
		var sum float32
		for _, difficulty := range difficulties {
			sum += difficulty
		}
		puzzles[index] = puzzle{puzzleID, sum / float32(len(difficulties)), -1}
		index++
	}

	//Sort the puzzles by relative user difficulty
	//We actually don't need the wrapper, since it will modify the underlying slice.
	sort.Sort(byUserRelativeDifficulty{puzzles})

	//Merge in the difficulty ratings from the server.
	difficultyRatings := <-difficutlyRatingsChan

	for i, puzzle := range puzzles {
		difficulty, ok := difficultyRatings[puzzle.id]
		if ok {
			puzzle.difficultyRating = difficulty
		}
		//It's not a pointer so we have to copy it back.
		puzzles[i] = puzzle
	}

	//Now print the results to stdout.

	csvOut := csv.NewWriter(os.Stdout)

	for _, puzzle := range puzzles {
		csvOut.Write([]string{strconv.Itoa(puzzle.id), strconv.Itoa(puzzle.difficultyRating), fmt.Sprintf("%g", puzzle.userRelativeDifficulty)})
	}

	csvOut.Flush()

}

func getPuzzleDifficultyRatings(config *dbConfig, result chan map[int]int) {

	db := mysql.New("tcp", "", config.Url, config.Username, config.Password, config.DbName)

	if err := db.Connect(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	res, err := db.Start("select %s, %s from %s", config.PuzzlesID, config.PuzzlesDifficulty, config.PuzzlesTable)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	puzzles := make(map[int]int)

	for {

		row, _ := res.GetRow()

		if row == nil {
			break
		}

		puzzles[row.Int(0)] = row.Int(1)
	}

	result <- puzzles

}
