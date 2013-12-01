package main

import (
	"dokugen/sudoku"
	"flag"
)

//TODO: let people pass in a filename to export to.

var GENERATE bool
var HELP bool
var PUZZLE_TO_SOLVE string
var NUM int

func main() {

	flag.BoolVar(&GENERATE, "g", false, "if true, will generate a puzzle.")
	flag.BoolVar(&HELP, "h", false, "If provided, will print help and exit.")
	flag.IntVar(&NUM, "n", 1, "Number of things to generate")
	flag.StringVar(&PUZZLE_TO_SOLVE, "s", "", "If provided, will solve the puzzle at the given filename and print solution.")

	flag.Parse()

	if HELP {
		flag.PrintDefaults()
		return
	}

	if GENERATE {
		for i := 0; i < NUM; i++ {
			grid := sudoku.GenerateGrid()
			print(grid.DataString())
			print("\n\n")
		}
		return
	}

	if PUZZLE_TO_SOLVE != "" {
		grid := sudoku.NewGrid()
		grid.LoadFromFile(PUZZLE_TO_SOLVE)
		//TODO: detect if the load failed.
		grid.Solve()
		print(grid.DataString())
		return
	}

	//If we get to here, print defaults.
	flag.PrintDefaults()

}
