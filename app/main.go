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
var PRINT_STATS bool
var WALKTHROUGH bool

func main() {

	//TODO: figure out how to test this.

	flag.BoolVar(&GENERATE, "g", false, "if true, will generate a puzzle.")
	flag.BoolVar(&HELP, "h", false, "If provided, will print help and exit.")
	flag.IntVar(&NUM, "n", 1, "Number of things to generate")
	flag.BoolVar(&PRINT_STATS, "p", false, "If provided, will print stats.")
	flag.StringVar(&PUZZLE_TO_SOLVE, "s", "", "If provided, will solve the puzzle at the given filename and print solution.")
	flag.BoolVar(&WALKTHROUGH, "w", false, "If provided, will print out a walkthrough to solve the provided puzzle.")

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
			if PRINT_STATS {
				print("\n\n")
				print(grid.Difficulty())
			}
		}
		return
	}

	if PUZZLE_TO_SOLVE != "" {
		grid := sudoku.NewGrid()
		grid.LoadFromFile(PUZZLE_TO_SOLVE)
		//TODO: detect if the load failed.

		//TODO: use of this option leads to a busy loop somewhere... Is it related to the generate-multiple-and-difficulty hang?
		if WALKTHROUGH {
			print(grid.HumanWalkthrough())
			print("\n\n")
		}
		if PRINT_STATS {
			print("\n\n")
			print(grid.Difficulty())
		}
		grid.Solve()
		print(grid.DataString())

		return
	}

	//If we get to here, print defaults.
	flag.PrintDefaults()

}
