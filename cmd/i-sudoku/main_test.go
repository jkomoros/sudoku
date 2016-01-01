package main

import (
	"github.com/jkomoros/sudoku"
	"testing"
)

func TestEnsureSelected(t *testing.T) {
	model := newModel()

	if model.Selected == nil {
		t.Error("New model had no selected cell")
	}

	model.Selected = nil
	model.EnsureSelected()

	if model.Selected == nil {
		t.Error("Model after EnsureSelected still had no selected cell")
	}
}

func TestMoveSelectionLeft(t *testing.T) {
	model := newModel()

	if model.Selected == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected.Row() != 0 || model.Selected.Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected)
	}

	model.MoveSelectionLeft()

	if model.Selected.Row() != 0 || model.Selected.Col() != 0 {
		t.Error("Wrong cell selected after move left at bounds", model.Selected)
	}

	model.Selected = model.grid.Cell(1, 1)

	model.MoveSelectionLeft()

	if model.Selected.Row() != 1 || model.Selected.Col() != 0 {
		t.Error("Wrong cell selected after move left", model.Selected)
	}
}

func TestMoveSelectionRight(t *testing.T) {
	model := newModel()

	if model.Selected == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected.Row() != 0 || model.Selected.Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected)
	}

	model.MoveSelectionRight()

	if model.Selected.Row() != 0 || model.Selected.Col() != 1 {
		t.Error("Wrong cell selected after move right", model.Selected)
	}

	model.Selected = model.grid.Cell(1, sudoku.DIM-1)

	model.MoveSelectionRight()

	if model.Selected.Row() != 1 || model.Selected.Col() != sudoku.DIM-1 {
		t.Error("Wrong cell selected after move right at bounds", model.Selected)
	}
}

func TestMoveSelectionUp(t *testing.T) {
	model := newModel()

	if model.Selected == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected.Row() != 0 || model.Selected.Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected)
	}

	model.MoveSelectionUp()

	if model.Selected.Row() != 0 || model.Selected.Col() != 0 {
		t.Error("Wrong cell selected after move up at bounds", model.Selected)
	}

	model.Selected = model.grid.Cell(1, 1)

	model.MoveSelectionUp()

	if model.Selected.Row() != 0 || model.Selected.Col() != 1 {
		t.Error("Wrong cell selected after move up", model.Selected)
	}
}

func TestMoveSelectionDown(t *testing.T) {
	model := newModel()

	if model.Selected == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected.Row() != 0 || model.Selected.Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected)
	}

	model.MoveSelectionDown()

	if model.Selected.Row() != 1 || model.Selected.Col() != 0 {
		t.Error("Wrong cell selected after move down", model.Selected)
	}

	model.Selected = model.grid.Cell(sudoku.DIM-1, 1)

	model.MoveSelectionDown()

	if model.Selected.Row() != sudoku.DIM-1 || model.Selected.Col() != 1 {
		t.Error("Wrong cell selected after move right at bounds", model.Selected)
	}
}

func TestEnsureGrid(t *testing.T) {
	model := newModel()

	if model.grid == nil {
		t.Fatal("New model had no grid")
	}

	oldData := model.grid.DataString()

	model.EnsureGrid()

	if model.grid.DataString() != oldData {
		t.Error("Ensure grid blew away a grid")
	}

	model.grid = nil

	model.EnsureGrid()

	if model.grid == nil {
		t.Error("EnsureGrid didn't create a grid")
	}
}
