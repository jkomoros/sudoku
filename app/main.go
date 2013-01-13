package main

import (
	"dokugen/sudoku"
)

//TODO: let people pass in puzzles to solve.
//TODO: let people pass in a filename to export to.

func main() {
	grid := sudoku.GenerateGrid()
	print(grid.DataString())
}
