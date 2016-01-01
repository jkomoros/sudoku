package main

import (
	"github.com/jkomoros/sudoku"
	"testing"
)

func TestEnsureSelected(t *testing.T) {
	model := newModel()

	if model.Selected() == nil {
		t.Error("New model had no selected cell")
	}

	model.SetSelected(nil)
	model.EnsureSelected()

	if model.Selected() == nil {
		t.Error("Model after EnsureSelected still had no selected cell")
	}
}

func TestSelected(t *testing.T) {
	model := newModel()
	next := model.grid.Cell(2, 2)
	model.SetSelected(next)
	if model.Selected() != next {
		t.Error("Set selected didn't change the selected cell.")
	}
}

func TestMoveSelectionLeft(t *testing.T) {
	model := newModel()

	if model.Selected() == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected())
	}

	model.MoveSelectionLeft()

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected after move left at bounds", model.Selected())
	}

	model.SetSelected(model.grid.Cell(1, 1))

	model.MoveSelectionLeft()

	if model.Selected().Row() != 1 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected after move left", model.Selected())
	}
}

func TestMoveSelectionRight(t *testing.T) {
	model := newModel()

	if model.Selected() == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected())
	}

	model.MoveSelectionRight()

	if model.Selected().Row() != 0 || model.Selected().Col() != 1 {
		t.Error("Wrong cell selected after move right", model.Selected())
	}

	model.SetSelected(model.grid.Cell(1, sudoku.DIM-1))

	model.MoveSelectionRight()

	if model.Selected().Row() != 1 || model.Selected().Col() != sudoku.DIM-1 {
		t.Error("Wrong cell selected after move right at bounds", model.Selected())
	}
}

func TestMoveSelectionUp(t *testing.T) {
	model := newModel()

	if model.Selected() == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected())
	}

	model.MoveSelectionUp()

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected after move up at bounds", model.Selected())
	}

	model.SetSelected(model.grid.Cell(1, 1))

	model.MoveSelectionUp()

	if model.Selected().Row() != 0 || model.Selected().Col() != 1 {
		t.Error("Wrong cell selected after move up", model.Selected())
	}
}

func TestMoveSelectionDown(t *testing.T) {
	model := newModel()

	if model.Selected() == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected())
	}

	model.MoveSelectionDown()

	if model.Selected().Row() != 1 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected after move down", model.Selected())
	}

	model.SetSelected(model.grid.Cell(sudoku.DIM-1, 1))

	model.MoveSelectionDown()

	if model.Selected().Row() != sudoku.DIM-1 || model.Selected().Col() != 1 {
		t.Error("Wrong cell selected after move right at bounds", model.Selected())
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

	for i, cell := range model.grid.Cells() {
		if cell.Number() != 0 && !cell.Locked() {
			t.Error("Grid from EnsureGrid didn't have numbers locked:", i, cell)
		}
	}
}

func TestSetSelectionNumber(t *testing.T) {
	model := newModel()

	var lockedCell *sudoku.Cell
	var unlockedCell *sudoku.Cell

	//Set an unlocked cell
	for _, cell := range model.grid.Cells() {
		if cell.Locked() {
			if lockedCell == nil {
				lockedCell = cell
			}
		}
		if !cell.Locked() {
			if unlockedCell == nil {
				unlockedCell = cell
			}
		}
		if lockedCell != nil && unlockedCell != nil {
			break
		}
	}

	model.SetSelected(unlockedCell)

	model.SetSelectedNumber(1)

	if model.Selected().Number() != 1 {
		t.Error("SetSelectionNumber didn't set to 1", model.Selected())
	}

	model.SetSelectedNumber(0)

	if model.Selected().Number() != 0 {
		t.Error("SetSelectionNumber didn't set to 0", model.Selected())
	}

	num := lockedCell.Number()

	//Pick a number that's not the one the cell is set to.
	numToSet := num + 1
	if numToSet >= sudoku.DIM {
		numToSet = 0
	}

	model.SetSelected(lockedCell)

	model.SetSelectedNumber(numToSet)

	if model.Selected().Number() == numToSet {
		t.Error("SetSelectionNumber modified a locked cell", lockedCell, num)
	}

}

