package main

import (
	"dokugen"
	"flag"
	"fmt"
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
	SYMMETRY            sudoku.SymmetryType
	SYMMETRY_PROPORTION float64
	MIN_DIFFICULTY      float64
	MAX_DIFFICULTY      float64
}

func init() {
	//grid.Difficulty can make use of a number of processes simultaneously.
	runtime.GOMAXPROCS(6)
}

func main() {

	//TODO: figure out how to test this.

	var options appOptions

	flag.BoolVar(&options.GENERATE, "g", false, "if true, will generate a puzzle.")
	flag.BoolVar(&options.HELP, "h", false, "If provided, will print help and exit.")
	flag.IntVar(&options.NUM, "n", 1, "Number of things to generate")
	flag.BoolVar(&options.PRINT_STATS, "p", false, "If provided, will print stats.")
	flag.StringVar(&options.PUZZLE_TO_SOLVE, "s", "", "If provided, will solve the puzzle at the given filename and print solution.")
	flag.BoolVar(&options.WALKTHROUGH, "w", false, "If provided, will print out a walkthrough to solve the provided puzzle.")
	flag.StringVar(&options.RAW_SYMMETRY, "y", "vertical", "Valid values: 'none', 'both', 'horizontal', 'vertical")
	flag.Float64Var(&options.SYMMETRY_PROPORTION, "r", 0.7, "What proportion of cells should be filled according to symmetry")
	flag.Float64Var(&options.MIN_DIFFICULTY, "min", 0.0, "Minimum difficulty for generated puzzle")
	flag.Float64Var(&options.MAX_DIFFICULTY, "max", 1.0, "Maximum difficulty for generated puzzle")

	flag.Parse()

	options.RAW_SYMMETRY = strings.ToLower(options.RAW_SYMMETRY)
	switch options.RAW_SYMMETRY {
	case "none":
		options.SYMMETRY = sudoku.SYMMETRY_NONE
	case "both":
		options.SYMMETRY = sudoku.SYMMETRY_BOTH
	case "horizontal":
		options.SYMMETRY = sudoku.SYMMETRY_HORIZONTAL
	case "vertical":
		options.SYMMETRY = sudoku.SYMMETRY_VERTICAL
	default:
		log.Fatal("Unknown symmetry flag: ", options.RAW_SYMMETRY)
	}

	output := os.Stdout

	if options.HELP {
		flag.PrintDefaults()
		return
	}

	var grid *sudoku.Grid

	for i := 0; i < options.NUM; i++ {
		//TODO: allow the type of symmetry to be configured.
		if options.GENERATE {
			grid = generatePuzzle(options.MIN_DIFFICULTY, options.MAX_DIFFICULTY, options.SYMMETRY, options.SYMMETRY_PROPORTION)
			fmt.Fprintln(output, grid.DataString())
		} else if options.PUZZLE_TO_SOLVE != "" {
			//TODO: detect if the load failed.
			grid = sudoku.NewGrid()
			grid.LoadFromFile(options.PUZZLE_TO_SOLVE)
		}

		if grid == nil {
			//No grid to do anything with.
			log.Fatalln("No grid loaded.")
		}

		//TODO: use of this option leads to a busy loop somewhere... Is it related to the generate-multiple-and-difficulty hang?

		var directions sudoku.SolveDirections

		if options.WALKTHROUGH || options.PRINT_STATS {
			directions = grid.HumanSolution()
			if len(directions) == 0 {
				//We couldn't solve it. Let's check and see if the puzzle is well formed.
				if grid.HasMultipleSolutions() {
					//TODO: figure out why guesses wouldn't be used here effectively.
					log.Println("The puzzle had multiple solutions; that means it's not well-formed")
				}
			}
		}

		if options.WALKTHROUGH {
			fmt.Fprintln(output, directions.Walkthrough(grid))
		}
		if options.PRINT_STATS {
			fmt.Fprintln(output, grid.Difficulty())
			//TODO: consider actually printing out the Signals stats (with a Stats method on signals)
			fmt.Fprintln(output, strings.Join(directions.Stats(), "\n"))
		}
		if options.PUZZLE_TO_SOLVE != "" {
			grid.Solve()
			fmt.Fprintln(output, grid.DataString())
			//If we're asked to solve, n could only be 1 anyway.
			return
		}
		grid.Done()
	}

}

func puzzleDirectoryParts(symmetryType sudoku.SymmetryType, symmetryPercentage float64) []string {
	return []string{
		STORED_PUZZLES_DIRECTORY,
		"SYM_TYPE_" + strconv.Itoa(int(symmetryType)),
		"SYM_PERCENTAGE_" + strconv.FormatFloat(symmetryPercentage, 'f', -1, 64),
	}
}

func storePuzzle(grid *sudoku.Grid, difficulty float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64) bool {
	//TODO: we should include a hashed version of our difficulty weights file so we don't cache ones with old weights.
	directoryParts := puzzleDirectoryParts(symmetryType, symmetryPercentage)

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

func vendPuzzle(min float64, max float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64) *sudoku.Grid {
	directory := filepath.Join(puzzleDirectoryParts(symmetryType, symmetryPercentage)...)

	if files, err := ioutil.ReadDir(directory); os.IsNotExist(err) {
		//The directory doesn't exist.
		return nil
	} else {
		//OK, the directory exists, now see which puzzles are there and if any fit. If one does, vend it and delete the file.
		for _, file := range files {
			//See what this actually returns.
			log.Println(file.Name())
		}
	}
	return nil
}

func generatePuzzle(min float64, max float64, symmetryType sudoku.SymmetryType, symmetryPercentage float64) *sudoku.Grid {
	var result *sudoku.Grid

	result = vendPuzzle(min, max, symmetryType, symmetryPercentage)

	if result != nil {
		log.Println("Vending a puzzle from the cache.")
		return result
	}

	//We'll have to generate one ourselves.
	count := 0
	for {
		log.Println("Attempt", count, "at generating puzzle.")

		result = sudoku.GenerateGrid(symmetryType, symmetryPercentage)

		difficulty := result.Difficulty()

		if difficulty >= min && difficulty <= max {
			return result
		}

		log.Println("Rejecting grid of difficulty", difficulty)
		if storePuzzle(result, difficulty, symmetryType, symmetryPercentage) {
			log.Println("Stored the puzzle for future use.")
		}

		count++
	}
	return nil
}
