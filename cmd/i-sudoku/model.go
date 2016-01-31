package main

import (
	"github.com/jkomoros/sudoku"
)

type model struct {
	grid *sudoku.Grid
}

type modelMutator interface {
	Apply(m *model)
	//TODO: add an Undo method
}

type markMutator struct {
	row, col    int
	marksToggle map[int]bool
}

func (m *model) SetGrid(grid *sudoku.Grid) {
	m.grid = grid
}

func (m *model) SetMarks(row, col int, marksToggle map[int]bool) {
	mutator := &markMutator{row, col, marksToggle}
	mutator.Apply(m)
}

func (m *model) SetNumber(row, col int, num int) {
	cell := m.grid.Cell(row, col)
	if cell == nil {
		return
	}
	cell.SetNumber(num)
}

func (m *markMutator) Apply(model *model) {
	cell := model.grid.Cell(m.row, m.col)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		cell.SetMark(key, value)
	}
}
