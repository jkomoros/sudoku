/*
dokugen is a simple command line utility that exposes many of the basic functions of the
sudoku package. It's able to generate puzzles (with difficutly) and solve provided puzzles.
Run with -h to see help on how to use it.
*/
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gosuri/uiprogress"
	"github.com/jkomoros/sudoku"
	"github.com/jkomoros/sudoku/sdkconverter"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
)

//TODO: let people pass in a filename to export to.

const _STORED_PUZZLES_DB = ".puzzle_cache"

//Used as the grid to pass back when FAKE-GENERATE is true.
const TEST_GRID = `6|1|2|.|.|.|4|.|3
.|3|.|4|9|.|.|7|2
.|.|7|.|.|.|.|6|5
.|.|.|.|6|1|.|8|.
1|.|3|.|4|.|2|.|6
.|6|.|5|2|.|.|.|.
.|9|.|.|.|.|5|.|.
7|2|.|.|8|5|.|3|.
5|.|1|.|.|.|9|4|7`

type appOptions struct {
	GENERATE            bool
	HELP                bool
	PUZZLE_TO_SOLVE     string
	NUM                 int
	PRINT_STATS         bool
	WALKTHROUGH         bool
	RAW_SYMMETRY        string
	RAW_DIFFICULTY      string
	SYMMETRY            sudoku.SymmetryType
	SYMMETRY_PROPORTION float64
	MIN_FILLED_CELLS    int
	MIN_DIFFICULTY      float64
	MAX_DIFFICULTY      float64
	NO_CACHE            bool
	PUZZLE_FORMAT       string
	NO_PROGRESS         bool
	CSV                 bool
	CONVERTER           sdkconverter.SudokuPuzzleConverter
	//Only used in testing.
	FAKE_GENERATE bool
	flagSet       *flag.FlagSet
	progress      *uiprogress.Progress
}

type outputWriter struct {
	output    io.ReadWriter
	csvWriter *csv.Writer
	tempRec   []string
}

var difficultyRanges map[string]struct {
	low, high float64
}

func init() {
	//grid.Difficulty can make use of a number of processes simultaneously.
	runtime.GOMAXPROCS(6)

	difficultyRanges = map[string]struct{ low, high float64 }{
		"gentle": {0.0, 0.3},
		"easy":   {0.3, 0.6},
		"medium": {0.6, 0.7},
		"tough":  {0.7, 1.0},
	}
}

func (o *outputWriter) Write(output string, extra string) {
	//Extra will only be written if not going to CSV
	if o.csvWriter == nil {
		fmt.Fprintln(o.output, output)
		if extra != "" {
			fmt.Fprintln(o.output, extra)
		}
		return
	}
	o.tempRec = append(o.tempRec, output)
}

func (o *outputWriter) EndOfRec() {
	if o.csvWriter == nil {
		return
	}
	o.csvWriter.Write(o.tempRec)
	o.tempRec = nil
}

func (o *outputWriter) Done() {
	if o.csvWriter == nil {
		return
	}
	o.csvWriter.Flush()
}

func NewOutputWriter(options *appOptions, output io.ReadWriter) *outputWriter {
	var csvWriter *csv.Writer
	if options.CSV {
		csvWriter = csv.NewWriter(output)
	}
	return &outputWriter{output, csvWriter, nil}
}

