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

type numberMutator struct {
	row, col int
	number   int
	//Necessary so we can undo.
	oldNumber int
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
	mutator := &numberMutator{row, col, num, cell.Number()}
	mutator.Apply(m)
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

func (n *numberMutator) Apply(model *model) {
	cell := model.grid.Cell(n.row, n.col)
	if cell == nil {
		return
	}
	cell.SetNumber(n.number)
}
