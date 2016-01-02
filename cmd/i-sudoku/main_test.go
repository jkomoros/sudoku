package main

import (
	"github.com/jkomoros/sudoku"
	"strings"
	"testing"
)

func TestNewGrid(t *testing.T) {
	model := newModel()

	//We want to make sure that if a cell is selected and we make a new grid,
	//the m.selected cell is in the new grid, not the old. This is actually
	//hard to do. We set selected to 3,3, and then regenerate NewPuzzles until
	//the new puzzle has am empty cell there. Then we set that number, and
	//then make sure it was set.

	unfilledNumberFound := false
	for i := 0; i < 100; i++ {
		model.SetSelected(model.grid.Cell(3, 3))
		model.NewGrid()
		if model.grid.Cell(3, 3).Number() == 0 {
			//Found one!
			model.SetSelectedNumber(3)
			if model.grid.Cell(3, 3).Number() != 3 {
				t.Error("When creating a new grid, the selected cell was in the old grid.")
			}
			unfilledNumberFound = true
			break
		}
	}
	if !unfilledNumberFound {
		t.Error("In 100 times, did not find a new grid where 3,3 was unfilled.")
	}
}

func TestConsoleMessage(t *testing.T) {
	model := newModel()

	if model.consoleMessage != "" {
		t.Error("Model started out with non-empty console message", model.consoleMessage)
	}

	model.SetConsoleMessage("Test", false)

	if model.consoleMessage != "Test" {
		t.Fatal("SetConsoleMessage didn't work.")
	}

	model.EndOfEventLoop()

	if model.consoleMessage != "Test" {
		t.Error("A long lived console message didn't last past event loop.")
	}

	model.SetConsoleMessage("Short", true)

	if model.consoleMessage != "Short" {
		t.Error("Setting a short console message failed")
	}

	model.EndOfEventLoop()

	if model.consoleMessage != "" {
		t.Error("A short lived console message wasn't cleared at end of event loop.")
	}

	//Test wrapping up long messages
	model.outputWidth = 30

	model.SetConsoleMessage(MARKS_MODE_FAIL_NUMBER+MARKS_MODE_FAIL_NUMBER+MARKS_MODE_FAIL_NUMBER, false)

	for i, line := range strings.Split(model.consoleMessage, "\n") {
		if len(line) > 30 {
			t.Error("Line", i, "of long output is greater than 30 chars")
		}
	}

	model.ClearConsole()

	if model.consoleMessage != "" {
		t.Error("m.ClearConsole didn't clear the console.")
	}

}

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

	model.SetSelectedNumber(1)
	model.SetSelectedNumber(1)

	if model.Selected().Number() != 0 {
		t.Error("Setting the same number on the selected cell didn't set it back to 0", model.Selected())
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
