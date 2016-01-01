package main

import (
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

func TestMoveSelection(t *testing.T) {
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
