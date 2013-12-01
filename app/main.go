package main

import (
	"dokugen/sudoku"
	"flag"
)

//TODO: let people pass in puzzles to solve.
//TODO: let people pass in a filename to export to.

var GENERATE bool

func main() {

	flag.BoolVar(&GENERATE, "g", false, "if true, will generate a puzzle.")

	flag.Parse()

	if GENERATE {
		grid := sudoku.GenerateGrid()
		print(grid.DataString())
	}

}