func TestToggleSelectedMark(t *testing.T) {
	model := newModel()

	var lockedCell *sudoku.Cell
	var unlockedCell *sudoku.Cell

	//Set an unlocked cell
	for _, cell := range model.grid.Cells() {
		if cell.Locked() {
			if lockedCell == nil {
				lockedCell = cell
			}
		}
		if !cell.Locked() {
			if unlockedCell == nil {
				unlockedCell = cell
			}
		}
		if lockedCell != nil && unlockedCell != nil {
			break
		}
	}

	model.SetSelected(unlockedCell)

	model.ToggleSelectedMark(1)

	if !model.Selected().Mark(1) {
		t.Error("ToggleSelectedMark didn't mark 1", model.Selected())
	}

	model.ToggleSelectedMark(1)

	if model.Selected().Mark(1) {
		t.Error("ToggleSelectedMark didn't unmark 1", model.Selected())
	}

	model.SetSelected(lockedCell)

	model.ToggleSelectedMark(1)

	if model.Selected().Mark(1) {
		t.Error("ToggleSelectedMark modified a locked cell", lockedCell)
	}
}

func TestMode(t *testing.T) {
	model := newModel()

	//Add empty grid.
	model.grid = sudoku.NewGrid()
	model.SetSelected(nil)

	if model.StatusLine() != STATUS_DEFAULT {
		t.Error("Didn't get default status line in default mode.")
	}

	model.ModeInputNumber(1)

	if model.Selected().Number() != 1 {
		t.Error("InputNumber in default mode didn't add a number")
	}

	model.MoveSelectionRight()

	model.ModeEnterMarkMode()
	if model.StatusLine() != STATUS_MARKING+"[]"+STATUS_MARKING_POSTFIX {
		t.Error("In mark mode with no marks, didn't get expected", model.StatusLine())
	}
	model.ModeInputNumber(1)
	model.ModeInputNumber(2)
	if model.StatusLine() != STATUS_MARKING+"[1 2]"+STATUS_MARKING_POSTFIX {
		t.Error("In makr mode with two marks, didn't get expected", model.StatusLine())
	}
	model.ModeCommitMarkMode()
	if model.StatusLine() != STATUS_DEFAULT {
		t.Error("After commiting marks, didn't have default status", model.StatusLine())
	}

	if model.Selected().Number() != 0 {
		t.Error("InputNumber in mark mode set the number", model.Selected())
	}

	if !model.Selected().Mark(1) {
		t.Error("InputNumber in mark mode didn't set the first mark", model.Selected())
	}

	if !model.Selected().Mark(2) {
		t.Error("InputNumber in mark mode didn't set the second mark", model.Selected())
	}

	model.MoveSelectionRight()

	model.ModeEnterMarkMode()
	model.ModeInputNumber(1)
	model.ModeInputNumber(2)
	model.ModeCancelMarkMode()

	if model.StatusLine() != STATUS_DEFAULT {
		t.Error("After canceling mark mode, status didn't go back to default.", model.StatusLine())
	}

	if model.Selected().Mark(1) || model.Selected().Mark(2) {
		t.Error("InputNumber in canceled mark mode still set marks")
	}

	model.MoveSelectionRight()

	model.ModeInputNumber(1)

	if model.Selected().Number() != 1 {
		t.Error("InputNumber after cancled mark and another InputNum didn't set num", model.Selected())
	}

	if !model.ModeInputEsc() {
		t.Error("ModeInputEsc not in mark enter mode didn't tell us to quit.")
	}

	model.MoveSelectionRight()

	model.ModeEnterMarkMode()
	if model.ModeInputEsc() {
		t.Error("ModeInputEsc in mark enter mode DID tell us to quit")
	}
	if model.marksToInput != nil {
		t.Error("ModeInputEsc in mark enter mode didn't exit mark enter mode")
	}

}
