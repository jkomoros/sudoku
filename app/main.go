package main

import (
	"dokugen/sudoku"
	"flag"
)

//TODO: let people pass in a filename to export to.

type appOptions struct {
	GENERATE        bool
	HELP            bool
	PUZZLE_TO_SOLVE string
	NUM             int
	PRINT_STATS     bool
	WALKTHROUGH     bool
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

	flag.Parse()

	if options.HELP {
		flag.PrintDefaults()
		return
	}

	if options.GENERATE {
		for i := 0; i < options.NUM; i++ {
			grid := sudoku.GenerateGrid()
			print(grid.DataString())
			print("\n\n")
			if options.PRINT_STATS {
				print("\n\n")
				print(grid.Difficulty())
			}
		}
		return
	}

	if options.PUZZLE_TO_SOLVE != "" {
		grid := sudoku.NewGrid()
		grid.LoadFromFile(options.PUZZLE_TO_SOLVE)
		//TODO: detect if the load failed.

		//TODO: use of this option leads to a busy loop somewhere... Is it related to the generate-multiple-and-difficulty hang?
		if options.WALKTHROUGH {
			print(grid.HumanWalkthrough())
			print("\n\n")
		}
		if options.PRINT_STATS {
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
