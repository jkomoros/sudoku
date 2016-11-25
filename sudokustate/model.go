/*

sudokustate manages modifications to a given grid, allowing easy undo/redo,
and keeping track of all moves.

Moves can be made either by setting a number on a given cell or setting
multiple marks on a given cell.

When multiple modifications are logically part of the same move (for the
purposes of undo/redo), a group can be created. Call `model.StartGroup`, make
the changes, then `model.FinishGroupAndExecute` to make all of the changes in
one 'move'.

*/
package sudokustate

import (
	"github.com/jkomoros/sudoku"
)

type Model struct {
	grid                   sudoku.MutableGrid
	currentCommand         *commandList
	commands               *commandList
	inProgressMultiCommand *multiCommand
}

type commandList struct {
	c    command
	next *commandList
	prev *commandList
}

type command interface {
	Apply(m *Model)
	Undo(m *Model)
	ModifiedCells(m *Model) sudoku.CellSlice
}

type baseCommand struct {
	row, col int
}

func (b *baseCommand) ModifiedCells(m *Model) sudoku.CellSlice {
	if m == nil || m.grid == nil {
		return nil
	}
	return sudoku.CellSlice{m.grid.Cell(b.row, b.col)}
}

func (m *multiCommand) ModifiedCells(model *Model) sudoku.CellSlice {
	var result sudoku.CellSlice

	for _, command := range m.commands {
		result = append(result, command.ModifiedCells(model)...)
	}

	return result
}

func (m *multiCommand) AddCommand(c command) {
	m.commands = append(m.commands, c)
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

type multiCommand struct {
	commands []command
}

//Grid returns the underlying Grid managed by this Model. It's a non-mutable
//reference to emphasize that all mutations should be done by the Model
//itself.
func (m *Model) Grid() sudoku.Grid {
	return m.grid
}

func (m *Model) executeCommand(c command) {
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

func (m *Model) LastModifiedCells() sudoku.CellSlice {
	if m.currentCommand == nil {
		return nil
	}

	return m.currentCommand.c.ModifiedCells(m)
}

//Undo returns true if there was something to undo.
func (m *Model) Undo() bool {
	if m.currentCommand == nil {
		return false
	}

	m.currentCommand.c.Undo(m)

	m.currentCommand = m.currentCommand.prev

	return true
}

//Redo returns true if there was something to redo.
func (m *Model) Redo() bool {

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

func (m *Model) StartGroup() {
	m.inProgressMultiCommand = &multiCommand{
		nil,
	}
}

func (m *Model) FinishGroupAndExecute() {
	if m.inProgressMultiCommand == nil {
		return
	}
	m.executeCommand(m.inProgressMultiCommand)
	m.inProgressMultiCommand = nil
}

func (m *Model) CancelGroup() {
	m.inProgressMultiCommand = nil
}

func (m *Model) InGroup() bool {
	return m.inProgressMultiCommand != nil
}

//Reset resets the current grid back to its fully unfilled states and discards
//all commands.
func (m *Model) Reset() {
	//TODO: test this.
	m.commands = nil
	m.currentCommand = nil
	if m.grid != nil {
		m.grid.ResetUnlockedCells()
	}
}

func (m *Model) SetGrid(grid sudoku.MutableGrid) {
	m.grid = grid
	m.Reset()
}

func (m *Model) SetMarks(row, col int, marksToggle map[int]bool) {
	//TODO: should this take a cellRef?
	command := m.newMarkCommand(row, col, marksToggle)
	if command == nil {
		return
	}
	if m.InGroup() {
		m.inProgressMultiCommand.AddCommand(command)
	} else {
		m.executeCommand(command)
	}
}

func (m *Model) newMarkCommand(row, col int, marksToggle map[int]bool) *markCommand {
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

func (m *Model) SetNumber(row, col int, num int) {
	//TODO: should this take CellRef?
	command := m.newNumberCommand(row, col, num)
	if command == nil {
		return
	}
	if m.InGroup() {
		m.inProgressMultiCommand.AddCommand(command)
	} else {
		m.executeCommand(command)
	}
}

func (m *Model) newNumberCommand(row, col int, num int) *numberCommand {
	cell := m.grid.Cell(row, col)

	if cell == nil {
		return nil
	}

	if cell.Number() == num {
		return nil
	}

	return &numberCommand{baseCommand{row, col}, num, cell.Number()}
}

func (m *markCommand) Apply(model *Model) {
	cell := model.grid.MutableCell(m.row, m.col)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		cell.SetMark(key, value)
	}
}

func (m *markCommand) Undo(model *Model) {
	cell := model.grid.MutableCell(m.row, m.col)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		//Set the opposite since we're undoing.
		cell.SetMark(key, !value)
	}
}

func (n *numberCommand) Apply(model *Model) {
	cell := model.grid.MutableCell(n.row, n.col)
	if cell == nil {
		return
	}
	cell.SetNumber(n.number)
}

func (n *numberCommand) Undo(model *Model) {
	cell := model.grid.MutableCell(n.row, n.col)
	if cell == nil {
		return
	}
	cell.SetNumber(n.oldNumber)
}

func (m *multiCommand) Apply(model *Model) {
	for _, command := range m.commands {
		command.Apply(model)
	}
}

func (m *multiCommand) Undo(model *Model) {
	for i := len(m.commands) - 1; i >= 0; i-- {
		m.commands[i].Undo(model)
	}
}
