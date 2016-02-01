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
