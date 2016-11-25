package sudokustate

import (
	"github.com/jkomoros/sudoku"
	"reflect"
	"testing"
)

func TestReset(t *testing.T) {
	model := &Model{}

	grid := sudoku.NewGrid()

	grid.MutableCell(3, 3).SetNumber(5)
	grid.LockFilledCells()

	snapshot := grid.Diagram(true)

	grid.MutableCell(4, 4).SetNumber(6)

	model.SetGrid(grid)

	if model.Grid().Cell(4, 4).Number() != 0 {
		t.Error("Expected grid to be reset after being set, but the unlocked cell remained:", model.Grid().Cell(4, 4).Number())
	}

	if model.snapshot != snapshot {
		t.Error("Got unexpected snapshot: got", model.snapshot, "expected", snapshot)
	}

}

func TestMarkMutator(t *testing.T) {
	model := &Model{}
	model.SetGrid(sudoku.NewGrid())

	cell := model.grid.MutableCell(0, 0)

	cell.SetMark(1, true)

	command := model.newMarkCommand(sudoku.CellRef{0, 0}, map[int]bool{1: true})

	if command != nil {
		t.Error("Got invalid command, expected nil", command)
	}

	command = model.newMarkCommand(sudoku.CellRef{0, 0}, map[int]bool{1: false, 2: true, 3: false})

	if command.Type() != "marks" {
		t.Error("Got unexpected type for marks command. Got", command.Type(), "expected 'marks'")
	}

	subCommands := command.SubCommands()

	if len(subCommands) != 1 {
		t.Fatal("Got wrong sized subcommands for marks", len(subCommands))
	}

	if subCommands[0] != command {
		t.Error("Sub command for marks was not self", subCommands[0])
	}

	if command.GroupInfo() != nil {
		t.Error("Expected nil info, got", command.GroupInfo())
	}

	command.Apply(model)

	if !cell.Marks().SameContentAs(sudoku.IntSlice{2}) {
		t.Error("Got wrong marks after mutating:", cell.Marks())
	}

	command.Undo(model)

	if !cell.Marks().SameContentAs(sudoku.IntSlice{1}) {
		t.Error("Got wrong marks after undoing:", cell.Marks())
	}

	if !reflect.DeepEqual(command.ModifiedCells(model), sudoku.CellRefSlice{sudoku.CellRef{0, 0}}) {
		t.Error("Didn't get right Modified Cells")
	}
}

func TestNumberMutator(t *testing.T) {
	model := &Model{}
	model.SetGrid(sudoku.NewGrid())

	cell := model.grid.Cell(0, 0)

	command := model.newNumberCommand(sudoku.CellRef{0, 0}, 0)

	if command != nil {
		t.Error("Got non-nil number command for a no op")
	}

	command = model.newNumberCommand(sudoku.CellRef{0, 0}, 1)

	if command.Type() != "number" {
		t.Error("Got unexpected type for number command. Got", command.Type(), "expected 'number'")
	}

	subCommands := command.SubCommands()

	if len(subCommands) != 1 {
		t.Fatal("Got wrong sized subcommands for number", len(subCommands))
	}

	if subCommands[0] != command {
		t.Error("Sub command for number was not self", subCommands[0])
	}

	if command.GroupInfo() != nil {
		t.Error("Expected nil info, got", command.GroupInfo())
	}

	command.Apply(model)

	if cell.Number() != 1 {
		t.Error("Number mutator didn't add the number")
	}

	command.Undo(model)

	if cell.Number() != 0 {
		t.Error("Number mutator didn't undo")
	}

	if !reflect.DeepEqual(command.ModifiedCells(model), sudoku.CellRefSlice{sudoku.CellRef{0, 0}}) {
		t.Error("Didn't get right Modified Cells")
	}

}

