package main

import (
	"github.com/jkomoros/sudoku"
)

type model struct {
	grid *sudoku.Grid
}

type command interface {
	Apply(m *model)
	Undo(m *model)
}

type markCommand struct {
	row, col    int
	marksToggle map[int]bool
}

type numberCommand struct {
	row, col int
	number   int
	//Necessary so we can undo.
	oldNumber int
}

func (m *model) SetGrid(grid *sudoku.Grid) {
	m.grid = grid
}

func (m *model) SetMarks(row, col int, marksToggle map[int]bool) {
	command := m.newMarkCommand(row, col, marksToggle)
	if command == nil {
		return
	}
	command.Apply(m)
}

func (m *model) newMarkCommand(row, col int, marksToggle map[int]bool) *markCommand {
	//Only keep marks in the toggle that won't be a no-op
	newMarksToggle := make(map[int]bool)

	cell := m.grid.Cell(row, col)

	if cell == nil {
		return nil
	}
	for key, value := range marksToggle {
		if cell.Mark(key) != value {
			//Good, keep it
			newMarksToggle[key] = value
		}
	}

	if len(newMarksToggle) == 0 {
		//The command would be a no op!
		return nil
	}

	return &markCommand{row, col, newMarksToggle}
}

func (m *model) SetNumber(row, col int, num int) {
	command := m.newNumberCommand(row, col, num)
	if command == nil {
		return
	}
	command.Apply(m)
}

func (m *model) newNumberCommand(row, col int, num int) *numberCommand {
	cell := m.grid.Cell(row, col)

	if cell == nil {
		return nil
	}

	if cell.Number() == num {
		return nil
	}

	return &numberCommand{row, col, num, cell.Number()}
}

func (m *markCommand) Apply(model *model) {
	cell := model.grid.Cell(m.row, m.col)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		cell.SetMark(key, value)
	}
}

func (m *markCommand) Undo(model *model) {
	cell := model.grid.Cell(m.row, m.col)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		//Set the opposite since we're undoing.
		cell.SetMark(key, !value)
	}
}

func (n *numberCommand) Apply(model *model) {
	cell := model.grid.Cell(n.row, n.col)
	if cell == nil {
		return
	}
	cell.SetNumber(n.number)
}

func (n *numberCommand) Undo(model *model) {
	cell := model.grid.Cell(n.row, n.col)
	if cell == nil {
		return
	}
	cell.SetNumber(n.oldNumber)
}
