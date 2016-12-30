/*

sudokuhistory manages modifications to a given grid, allowing easy undo/redo,
and keeping track of all moves.

Moves can be made either by setting a number on a given cell or setting
multiple marks on a given cell.

When multiple modifications are logically part of the same move (for the
purposes of undo/redo), a group can be created. Call `model.StartGroup`, make
the changes, then `model.FinishGroupAndExecute` to make all of the changes in
one 'move'.

*/
package sudokuhistory

import (
	"github.com/jkomoros/sudoku"
	"github.com/jkomoros/sudoku/sdkconverter"
	"time"
)

//TODO: make sure that groups that change the same set of cells within a group
//apply/undo cleanly (does the undo need to go front to back or back to front?
//or perhaps do we need to not allow groups to contain multiple records with
//the same cell in them, to avoid this)

//Model maintains all of the state and modifications to the grid. The zero-
//state is valid; create a new Model with sudokuhistory.Model{}.
type Model struct {
	grid sudoku.MutableGrid
	//The place in the command list that we currently are (this could move
	//left or right down the list due to calls to undo or redo)
	currentCommand *commandList
	//The last command in the list of commands. currentCommand might not be
	//the same as commands if the user has called Undo some number of times.
	commands               *commandList
	inProgressGroupCommand *groupCommand
	//snapshot is a Diagram(true) of what the grid looked like when it was reset.
	snapshot string
	//started is the timestamp of when the grid was reset.
	started time.Time
}

type commandList struct {
	c    *groupCommand
	next *commandList
	prev *commandList
}

type groupCommand struct {
	subCommands []baseCommand
	description string
	//The duration of time since the started timestamp
	time time.Duration
}

type command interface {
	Apply(m *Model)
	Undo(m *Model)
	ModifiedCells(m *Model) sudoku.CellRefSlice
}

type baseCommand interface {
	command
	//The Marks (if this is a mark command). Returns nil otherwise.
	Marks() map[int]bool
	//The number associated with this command, or nil if not a number command.
	Number() *int
}

type basicCommand struct {
	ref sudoku.CellRef
}

func (b *basicCommand) Marks() map[int]bool {
	return nil
}

func (b *basicCommand) Number() *int {
	return nil
}

func (b *basicCommand) ModifiedCells(m *Model) sudoku.CellRefSlice {
	if m == nil || m.grid == nil {
		return nil
	}
	return sudoku.CellRefSlice{b.ref}
}

func (m *groupCommand) ModifiedCells(model *Model) sudoku.CellRefSlice {
	var result sudoku.CellRefSlice

	for _, command := range m.subCommands {
		result = append(result, command.ModifiedCells(model)...)
	}

	return result
}

func (m *groupCommand) Marks() map[int]bool {
	return nil
}

func (m *groupCommand) Number() *int {
	return nil
}

func (m *groupCommand) AddCommand(c baseCommand) {
	m.subCommands = append(m.subCommands, c)
}

type markCommand struct {
	basicCommand
	marksToggle map[int]bool
}

type numberCommand struct {
	basicCommand
	number int
	//Necessary so we can undo.
	oldNumber int
}

func (m *markCommand) Marks() map[int]bool {
	return m.marksToggle
}

func (n *numberCommand) Number() *int {
	//Indirect, since we're technically giving consumers a handle to change
	//it, which we don't want.
	result := n.number
	return &result
}

//Grid returns the underlying Grid managed by this Model. It's an immutable
//reference to emphasize that all mutations should be done by the Model
//itself.
func (m *Model) Grid() sudoku.Grid {
	return m.grid
}

func (m *Model) executeBaseCommand(c baseCommand) {
	if m.InGroup() {
		m.inProgressGroupCommand.AddCommand(c)
	} else {
		//If we're not manually in a group, we'll create a simple container
		//group and submit that.
		group := &groupCommand{
			[]baseCommand{c},
			"",
			0,
		}
		m.executeGroupCommand(group)
	}
}

