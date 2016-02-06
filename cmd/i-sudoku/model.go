package main

import (
	"github.com/jkomoros/sudoku"
)

type model struct {
	grid           *sudoku.Grid
	currentCommand *commandList
	commands       *commandList
}

type commandList struct {
	c    command
	next *commandList
	prev *commandList
}

type command interface {
	Apply(m *model)
	Undo(m *model)
	ModifiedCells(m *model) sudoku.CellSlice
}

type baseCommand struct {
	row, col int
}

func (b *baseCommand) ModifiedCells(m *model) sudoku.CellSlice {
	if m == nil || m.grid == nil {
		return nil
	}
	return sudoku.CellSlice{m.grid.Cell(b.row, b.col)}
}

type markCommand struct {
	baseCommand
	marksToggle map[int]bool
}

type numberCommand struct {
	baseCommand
	number int
	//Necessary so we can undo.
	oldNumber int
}

func (m *model) executeCommand(c command) {
	listItem := &commandList{
		c:    c,
		next: nil,
		prev: m.currentCommand,
	}

	m.commands = listItem
	if m.currentCommand != nil {
		m.currentCommand.next = listItem
	}
	m.currentCommand = listItem

	c.Apply(m)
}

func (m *model) LastModifiedCells() sudoku.CellSlice {
	if m.currentCommand == nil {
		return nil
	}

	return m.currentCommand.c.ModifiedCells(m)
}

//Undo returns true if there was something to undo.
func (m *model) Undo() bool {
	if m.currentCommand == nil {
		return false
	}

	m.currentCommand.c.Undo(m)

	m.currentCommand = m.currentCommand.prev

	return true
}

//Redo returns true if there was something to redo.
func (m *model) Redo() bool {

	if m.commands == nil {
		return false
	}

	var commandToApply *commandList

	if m.currentCommand == nil {
		//If there is a non-nil commands, go all the way to the beginning,
		//because we're currently pointing at state 0
		commandToApply = m.commands
		for commandToApply.prev != nil {
			commandToApply = commandToApply.prev
		}
	} else {
		//Normaly operation is just to move to the next command in the list
		//and apply it.
		commandToApply = m.currentCommand.next
	}

	if commandToApply == nil {
		return false
	}

	m.currentCommand = commandToApply

	m.currentCommand.c.Apply(m)

	return true
}

func (m *model) SetGrid(grid *sudoku.Grid) {
	m.commands = nil
	m.currentCommand = nil
	m.grid = grid
}

func (m *model) SetMarks(row, col int, marksToggle map[int]bool) {
	command := m.newMarkCommand(row, col, marksToggle)
	if command == nil {
		return
	}
	m.executeCommand(command)
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

	return &markCommand{baseCommand{row, col}, newMarksToggle}
}

func (m *model) SetNumber(row, col int, num int) {
	command := m.newNumberCommand(row, col, num)
	if command == nil {
		return
	}
	m.executeCommand(command)
}

func (m *model) newNumberCommand(row, col int, num int) *numberCommand {
	cell := m.grid.Cell(row, col)

	if cell == nil {
		return nil
	}

	if cell.Number() == num {
		return nil
	}

	return &numberCommand{baseCommand{row, col}, num, cell.Number()}
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
