/*
dokugen is a simple command line utility that exposes many of the basic functions of the
sudoku package. It's able to generate puzzles (with difficutly) and solve provided puzzles.
Run with -h to see help on how to use it.
*/
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/gosuri/uiprogress"
	"github.com/jkomoros/sudoku"
	"github.com/jkomoros/sudoku/sdkconverter"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

//TODO: let people pass in a filename to export to.

const STORED_PUZZLES_DIRECTORY = ".puzzles"

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
	OUTPUT_CSV          bool
	CONVERTER           sdkconverter.SudokuPuzzleConverter
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

func defineFlags(flagSet *flag.FlagSet, options *appOptions) {
	flagSet.BoolVar(&options.GENERATE, "g", false, "if true, will generate a puzzle.")
	flagSet.BoolVar(&options.HELP, "h", false, "If provided, will print help and exit.")
	flagSet.IntVar(&options.NUM, "n", 1, "Number of things to generate")
	flagSet.BoolVar(&options.PRINT_STATS, "p", false, "If provided, will print stats.")
	flagSet.StringVar(&options.PUZZLE_TO_SOLVE, "s", "", "If provided, will solve the puzzle at the given filename and print solution.")
	flagSet.BoolVar(&options.WALKTHROUGH, "w", false, "If provided, will print out a walkthrough to solve the provided puzzle.")
	flagSet.StringVar(&options.RAW_SYMMETRY, "y", "vertical", "Valid values: 'none', 'both', 'horizontal', 'vertical")
	flagSet.Float64Var(&options.SYMMETRY_PROPORTION, "r", 0.7, "What proportion of cells should be filled according to symmetry")
	flagSet.IntVar(&options.MIN_FILLED_CELLS, "min-filled-cells", 0, "The minimum number of cells that should be filled in the generated puzzles.")
	flagSet.Float64Var(&options.MIN_DIFFICULTY, "min", 0.0, "Minimum difficulty for generated puzzle")
	flagSet.Float64Var(&options.MAX_DIFFICULTY, "max", 1.0, "Maximum difficulty for generated puzzle")
	flagSet.BoolVar(&options.NO_CACHE, "no-cache", false, "If provided, will not vend generated puzzles from the cache of previously generated puzzles.")
	//TODO: the format should also be how we interpret loads, too.
	flagSet.StringVar(&options.PUZZLE_FORMAT, "format", "sdk", "Which format to export puzzles from. Defaults to 'sdk'")
	flagSet.BoolVar(&options.OUTPUT_CSV, "csv", false, "Output the results in CSV.")
	flagSet.StringVar(&options.RAW_DIFFICULTY, "d", "", "difficulty, one of {gentle, easy, medium, tough}")
}

func (o *appOptions) fixUp() {
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
		log.Fatal("Unknown symmetry flag: ", o.RAW_SYMMETRY)
	}

	o.RAW_DIFFICULTY = strings.ToLower(o.RAW_DIFFICULTY)
	if o.RAW_DIFFICULTY != "" {
		vals, ok := difficultyRanges[o.RAW_DIFFICULTY]
		if !ok {
			log.Fatal("Invalid difficulty option:", o.RAW_DIFFICULTY)
		}
		o.MIN_DIFFICULTY = vals.low
		o.MAX_DIFFICULTY = vals.high
		log.Println("Using difficulty max:", strconv.FormatFloat(vals.high, 'f', -1, 64), "min:", strconv.FormatFloat(vals.low, 'f', -1, 64))
	}

	o.CONVERTER = sdkconverter.Converters[o.PUZZLE_FORMAT]

	if o.CONVERTER == nil {
		log.Fatal("Invalid format option:", o.PUZZLE_FORMAT)
	}
}

func getOptions(flagSet *flag.FlagSet, flagArguments []string) *appOptions {
	options := &appOptions{}
	defineFlags(flagSet, options)
	flagSet.Parse(flagArguments)
	options.fixUp()
	return options
}

