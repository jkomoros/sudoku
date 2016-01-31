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

	mutator := markCommand{0, 0, map[int]bool{1: false, 2: true}}

	mutator.Apply(model)

	if !cell.Marks().SameContentAs(sudoku.IntSlice{2}) {
		t.Error("Got wrong marks after mutating:", cell.Marks())
	}

	mutator.Undo(model)

	if !cell.Marks().SameContentAs(sudoku.IntSlice{1}) {
		t.Error("Got wrong marks after undoing:", cell.Marks())
	}
}

func TestNumberMutator(t *testing.T) {
	model := &model{}
	model.SetGrid(sudoku.NewGrid())

	cell := model.grid.Cell(0, 0)

	mutator := numberCommand{0, 0, 1, 0}

	mutator.Apply(model)

	if cell.Number() != 1 {
		t.Error("Number mutator didn't add the number")
	}

	mutator.Undo(model)

	if cell.Number() != 0 {
		t.Error("Number mutator didn't undo")
	}

}