func (m *Model) executeGroupCommand(c *groupCommand) {

	c.time = time.Since(m.started)

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

//LastModifiedCells returns the cells that were modified in the last action
//that was taken on the grid.
func (m *Model) LastModifiedCells() sudoku.CellRefSlice {
	if m.currentCommand == nil {
		return nil
	}

	return m.currentCommand.c.ModifiedCells(m)
}

//Undo rolls back a single action. It returns true if there was something to
//undo.
func (m *Model) Undo() bool {
	if m.currentCommand == nil {
		return false
	}

	m.currentCommand.c.Undo(m)

	m.currentCommand = m.currentCommand.prev

	return true
}

//Redo re-applies a single, previously undone, action. It returns true if
//there was something to redo.
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

//StartGroup begins a new group of actions. After calling StartGroup, all
//calls to SetMarks or SetNumber will be grouped into a single logical group
//of actions--that is, if they are undone they will all be undone at once.
//When a group is active, the modifications aren't actually made to the grid
//until FinishGroupAndExecute is called. Description is a description of what
//the group logically represents, primarily just for what will be shown in the
//digest.
func (m *Model) StartGroup(description string) {
	//TODO: allow setting a name when the group is created.
	m.inProgressGroupCommand = &groupCommand{
		nil,
		description,
		0,
	}
}

//FinishGroupAndExecute applies all of the modifications made inside of this
//group.
func (m *Model) FinishGroupAndExecute() {
	if m.inProgressGroupCommand == nil {
		return
	}
	m.executeGroupCommand(m.inProgressGroupCommand)
	m.inProgressGroupCommand = nil
}

//CancelGroup throws out the in-progress group of modifications. None of the
//modifications will actually be made to the grid.
func (m *Model) CancelGroup() {
	m.inProgressGroupCommand = nil
}

//InGroup returns true if a group is currently being built (that is,
//StartGroup was called, and neither FinishGroupAndExecute nor CancelGroup
//have been called.)
func (m *Model) InGroup() bool {
	return m.inProgressGroupCommand != nil
}

//Reset resets the current grid back to its fully unfilled state (resetting
//all state in non-locked cells)and discards all actions taken on the grid so
//far.
func (m *Model) Reset() {
	//TODO: test this.
	m.commands = nil
	m.currentCommand = nil
	if m.grid != nil {

		m.grid.ResetUnlockedCells()
		m.snapshot = sdkconverter.DataString(sdkconverter.DokuFormat, m.grid)
		m.started = time.Now()
	}
}

//SetGrid installs a new grid into the model, and then resets it. Note that
//the grid you set into the model will be used directly; modifications you
//make to it after installing into the model may get it in an odd state.
func (m *Model) SetGrid(grid sudoku.MutableGrid) {
	m.grid = grid
	m.Reset()
}

//SetMarks will set the specified marks onto a cell. An entry of True will set
//the mark, and entry of False will remove the mark in that slot. If InGroup()
//is false, the change will be done immediately; if InGroup() is true the
//change will not be applied until FinishGroupAndExecute() is called.
func (m *Model) SetMarks(ref sudoku.CellRef, marksToggle map[int]bool) {
	command := m.newMarkCommand(ref, marksToggle)
	if command == nil {
		return
	}
	m.executeBaseCommand(command)
}

func (m *Model) newMarkCommand(ref sudoku.CellRef, marksToggle map[int]bool) *markCommand {
	//Only keep marks in the toggle that won't be a no-op
	newMarksToggle := make(map[int]bool)

	cell := ref.Cell(m.grid)

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

	return &markCommand{basicCommand{ref}, newMarksToggle}
}

//SetNumber will set the specified number in a cell. If InGroup() is false,
//the change will be done immediately; if InGroup() is true the change will
//not be applied until FinishGroupAndExecute() is called.
func (m *Model) SetNumber(ref sudoku.CellRef, num int) {
	command := m.newNumberCommand(ref, num)
	if command == nil {
		return
	}
	m.executeBaseCommand(command)
}

func (m *Model) newNumberCommand(ref sudoku.CellRef, num int) *numberCommand {
	cell := ref.Cell(m.grid)

	if cell == nil {
		return nil
	}

	if cell.Number() == num {
		return nil
	}

	return &numberCommand{basicCommand{ref}, num, cell.Number()}
}

func (m *markCommand) Apply(model *Model) {
	cell := m.ref.MutableCell(model.grid)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		cell.SetMark(key, value)
	}
}

func (m *markCommand) Undo(model *Model) {
	cell := m.ref.MutableCell(model.grid)
	if cell == nil {
		return
	}
	for key, value := range m.marksToggle {
		//Set the opposite since we're undoing.
		cell.SetMark(key, !value)
	}
}

func (n *numberCommand) Apply(model *Model) {
	cell := n.ref.MutableCell(model.grid)
	if cell == nil {
		return
	}
	cell.SetNumber(n.number)
}

func (n *numberCommand) Undo(model *Model) {
	cell := n.ref.MutableCell(model.grid)
	if cell == nil {
		return
	}
	cell.SetNumber(n.oldNumber)
}

func (m *groupCommand) Apply(model *Model) {
	for _, command := range m.subCommands {
		command.Apply(model)
	}
}

func (m *groupCommand) Undo(model *Model) {
	for i := len(m.subCommands) - 1; i >= 0; i-- {
		m.subCommands[i].Undo(model)
	}
}