func main() {
	flagSet := flag.CommandLine
	process(getOptions(flagSet, os.Args[1:]), os.Stdout, os.Stderr)
}

func process(options *appOptions, output io.ReadWriter, errOutput io.ReadWriter) {

	if options.HELP {
		flag.PrintDefaults()
		return
	}

	var grid *sudoku.Grid

	var csvWriter *csv.Writer
	var csvRec []string

	if options.OUTPUT_CSV {
		csvWriter = csv.NewWriter(output)
	}

	var bar *uiprogress.Bar

	//TODO: do more useful / explanatory printing here.
	if options.NUM > 1 {
		uiprogress.DefaultProgress.Out = errOutput
		uiprogress.Start()
		bar = uiprogress.AddBar(options.NUM).PrependElapsed().AppendCompleted()
	}

	for i := 0; i < options.NUM; i++ {

		if options.OUTPUT_CSV {
			csvRec = nil
		}

		//TODO: allow the type of symmetry to be configured.
		if options.GENERATE {
			grid = generatePuzzle(options.MIN_DIFFICULTY, options.MAX_DIFFICULTY, options.SYMMETRY, options.SYMMETRY_PROPORTION, options.MIN_FILLED_CELLS, options.NO_CACHE)
			//TODO: factor out all of this double-printing.
			if options.OUTPUT_CSV {
				csvRec = append(csvRec, options.CONVERTER.DataString(grid))
			} else {
				fmt.Fprintln(output, options.CONVERTER.DataString(grid))
			}
		} else if options.PUZZLE_TO_SOLVE != "" {
			//TODO: detect if the load failed.
			grid = sudoku.NewGrid()

			data, err := ioutil.ReadFile(options.PUZZLE_TO_SOLVE)

			if err != nil {
				log.Fatalln("Read error for specified file:", err)
			}

			//TODO: shouldn't a load method have a way to say the string provided is invalid?
			options.CONVERTER.Load(grid, string(data))
		}

		if grid == nil {
			//No grid to do anything with.
			log.Fatalln("No grid loaded.")
		}

		//TODO: use of this option leads to a busy loop somewhere... Is it related to the generate-multiple-and-difficulty hang?

		var directions *sudoku.SolveDirections

		if options.WALKTHROUGH || options.PRINT_STATS {
			directions = grid.HumanSolution(nil)
			if len(directions.Steps) == 0 {
				//We couldn't solve it. Let's check and see if the puzzle is well formed.
				if grid.HasMultipleSolutions() {
					//TODO: figure out why guesses wouldn't be used here effectively.
					log.Println("The puzzle had multiple solutions; that means it's not well-formed")
				}
			}
		}

		if options.WALKTHROUGH {
			if options.OUTPUT_CSV {
				csvRec = append(csvRec, directions.Walkthrough())
			} else {
				fmt.Fprintln(output, directions.Walkthrough())
			}
		}
		if options.PRINT_STATS {
			if options.OUTPUT_CSV {
				csvRec = append(csvRec, strconv.FormatFloat(grid.Difficulty(), 'f', -1, 64))
				//We won't print out the directions.Stats() like we do for just printing to stdout,
				//because that's mostly noise in this format.
			} else {
				fmt.Fprintln(output, grid.Difficulty())
				//TODO: consider actually printing out the Signals stats (with a Stats method on signals)
				fmt.Fprintln(output, strings.Join(directions.Stats(), "\n"))
			}
		}
		if options.PUZZLE_TO_SOLVE != "" {
			grid.Solve()
			if options.OUTPUT_CSV {
				csvRec = append(csvRec, options.CONVERTER.DataString(grid))
			} else {
				fmt.Fprintln(output, options.CONVERTER.DataString(grid))

			}
		}

		if options.OUTPUT_CSV {
			csvWriter.Write(csvRec)
		}

		if options.PUZZLE_TO_SOLVE != "" {
			//If we're asked to solve, n could only be 1 anyway.
			return
		}
		grid.Done()
		if options.NUM > 1 {
			bar.Incr()
		}
	}
	if options.OUTPUT_CSV {
		csvWriter.Flush()
	}
}

