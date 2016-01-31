package main

import (
	"github.com/jkomoros/sudoku"
)

type model struct {
	grid *sudoku.Grid
}

type modelMutator interface {
	Apply(m *model)
	Undo(m *model)
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
	//TODO: add a copy of marksToggle that only has valid marks for the current state.
	//right now if marksToggle has a no-op instruction, it can get in weird states.
	mutator := &markMutator{row, col, marksToggle}
	mutator.Apply(m)
}

func (m *model) SetNumber(row, col int, num int) {
	//TODO: this should also only add an action if the specified cell is not already num.
	cell := m.grid.Cell(row, col)
	if cell == nil {
		return
	}
	mutator := &numberMutator{row, col, num, cell.Number()}
	mutator.Apply(m)
}

//TODO: implement model.new{Mark|Number}Mutator, which only return a
//modelMutator if it wouldn't be a no-op. Then, test that they return nil if
//it would be a no op, including omitting marks that would be a no op.

func (m *markMutator) Apply(model *model) {
	cell := model.grid.Cell(m.row, m.col)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		cell.SetMark(key, value)
	}
}

func (m *markMutator) Undo(model *model) {
	cell := model.grid.Cell(m.row, m.col)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		//Set the opposite since we're undoing.
		cell.SetMark(key, !value)
	}
}

func (n *numberMutator) Apply(model *model) {
	cell := model.grid.Cell(n.row, n.col)
	if cell == nil {
		return
	}
	cell.SetNumber(n.number)
}

func (n *numberMutator) Undo(model *model) {
	//TODO: test this method
	cell := model.grid.Cell(n.row, n.col)
	if cell == nil {
		return
	}
	cell.SetNumber(n.oldNumber)
}
