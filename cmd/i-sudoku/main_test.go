package main

import (
	"bytes"
	"flag"
	"github.com/jkomoros/sudoku"
	"github.com/jkomoros/sudoku/sdkconverter"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const FAST_MOVE_TEST_GRID = `.|.|.|.|1|.|.|.|.
.|.|.|.|1|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|1|.|.|.|.
1|1|.|1|.|1|.|1|1
.|.|.|.|1|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|1|.|.|.|.
.|.|.|.|1|.|.|.|.`

func TestNewGrid(t *testing.T) {
	model := newController()

	//We want to make sure that if a cell is selected and we make a new grid,
	//the m.selected cell is in the new grid, not the old. This is actually
	//hard to do. We set selected to 3,3, and then regenerate NewPuzzles until
	//the new puzzle has am empty cell there. Then we set that number, and
	//then make sure it was set.

	unfilledNumberFound := false
	for i := 0; i < 100; i++ {
		model.SetSelected(model.Grid().Cell(3, 3))
		model.NewGrid()
		if model.Grid().Cell(3, 3).Number() == 0 {
			//Found one!
			model.SetSelectedNumber(3)
			if model.Grid().Cell(3, 3).Number() != 3 {
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
	model := newController()

	if model.consoleMessage != "" {
		t.Error("Model started out with non-empty console message", model.consoleMessage)
	}

	model.SetConsoleMessage("Test", false)

	if model.consoleMessage != "Test" {
		t.Fatal("SetConsoleMessage didn't work.")
	}

	model.WillProcessEvent()

	if model.consoleMessage != "Test" {
		t.Error("A long lived console message didn't last past event loop.")
	}

	model.SetConsoleMessage("Short", true)

	if model.consoleMessage != "Short" {
		t.Error("Setting a short console message failed")
	}

	model.WillProcessEvent()

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
	model := newController()

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
	model := newController()
	next := model.Grid().Cell(2, 2)
	model.SetSelected(next)
	if model.Selected() != next {
		t.Error("Set selected didn't change the selected cell.")
	}
}

func TestMoveSelectionLeft(t *testing.T) {
	model := newController()

	if model.Selected() == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected())
	}

	model.MoveSelectionLeft(false)

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected after move left at bounds", model.Selected())
	}

	model.SetSelected(model.Grid().Cell(1, 1))

	model.MoveSelectionLeft(false)

	if model.Selected().Row() != 1 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected after move left", model.Selected())
	}

	//Test fast move
	model.SetGrid(sdkconverter.Load(FAST_MOVE_TEST_GRID))

	model.SetSelected(model.Grid().Cell(4, 4))

	model.MoveSelectionLeft(true)

	if model.Selected() != model.Grid().Cell(4, 2) {
		t.Error("Fast move didn't skip the locked cell", model.Selected())
	}
	//No more spots to the left; should stay still.
	model.MoveSelectionLeft(true)

	if model.Selected() != model.Grid().Cell(4, 2) {
		t.Error("Fast move moved even though no more cells left in that direction")
	}
}

func TestMoveSelectionRight(t *testing.T) {
	model := newController()

	if model.Selected() == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected())
	}

	model.MoveSelectionRight(false)

	if model.Selected().Row() != 0 || model.Selected().Col() != 1 {
		t.Error("Wrong cell selected after move right", model.Selected())
	}

	model.SetSelected(model.Grid().Cell(1, sudoku.DIM-1))

	model.MoveSelectionRight(false)

	if model.Selected().Row() != 1 || model.Selected().Col() != sudoku.DIM-1 {
		t.Error("Wrong cell selected after move right at bounds", model.Selected())
	}

	//Test fast move
	model.SetGrid(sdkconverter.Load(FAST_MOVE_TEST_GRID))

	model.SetSelected(model.Grid().Cell(4, 4))

	model.MoveSelectionRight(true)

	if model.Selected() != model.Grid().Cell(4, 6) {
		t.Error("Fast move didn't skip the locked cell", model.Selected())
	}
	//No more spots to the left; should stay still.
	model.MoveSelectionRight(true)

	if model.Selected() != model.Grid().Cell(4, 6) {
		t.Error("Fast move moved even though no more cells left in that direction")
	}
}

func TestMoveSelectionUp(t *testing.T) {
	model := newController()

	if model.Selected() == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected())
	}

	model.MoveSelectionUp(false)

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected after move up at bounds", model.Selected())
	}

	model.SetSelected(model.Grid().Cell(1, 1))

	model.MoveSelectionUp(false)

	if model.Selected().Row() != 0 || model.Selected().Col() != 1 {
		t.Error("Wrong cell selected after move up", model.Selected())
	}

	//Test fast move
	model.SetGrid(sdkconverter.Load(FAST_MOVE_TEST_GRID))

	model.SetSelected(model.Grid().Cell(4, 4))

	model.MoveSelectionUp(true)

	if model.Selected() != model.Grid().Cell(2, 4) {
		t.Error("Fast move didn't skip the locked cell", model.Selected())
	}
	//No more spots to the left; should stay still.
	model.MoveSelectionUp(true)

	if model.Selected() != model.Grid().Cell(2, 4) {
		t.Error("Fast move moved even though no more cells left in that direction")
	}
}