func defineFlags(options *appOptions) {
	options.flagSet.BoolVar(&options.GENERATE, "g", false, "if true, will generate a puzzle.")
	options.flagSet.BoolVar(&options.HELP, "h", false, "If provided, will print help and exit.")
	options.flagSet.IntVar(&options.NUM, "n", 1, "Number of things to generate")
	options.flagSet.BoolVar(&options.PRINT_STATS, "p", false, "If provided, will print stats.")
	options.flagSet.StringVar(&options.PUZZLE_TO_SOLVE, "s", "", "If provided, will solve the puzzle at the given filename and print solution. If -csv is provided, will expect the file to be a csv where the first column of each row is a puzzle in the specified puzzle format.")
	options.flagSet.BoolVar(&options.WALKTHROUGH, "w", false, "If provided, will print out a walkthrough to solve the provided puzzle.")
	options.flagSet.StringVar(&options.RAW_SYMMETRY, "y", "vertical", "Valid values: 'none', 'both', 'horizontal', 'vertical")
	options.flagSet.Float64Var(&options.SYMMETRY_PROPORTION, "r", 0.7, "What proportion of cells should be filled according to symmetry")
	options.flagSet.IntVar(&options.MIN_FILLED_CELLS, "min-filled-cells", 0, "The minimum number of cells that should be filled in the generated puzzles.")
	options.flagSet.Float64Var(&options.MIN_DIFFICULTY, "min", 0.0, "Minimum difficulty for generated puzzle")
	options.flagSet.Float64Var(&options.MAX_DIFFICULTY, "max", 1.0, "Maximum difficulty for generated puzzle")
	options.flagSet.BoolVar(&options.NO_CACHE, "no-cache", false, "If provided, will not vend generated puzzles from the cache of previously generated puzzles.")
	//TODO: the format should also be how we interpret loads, too.
	options.flagSet.StringVar(&options.PUZZLE_FORMAT, "format", "sdk", "Which format to export puzzles from. Defaults to 'sdk'")
	options.flagSet.BoolVar(&options.CSV, "csv", false, "Export CSV, and expect inbound puzzle files to be a CSV with a puzzle per row.")
	options.flagSet.StringVar(&options.RAW_DIFFICULTY, "d", "", "difficulty, one of {gentle, easy, medium, tough}")
	options.flagSet.BoolVar(&options.NO_PROGRESS, "no-progress", false, "If provided, will not print a progress bar")
}

//If it returns true, the program should quit.
func (o *appOptions) fixUp(errOutput io.ReadWriter) bool {

	if errOutput == nil {
		errOutput = os.Stderr
	}

	logger := log.New(errOutput, "", log.LstdFlags)

	o.RAW_SYMMETRY = strings.ToLower(o.RAW_SYMMETRY)
	switch o.RAW_SYMMETRY {
	case "none":
		o.SYMMETRY = sudoku.SYMMETRY_NONE
	case "both":
		o.SYMMETRY = sudoku.SYMMETRY_BOTH
	case "horizontal":
		o.SYMMETRY = sudoku.SYMMETRY_HORIZONTAL
	case "vertical":
		o.SYMMETRY = sudoku.SYMMETRY_VERTICAL
	default:
		logger.Println("Unknown symmetry flag: ", o.RAW_SYMMETRY)
		return true
	}

	o.RAW_DIFFICULTY = strings.ToLower(o.RAW_DIFFICULTY)
	if o.RAW_DIFFICULTY != "" {
		vals, ok := difficultyRanges[o.RAW_DIFFICULTY]
		if !ok {
			logger.Println("Invalid difficulty option:", o.RAW_DIFFICULTY)
			return true
		}
		o.MIN_DIFFICULTY = vals.low
		o.MAX_DIFFICULTY = vals.high
		logger.Println("Using difficulty max:", strconv.FormatFloat(vals.high, 'f', -1, 64), "min:", strconv.FormatFloat(vals.low, 'f', -1, 64))
	}

	o.CONVERTER = sdkconverter.Converters[o.PUZZLE_FORMAT]

	if o.CONVERTER == nil {
		logger.Println("Invalid format option:", o.PUZZLE_FORMAT)
		return true
	}
	return false
}

func getOptions(flagSet *flag.FlagSet, flagArguments []string, errOutput io.ReadWriter) *appOptions {
	options := &appOptions{flagSet: flagSet}
	defineFlags(options)
	flagSet.Parse(flagArguments)
	if options.fixUp(errOutput) {
		os.Exit(1)
	}
	return options
}

func main() {
	flagSet := flag.CommandLine
	process(getOptions(flagSet, os.Args[1:], nil), os.Stdout, os.Stderr)
}

