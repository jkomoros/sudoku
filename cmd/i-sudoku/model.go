package main

import (
	"github.com/jkomoros/sudoku"
)

type model struct {
	grid *sudoku.Grid
}

func (m *model) SetGrid(grid *sudoku.Grid) {
	m.grid = grid
}

func (m *model) SetMarks(row, col int, marksToggle map[int]bool) {
	cell := m.grid.Cell(row, col)
	if cell == nil {
		return
	}
	for key, value := range marksToggle {
		cell.SetMark(key, value)
	}
}

func (m *model) SetNumber(row, col int, num int) {
	cell := m.grid.Cell(row, col)
	if cell == nil {
		return
	}
	cell.SetNumber(num)
}