func TestMoveSelectionDown(t *testing.T) {
	model := newController()

	if model.Selected() == nil {
		t.Fatal("No selected cell")
	}

	if model.Selected().Row() != 0 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected to start", model.Selected())
	}

	model.MoveSelectionDown(false)

	if model.Selected().Row() != 1 || model.Selected().Col() != 0 {
		t.Error("Wrong cell selected after move down", model.Selected())
	}

	model.SetSelected(model.Grid().Cell(sudoku.DIM-1, 1))

	model.MoveSelectionDown(false)

	if model.Selected().Row() != sudoku.DIM-1 || model.Selected().Col() != 1 {
		t.Error("Wrong cell selected after move right at bounds", model.Selected())
	}

	//Test fast move
	model.SetGrid(sdkconverter.Load(FAST_MOVE_TEST_GRID))

	model.SetSelected(model.Grid().Cell(4, 4))

	model.MoveSelectionDown(true)

	if model.Selected() != model.Grid().Cell(6, 4) {
		t.Error("Fast move didn't skip the locked cell", model.Selected())
	}
	//No more spots to the left; should stay still.
	model.MoveSelectionDown(true)

	if model.Selected() != model.Grid().Cell(6, 4) {
		t.Error("Fast move moved even though no more cells left in that direction")
	}
}

func TestFillSelectedWithLegalMarks(t *testing.T) {
	model := newController()
	model.SetGrid(sudoku.NewGrid())

	model.ToggleSelectedMark(4)

	for i := 3; i < sudoku.DIM; i++ {
		cell := model.Grid().Cell(0, i)
		cell.SetNumber(i + 1)
	}

	model.FillSelectedWithLegalMarks()

	if !model.Selected().Marks().SameAs(sudoku.IntSlice{1, 2, 3}) {
		t.Error("Fill selected marks on a cell got wrong result", model.Selected().Marks())
	}

	model.ToggleSelectedMark(4)

	model.RemoveInvalidMarksFromSelected()

	if !model.Selected().Marks().SameAs(sudoku.IntSlice{1, 2, 3}) {
		t.Error("Remove invalid marks got wrong result", model.Selected().Marks())
	}
}

func TestEnsureGrid(t *testing.T) {
	model := newController()

	if model.Grid() == nil {
		t.Fatal("New model had no grid")
	}

	oldData := model.Grid().DataString()

	model.EnsureGrid()

	if model.Grid().DataString() != oldData {
		t.Error("Ensure grid blew away a grid")
	}

	model.SetGrid(nil)
	model.EnsureGrid()

	if model.Grid() == nil {
		t.Error("EnsureGrid didn't create a grid")
	}

	for i, cell := range model.Grid().Cells() {
		if cell.Number() != 0 && !cell.Locked() {
			t.Error("Grid from EnsureGrid didn't have numbers locked:", i, cell)
		}
	}
}

func TestSetSelectionNumber(t *testing.T) {
	model := newController()

	var lockedCell *sudoku.Cell
	var unlockedCell *sudoku.Cell

	//Set an unlocked cell
	for _, cell := range model.Grid().Cells() {
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

	if lockedCell == nil {
		t.Fatal("Couldn't find a locked cell to test")
	}

	if unlockedCell == nil {
		t.Fatal("Couldn't find an unlocked cell to test")
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
	model := newController()

	var lockedCell *sudoku.Cell
	var unlockedCell *sudoku.Cell

	//Set an unlocked cell
	for _, cell := range model.Grid().Cells() {
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

	if lockedCell == nil {
		t.Fatal("Couldn't find a locked cell to test")
	}

	if unlockedCell == nil {
		t.Fatal("Couldn't find an unlocked cell to test")
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

func TestSaveGrid(t *testing.T) {
	c := newController()

	testFileName := "test_puzzles/TEMPORARY_TEST_FILE.doku"

	//Make sure the test file doesn't exist.

	if _, err := os.Stat(testFileName); !os.IsNotExist(err) {
		//it does exist.
		os.Remove(testFileName)
	}

	c.SetFilename(testFileName)
	c.SaveGrid()

	gridState := c.Grid().Diagram(true)

	if _, err := os.Stat(testFileName); os.IsNotExist(err) {
		t.Fatal("Saving grid didn't actually save a file")
	}

	//Blow away the grid
	c.SetGrid(nil)

	c.LoadGridFromFile(testFileName)

	newGridState := c.Grid().Diagram(true)

	if gridState != newGridState {
		t.Error("The file that was saved didn't put the grid back in the same state.")
	}

	//Now remove the file
	os.Remove(testFileName)

}

//Callers should call fixUpOptions after receiving this.
func getDefaultOptions() *appOptions {
	options := &appOptions{
		flagSet: flag.NewFlagSet("main", flag.ExitOnError),
	}
	defineFlags(options)
	options.flagSet.Parse([]string{})
	return options
}

func TestProvideStarterPuzzle(t *testing.T) {
	options := getDefaultOptions()
	options.START_PUZZLE_FILENAME = "test_puzzles/converter_one.sdk"

	errOutput := &bytes.Buffer{}

	c := makeMainController(options, errOutput)

	errorReaderBytes, _ := ioutil.ReadAll(errOutput)

	errMessage := string(errorReaderBytes)

	goldenPuzzle, _ := ioutil.ReadFile("test_puzzles/converter_one.sdk")

	if c.Grid().DataString() != string(goldenPuzzle) {
		t.Error("Loading a normal puzzle with command line option failed.")
	}

	if errMessage != "" {
		t.Error("Got error message on successful read", errMessage)
	}

	if c.Filename() != options.START_PUZZLE_FILENAME {
		t.Error("Loading from file didn't set filename in controller")
	}
	//TODO: test that an invalid file errors
	//TODO: test that an invalid puzzle string errors
}