func process(options *appOptions, output io.ReadWriter, errOutput io.ReadWriter) {

	options.flagSet.SetOutput(errOutput)

	if options.HELP {
		options.flagSet.PrintDefaults()
		return
	}

	logger := log.New(errOutput, "", log.LstdFlags)

	var grid *sudoku.Grid

	writer := NewOutputWriter(options, output)

	var bar *uiprogress.Bar

	//TODO: do more useful / explanatory printing here.
	if options.NUM > 1 && !options.NO_PROGRESS {
		options.progress = uiprogress.New()
		options.progress.Out = errOutput
		options.progress.Start()
		bar = options.progress.AddBar(options.NUM).PrependElapsed().AppendCompleted()
	}

	var incomingPuzzles []*sudoku.Grid

	if options.PUZZLE_TO_SOLVE != "" {
		//There are puzzles to load up.

		data, err := ioutil.ReadFile(options.PUZZLE_TO_SOLVE)

		if err != nil {
			logger.Fatalln("Read error for specified file:", err)
		}

		var tempGrid *sudoku.Grid

		var puzzleData []string

		if options.CSV {
			//Load up multiple.
			csvReader := csv.NewReader(bytes.NewReader(data))
			rows, err := csvReader.ReadAll()
			if err != nil {
				logger.Fatalln("The provided input CSV was not a valid CSV:", err)
			}
			for _, row := range rows {
				puzzleData = append(puzzleData, row[0])
			}
		} else {
			//Just load up a single file worth.
			puzzleData = []string{string(data)}
		}

		for _, puzz := range puzzleData {

			tempGrid = sudoku.NewGrid()

			//TODO: shouldn't a load method have a way to say the string provided is invalid?
			options.CONVERTER.Load(tempGrid, string(puzz))

			incomingPuzzles = append(incomingPuzzles, tempGrid)
		}
		//Tell the main loop how many puzzles to expect.
		//TODO: this feels a bit like a hack, doesn't it? options.NUM is normally a user input value.
		options.NUM = len(incomingPuzzles)

	}

	for i := 0; i < options.NUM; i++ {

		//TODO: allow the type of symmetry to be configured.
		if options.GENERATE {
			if options.FAKE_GENERATE {
				grid = sudoku.NewGrid()
				grid.LoadSDK(TEST_GRID)
			} else {
				grid = generatePuzzle(options.MIN_DIFFICULTY, options.MAX_DIFFICULTY, options.SYMMETRY, options.SYMMETRY_PROPORTION, options.MIN_FILLED_CELLS, options.NO_CACHE, logger)
			}
			writer.Write(options.CONVERTER.DataString(grid), "")
		} else if len(incomingPuzzles)-1 >= i {
			//Load up an inbound puzzle
			grid = incomingPuzzles[i]
		}

		if grid == nil {
			//No grid to do anything with.
			logger.Fatalln("No grid loaded.")
		}

		//TODO: use of this option leads to a busy loop somewhere... Is it related to the generate-multiple-and-difficulty hang?

		var directions *sudoku.SolveDirections

		if options.WALKTHROUGH || options.PRINT_STATS {
			directions = grid.HumanSolution(nil)
			if len(directions.Steps) == 0 {
				//We couldn't solve it. Let's check and see if the puzzle is well formed.
				if grid.HasMultipleSolutions() {
					//TODO: figure out why guesses wouldn't be used here effectively.
					logger.Println("The puzzle had multiple solutions; that means it's not well-formed")
				}
			}
		}

		if options.WALKTHROUGH {
			writer.Write(directions.Walkthrough(), "")
		}
		if options.PRINT_STATS {
			writer.Write(strconv.FormatFloat(grid.Difficulty(), 'f', -1, 64),
				strings.Join(directions.Stats(), "\n"))
		}
		//TODO: using the existence of options.PUZZLE_TO_SOLVE as the way to detect that
		//we are working on an inbound puzzle seems a bit hackish.
		if options.PUZZLE_TO_SOLVE != "" {
			grid.Solve()
			writer.Write(options.CONVERTER.DataString(grid), "")
		}

		writer.EndOfRec()

		grid.Done()
		if bar != nil {
			bar.Incr()
		}
	}
	writer.Done()
}

type StoredPuzzle struct {
	Options    *sudoku.GenerationOptions
	Difficulty float64
	//In DOKU format
	PuzzleData string
}

