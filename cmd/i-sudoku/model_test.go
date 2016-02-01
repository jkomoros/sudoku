package main

import (
	"github.com/jkomoros/sudoku"
	"testing"
)

func TestMarkMutator(t *testing.T) {
	model := &model{}
	model.SetGrid(sudoku.NewGrid())

	cell := model.grid.Cell(0, 0)

	cell.SetMark(1, true)

	command := model.newMarkCommand(0, 0, map[int]bool{1: true})

	if command != nil {
		t.Error("Got invalid command, expected nil", command)
	}

	command = model.newMarkCommand(0, 0, map[int]bool{1: false, 2: true, 3: false})

	command.Apply(model)

	if !cell.Marks().SameContentAs(sudoku.IntSlice{2}) {
		t.Error("Got wrong marks after mutating:", cell.Marks())
	}

	command.Undo(model)

	if !cell.Marks().SameContentAs(sudoku.IntSlice{1}) {
		t.Error("Got wrong marks after undoing:", cell.Marks())
	}
}

func TestNumberMutator(t *testing.T) {
	model := &model{}
	model.SetGrid(sudoku.NewGrid())

	cell := model.grid.Cell(0, 0)

	command := model.newNumberCommand(0, 0, 0)

	if command != nil {
		t.Error("Got non-nil number command for a no op")
	}

	command = model.newNumberCommand(0, 0, 1)

	command.Apply(model)

	if cell.Number() != 1 {
		t.Error("Number mutator didn't add the number")
	}

	command.Undo(model)

	if cell.Number() != 0 {
		t.Error("Number mutator didn't undo")
	}

}

func TestUndoRedo(t *testing.T) {

	model := &model{}
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

	model.SetNumber(0, 0, 1)

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))

	model.SetNumber(0, 1, 2)

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))

	model.SetNumber(0, 0, 3)

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))

	model.SetMarks(0, 2, map[int]bool{3: true, 4: true})

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))

	model.SetMarks(0, 2, map[int]bool{1: true, 4: false})

	rememberedStates = append(rememberedStates, model.grid.Diagram(true))

	if model.Redo() {
		t.Error("Able to redo even though at end")
	}

	for i := len(rememberedStates) - 1; i >= 1; i-- {
		if model.grid.Diagram(true) != rememberedStates[i] {
			t.Error("Remembere state wrong for state", i)
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

		if !model.Redo() {
			t.Error("Unable to redo")
		}
	}

	model.SetNumber(2, 0, 3)

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
