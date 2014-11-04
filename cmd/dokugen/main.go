package main

import (
	"dokugen"
	"flag"
	"fmt"
	"os"
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

	output := os.Stdout

	if options.HELP {
		flag.PrintDefaults()
		return
	}

	if options.GENERATE {
		for i := 0; i < options.NUM; i++ {
			//TODO: allow the type of symmetry to be configured.
			grid := sudoku.GenerateGrid(sudoku.SYMMETRY_NONE)
			fmt.Fprintln(output, grid.DataString())
			fmt.Fprintln(output, "\n")
			if options.PRINT_STATS {
				fmt.Fprintln(output, "\n")
				fmt.Fprintln(output, grid.Difficulty())
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
			fmt.Fprintln(output, grid.HumanWalkthrough())
			fmt.Fprintln(output, "\n")
		}
		if options.PRINT_STATS {
			fmt.Fprintln(output, "\n")
			fmt.Fprintln(output, grid.Difficulty())
		}
		grid.Solve()
		fmt.Fprintln(output, grid.DataString())

		return
	}

	//If we get to here, print defaults.
	flag.PrintDefaults()

}