func puzzleDirectoryParts(symmetryType sudoku.SymmetryType, symmetryPercentage float64, minFilledCells int) []string {
	return []string{
		STORED_PUZZLES_DIRECTORY,
		"SYM_TYPE_" + strconv.Itoa(int(symmetryType)),
		"SYM_PERCENTAGE_" + strconv.FormatFloat(symmetryPercentage, 'f', -1, 64),
		"MIN_FILED_CELLS_" + strconv.Itoa(minFilledCells),
	}
}

func storePuzzle(grid *sudoku.Grid, difficulty float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64, minFilledCells int) bool {
	//TODO: we should include a hashed version of our difficulty weights file so we don't cache ones with old weights.
	directoryParts := puzzleDirectoryParts(symmetryType, symmetryPercentage, minFilledCells)

	fileNamePart := strconv.FormatFloat(difficulty, 'f', -1, 64) + ".sdk"

	pathSoFar := ""

	for i, part := range directoryParts {
		if i == 0 {
			pathSoFar = part
		} else {
			pathSoFar = filepath.Join(pathSoFar, part)
		}
		if _, err := os.Stat(pathSoFar); os.IsNotExist(err) {
			//need to create it.
			os.Mkdir(pathSoFar, 0700)
		}
	}

	fileName := filepath.Join(pathSoFar, fileNamePart)

	file, err := os.Create(fileName)

	if err != nil {
		log.Println(err)
		return false
	}

	defer file.Close()

	puzzleText := grid.DataString()

	n, err := io.WriteString(file, puzzleText)

	if err != nil {
		log.Println(err)
		return false
	} else {
		if n < len(puzzleText) {
			log.Println("Didn't write full file, only wrote", n, "bytes of", len(puzzleText))
			return false
		}
	}
	return true
}

func vendPuzzle(min float64, max float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64, minFilledCells int) *sudoku.Grid {

	directory := filepath.Join(puzzleDirectoryParts(symmetryType, symmetryPercentage, minFilledCells)...)

	if files, err := ioutil.ReadDir(directory); os.IsNotExist(err) {
		//The directory doesn't exist.
		return nil
	} else {
		//OK, the directory exists, now see which puzzles are there and if any fit. If one does, vend it and delete the file.
		for _, file := range files {
			//See what this actually returns.
			filenameParts := strings.Split(file.Name(), ".")

			//Remember: there's a dot in the filename due to the float seperator.
			//TODO: shouldn't "sdk" be in a constant somewhere?
			if len(filenameParts) != 3 || filenameParts[2] != "sdk" {
				continue
			}

			difficulty, err := strconv.ParseFloat(strings.Join(filenameParts[0:2], "."), 64)
			if err != nil {
				continue
			}

			if min <= difficulty && difficulty <= max {
				//Found a puzzle!
				grid := sudoku.NewGrid()
				fullFileName := filepath.Join(directory, file.Name())
				grid.LoadFromFile(fullFileName)
				os.Remove(fullFileName)
				return grid
			}
		}
	}
	return nil
}

func generatePuzzle(min float64, max float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64, minFilledCells int, skipCache bool) *sudoku.Grid {
	var result *sudoku.Grid

	if !skipCache {
		result = vendPuzzle(min, max, symmetryType, symmetryPercentage, minFilledCells)

		if result != nil {
			log.Println("Vending a puzzle from the cache.")
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
			log.Println("Attempt", count, "at generating puzzle.")
		}

		result = sudoku.GenerateGrid(&options)

		difficulty := result.Difficulty()

		if difficulty >= min && difficulty <= max {
			return result
		}

		log.Println("Rejecting grid of difficulty", difficulty)
		if storePuzzle(result, difficulty, symmetryType, symmetryPercentage, minFilledCells) {
			log.Println("Stored the puzzle for future use.")
		}

		count++
	}
	return nil
}
