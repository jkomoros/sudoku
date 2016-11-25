package sudokustate

import (
	"github.com/jkomoros/sudoku"
)

type digest struct {
	Puzzle string
	Moves  []digestMove
}

type digestMove struct {
	Type   string
	Cell   sudoku.CellRef
	Marks  map[int]bool
	Time   int
	Number int
	Group  digestGroup
}

type digestGroup struct {
	Type string
	ID   int
}