func TestGroups(t *testing.T) {
	model := &Model{}
	model.SetGrid(sudoku.NewGrid())

	model.SetNumber(sudoku.CellRef{0, 0}, 1)

	if model.grid.Cell(0, 0).Number() != 1 {
		t.Fatal("Setting number outside group didn't set number")
	}

	model.SetMarks(sudoku.CellRef{0, 1}, map[int]bool{3: true})

	if !model.grid.Cell(0, 1).Marks().SameContentAs(sudoku.IntSlice{3}) {
		t.Fatal("Setting marks outside of group didn't set marks")
	}

	state := model.grid.Diagram(true)

	if model.InGroup() {
		t.Error("Model reported being in group even though it wasn't")
	}

	model.StartGroup("foo")

	if !model.InGroup() {
		t.Error("Model didn't report being in a group even though it was")
	}

	model.SetNumber(sudoku.CellRef{0, 2}, 1)
	model.SetMarks(sudoku.CellRef{0, 3}, map[int]bool{3: true})

	if model.grid.Diagram(true) != state {
		t.Error("Within a group setnumber and setmarks mutated the grid")
	}

	model.FinishGroupAndExecute()

	command := model.currentCommand.c

	if command.Type() != "group" {
		t.Error("Got unexpected type for group command. Got", command.Type(), "expected 'group'")
	}

	subCommands := command.SubCommands()

	if len(subCommands) != 2 {
		t.Fatal("Got wrong sized subcommands for marks", len(subCommands))
	}

	if subCommands[0].Type() != "number" {
		t.Error("Sub command #1 was not number")
	}

	if subCommands[1].Type() != "marks" {
		t.Error("Sub command #2 was not marks")
	}

	if command.GroupInfo() == nil {
		t.Error("Expected non-nil info, got", command.GroupInfo())
	}

	if command.GroupInfo().ID != 0 {
		t.Error("Expected ID of 0, got", command.GroupInfo().ID)
	}

	if command.GroupInfo().Description != "foo" {
		t.Error("Expected description of 'foo', got:", command.GroupInfo().Description)
	}

	if model.InGroup() {
		t.Error("After finishing a group model still said it was in group")
	}

	if model.grid.Diagram(true) == state {
		t.Error("Commiting a group didn't mutate grid")
	}

	if model.grid.Cell(0, 2).Number() != 1 {
		t.Error("Commiting a group didn't set the number")
	}

	if !model.grid.Cell(0, 3).Marks().SameContentAs(sudoku.IntSlice{3}) {
		t.Error("Commiting a group didn't set the right marks")
	}

	model.Undo()

	if model.grid.Diagram(true) != state {
		t.Error("Undoing a group update didn't set the grid back to same state")
	}

	model.StartGroup("discardable")
	model.CancelGroup()

	if model.InGroup() {
		t.Error("After canceling a group the model still thought it was in one.")
	}
}

func TestUndoRedo(t *testing.T) {

	model := &Model{}
	model.SetGrid(sudoku.NewGrid())

	if model.Undo() {
		t.Error("Could undo on a fresh grid")
	}

	if model.Redo() {
		t.Error("Could redo on a fresh grid")
	}

	rememberedStates := []string{
		model.grid.Diagram(true),
	}

	rememberedModfiedCells := []sudoku.CellRefSlice{
		nil,
	}

	model.SetNumber(sudoku.CellRef{0, 0}, 1)

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))
	rememberedModfiedCells = append(rememberedModfiedCells, sudoku.CellRefSlice{sudoku.CellRef{0, 0}})

	model.SetNumber(sudoku.CellRef{0, 1}, 2)

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))
	rememberedModfiedCells = append(rememberedModfiedCells, sudoku.CellRefSlice{sudoku.CellRef{0, 1}})

	model.SetNumber(sudoku.CellRef{0, 0}, 3)

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))
	rememberedModfiedCells = append(rememberedModfiedCells, sudoku.CellRefSlice{sudoku.CellRef{0, 0}})

	model.SetMarks(sudoku.CellRef{0, 2}, map[int]bool{3: true, 4: true})

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))
	rememberedModfiedCells = append(rememberedModfiedCells, sudoku.CellRefSlice{sudoku.CellRef{0, 2}})

	model.SetMarks(sudoku.CellRef{0, 2}, map[int]bool{1: true, 4: false})

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))
	rememberedModfiedCells = append(rememberedModfiedCells, sudoku.CellRefSlice{sudoku.CellRef{0, 2}})

	if model.Redo() {
		t.Error("Able to redo even though at end")
	}

	for i := len(rememberedStates) - 1; i >= 1; i-- {
		if model.grid.Diagram(true) != rememberedStates[i] {
			t.Error("Remembere state wrong for state", i)
		}
		if !reflect.DeepEqual(model.LastModifiedCells(), rememberedModfiedCells[i]) {
			t.Error("Wrong last modified cells", i)
		}
		if !model.Undo() {
			t.Error("Couldn't undo early: ", i)
		}
	}

	//Verify we can't undo at beginning

	if model.Undo() {
		t.Error("Could undo even though it was the beginning.")
	}

	for i := 0; i < 3; i++ {
		if model.grid.Diagram(true) != rememberedStates[i] {
			t.Error("Remembered states wrong for state", i, "when redoing")
		}
		if !reflect.DeepEqual(model.LastModifiedCells(), rememberedModfiedCells[i]) {
			t.Error("Wrong last modified cells", i)
		}

		if !model.Redo() {
			t.Error("Unable to redo")
		}
	}

	model.SetNumber(sudoku.CellRef{2, 0}, 3)

	if model.Redo() {
		t.Error("Able to redo even though just spliced in a new move.")
	}

	//verify setting a new grid clears history

	model.SetGrid(sudoku.NewGrid())

	if model.Undo() {
		t.Error("Could undo on a new grid")
	}

	if model.Redo() {
		t.Error("Could undo on an old grid")
	}
}