//TODO: take a sudoku.GenerationOptions to simplify signature
func storePuzzle(dbName string, grid *sudoku.Grid, difficulty float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64, minFilledCells int, logger *log.Logger) bool {

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		logger.Fatalln("Couldn't open DB file", err)
		return false
	}
	defer db.Close()

	converter := sdkconverter.Converters["doku"]

	if converter == nil {
		logger.Fatalln("Couldn't find doku converter")
	}

	puzzleData := converter.DataString(grid)

	if puzzleData == "" {
		logger.Fatalln("Puzzle didn't convert to doku format cleanly")
	}

	puzzleObj := &StoredPuzzle{
		Options: &sudoku.GenerationOptions{
			Symmetry:           symmetryType,
			SymmetryPercentage: symmetryPercentage,
			MinFilledCells:     minFilledCells,
		},
		Difficulty: difficulty,
		PuzzleData: puzzleData,
	}

	jsonBlob, err := json.MarshalIndent(puzzleObj, "", "    ")
	if err != nil {
		logger.Fatalln("Json couldn't be marshalled", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(sudoku.DIFFICULTY_MODEL))
		if err != nil {
			return err
		}

		id, err := bucket.NextSequence()

		if err != nil {
			return err
		}

		err = bucket.Put([]byte(strconv.Itoa(int(id))), []byte(jsonBlob))
		if err != nil {
			return err
		}

		//It worked
		return nil

	})

	if err != nil {
		logger.Fatalln("Transacation failed: ", err)
		return false
	}

	return true
}

//TODO: take a sudoku.GenerationOptions to simplify signature
func vendPuzzle(dbName string, min float64, max float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64, minFilledCells int) *sudoku.Grid {

	//TODO: test storePuzzle and vendPuzzle

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		//TODO: pass in logger
		log.Fatalln("Couldn't open DB file", err)
		return nil
	}
	defer db.Close()

	converter := sdkconverter.Converters["doku"]

	if converter == nil {
		log.Fatalln("Couldn't find doku converter")
	}

	var finalPuzzle string

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(sudoku.DIFFICULTY_MODEL))
		if err != nil {
			log.Println(err)
			return err
		}

		matchingPuzzles := make(map[string]string)

		bucket.ForEach(func(key, value []byte) error {
			puzzleInfo := &StoredPuzzle{}

			err := json.Unmarshal(value, puzzleInfo)
			if err != nil {
				return err
			}

			if puzzleInfo.Options.MinFilledCells != minFilledCells {
				//Doesn't match
				return nil
			}

			if puzzleInfo.Options.SymmetryPercentage != symmetryPercentage {
				//Doesn't match
				return nil
			}

			if puzzleInfo.Options.Symmetry != symmetryType {
				//Doesn't match
				return nil
			}

			if puzzleInfo.Difficulty > max || puzzleInfo.Difficulty < min {
				//Doesn't match
				return nil
			}

			//Does match!
			matchingPuzzles[string(key)] = puzzleInfo.PuzzleData
			return nil

		})

		//Select one at random

		if len(matchingPuzzles) == 0 {
			//No puzzles matched.
			return nil
		}

		var keys []string

		for key, _ := range matchingPuzzles {
			keys = append(keys, key)
		}

		key := keys[rand.Intn(len(keys))]

		finalPuzzle = matchingPuzzles[key]

		//Doesn't matter that much if we can't delete the key.
		err = bucket.Delete([]byte(key))

		if err != nil {
			//TODO: shouldn't we ahve a logger here?
			log.Println("Couldn't delete the key we picked:", err)
		}

		return nil

	})

	if finalPuzzle == "" {
		return nil
	}

	grid := sudoku.NewGrid()

	converter.Load(grid, finalPuzzle)

	return grid
}

func generatePuzzle(min float64, max float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64, minFilledCells int, skipCache bool, logger *log.Logger) *sudoku.Grid {
	var result *sudoku.Grid

	if !skipCache {
		result = vendPuzzle(_STORED_PUZZLES_DB, min, max, symmetryType, symmetryPercentage, minFilledCells)

		if result != nil {
			logger.Println("Vending a puzzle from the cache.")
			return result
		}
	}

	options := sudoku.GenerationOptions{
		Symmetry:           symmetryType,
		SymmetryPercentage: symmetryPercentage,
		MinFilledCells:     minFilledCells,
	}

	//We'll have to generate one ourselves.
	count := 0
	for {
		//The first time we don't bother saying what number attemp it is, because if the first run is likely to generate a useable puzzle it's just noise.
		if count != 0 {
			logger.Println("Attempt", count, "at generating puzzle.")
		}

		result = sudoku.GenerateGrid(&options)

		difficulty := result.Difficulty()

		if difficulty >= min && difficulty <= max {
			return result
		}

		logger.Println("Rejecting grid of difficulty", difficulty)
		if storePuzzle(_STORED_PUZZLES_DB, result, difficulty, symmetryType, symmetryPercentage, minFilledCells, logger) {
			logger.Println("Stored the puzzle for future use.")
		}

		count++
	}
	return nil
}
