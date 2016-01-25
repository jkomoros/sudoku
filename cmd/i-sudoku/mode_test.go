package main

import (
	"github.com/jkomoros/sudoku"
	"github.com/nsf/termbox-go"
	"strconv"
	"testing"
	"unicode/utf8"
)

func sendKeyEvent(c *mainController, k termbox.Key) {
	evt := termbox.Event{
		Type: termbox.EventKey,
		Key:  k,
	}
	c.mode.handleInput(c, evt)
}

func sendNumberEvent(c *mainController, num int) {
	ch, _ := utf8.DecodeRuneInString(strconv.Itoa(num))
	evt := termbox.Event{
		Type: termbox.EventKey,
		Ch:   ch,
	}
	c.mode.handleInput(c, evt)
}

//TODO: use sendCharEvent to verify that chars in all states do what they should.
func sendCharEvent(c *mainController, ch rune) {
	evt := termbox.Event{
		Type: termbox.EventKey,
		Ch:   ch,
	}
	c.mode.handleInput(c, evt)
}

func TestDefaultMode(t *testing.T) {
	model := newController()
	//Add empty grid.
	model.SetGrid(sudoku.NewGrid())

	if model.mode != MODE_DEFAULT {
		t.Error("model didn't start in default state")
	}

	if MODE_DEFAULT.statusLine(model) != STATUS_DEFAULT {
		t.Error("Didn't get default status line in default mode.")
	}

	if model.mode.cursorLocation(model) != -1 {
		t.Error("Cursor not off screen in default mode.")
	}

	sendNumberEvent(model, 1)

	if model.Selected().Number() != 1 {
		t.Error("InputNumber in default mode didn't add a number")
	}

	model.MoveSelectionRight(false)

	sendCharEvent(model, '!')

	if !model.Selected().Mark(1) {
		t.Error("Sending a shifted 1 on a cell didn't turn on the 1 mark")
	}

	sendCharEvent(model, '!')

	if model.Selected().Mark(1) {
		t.Error("Sending a shifted 1 on a cell with a 1 mark didn't remove it")
	}

	sendKeyEvent(model, termbox.KeyEsc)

	if model.exitNow {
		t.Error("ModeInputEsc in DEFAULT_STATE did tell us to quit, but it shouldn't")
	}

	//Reset to a normal, solvable grid.

	model = newController()

	oldSelected := model.Selected()

	sendCharEvent(model, 'h')

	if model.consoleMessage == "" {
		t.Error("Asking for hint didn't give a hint")
	}

	if model.consoleMessageShort {
		t.Error("The hint message was short")
	}

	if model.mode != MODE_DEFAULT {
		t.Error("Choosing hint didn't lead back to default mode")
	}

	//Technically, this test has a 1/81 % chance of flaking...
	//TODO: make this test not flaky
	if oldSelected == model.Selected() {
		t.Error("Hint didn't automatically select the cell specified by the hint.")
	}

	lastStep := model.lastShownHint.Steps[len(model.lastShownHint.Steps)-1]
	correctNum := lastStep.TargetNums[0]
	wrongNum := correctNum + 1
	if wrongNum > sudoku.DIM {
		wrongNum = 1
	}
	model.SetSelectedNumber(wrongNum)

	if model.consoleMessage == "" {
		t.Error("Console message was cleared even though wrong number was entered")
	}

	model.SetSelectedNumber(correctNum)

	if model.consoleMessage != "" {
		t.Error("Console was not cleared after right hint number was entered.")
	}

	sendCharEvent(model, 'h')

	hintCell := model.Selected()

	//Move to a different cell to confirm that 'ENTER' reselects the cell and fills the number.
	if hintCell.Row() < sudoku.DIM-1 {
		model.MoveSelectionDown(false)
	} else {
		model.MoveSelectionUp(false)
	}

	sendKeyEvent(model, termbox.KeyEnter)

	if model.Selected() != hintCell {
		t.Error("Wrong cell had hint entered")
	}

	if model.Selected().Number() == 0 {
		t.Error("Accepting hint didn't fill the cell")
	}

	sendCharEvent(model, 'f')

	if !model.FastMode() {
		t.Error("'f' from default mode didn't enter fast move mode")
	}

	sendCharEvent(model, 'f')

	if model.FastMode() {
		t.Error("'f' again didn't turn off fast move mode")
	}
}

func TestReset(t *testing.T) {
	m := newController()

	var cell *sudoku.Cell

	//find an unfilled cell.
	for _, candidateCell := range m.Grid().Cells() {
		if candidateCell.Locked() {
			continue
		}
		cell = candidateCell
		break
	}
	//Enter command mode
	sendCharEvent(m, 'c')

	cell.SetNumber(3)

	sendCharEvent(m, 'r')

	if cell.Number() == 0 {
		t.Error("The grid was reset without confirmation")
	}

	sendCharEvent(m, 'y')

	if cell.Number() != 0 {
		t.Error("After confirming reset it wasn't.")
	}
}

func TestHintOnSolvedGrid(t *testing.T) {
	//This used to crash before we fixed it, so adding a regression test.
	model := newController()
	grid := sudoku.NewGrid()
	grid.Fill()
	model.SetGrid(grid)

	showHint(model)

	//If we didn't crash, we're good.

}

func TestSingleMarkEnter(t *testing.T) {
	model := newController()

	model.SetGrid(sudoku.NewGrid())

	//Test that it fails on 0 mark cell

	sendKeyEvent(model, termbox.KeyEnter)

	if model.Selected().Number() != 0 {
		t.Error("Enter on 0 mark cell put a mark")
	}

	if model.consoleMessage != SINGLE_FILL_MORE_THAN_ONE_MARK {
		t.Error("Enter on 0 mark cell did not print error message")
	}

	//Test that it works on one mark cell

	model.Selected().SetMark(1, true)

	sendKeyEvent(model, termbox.KeyEnter)

	if model.Selected().Number() == 0 {
		t.Error("Enter on 1 mark cell set wrong num marks")
	}

	if model.Selected().Number() != 1 {
		t.Error("Enter on 1 mark cell didn't set right mark")
	}

	//Test that it fails on two mark cells

	model.MoveSelectionRight(false)

	model.Selected().SetMark(1, true)
	model.Selected().SetMark(2, true)

	sendKeyEvent(model, termbox.KeyEnter)

	if model.Selected().Number() != 0 {
		t.Error("Enter on 2 mark cell put a mark: ", model.Selected())
	}

	if model.consoleMessage != SINGLE_FILL_MORE_THAN_ONE_MARK {
		t.Error("Enter on 2 mark cell did not print error message")
	}

	//Test that it fails on a locked cell

	model.MoveSelectionRight(false)
	model.SetSelectedNumber(1)
	model.Selected().Lock()
	model.Selected().SetMark(2, true)

	sendKeyEvent(model, termbox.KeyEnter)

	if model.Selected().Number() != 1 {
		t.Error("Enter on locked cell put a mark")
	}

	if model.consoleMessage != DEFAULT_MODE_FAIL_LOCKED {
		t.Error("Enter on locked cell did not print error message:", model.consoleMessage)
	}

	//Now, test that if you hit enter on a hint cell it fills hint.

	model = newController()

	sendCharEvent(model, 'h')

	lastStep := model.lastShownHint.Steps[len(model.lastShownHint.Steps)-1]
	num := lastStep.TargetNums[0]

	markNum := num + 1

	if markNum > sudoku.DIM {
		markNum = 1
	}

	//hint cell is now selected.

	model.ToggleSelectedMark(markNum)

	sendKeyEvent(model, termbox.KeyEnter)

	if model.Selected().Number() == markNum {
		t.Error("Enter on hint cell set single mark num")
	}

	if model.Selected().Number() != num {
		t.Error("Enter on hint cell with a single mark did not fill hint.")
	}
}

func TestEnterMarksMode(t *testing.T) {
	model := newController()
	//Add empty grid.
	model.SetGrid(sudoku.NewGrid())

	model.ToggleMarkMode()

	if !model.MarkMode() {
		t.Error("Failed to enter marks state")
	}
	sendNumberEvent(model, 1)
	sendNumberEvent(model, 2)

	if model.Selected().Number() != 0 {
		t.Error("InputNumber in mark mode set the number", model.Selected())
	}

	if !model.Selected().Mark(1) {
		t.Error("InputNumber in mark mode didn't set the first mark", model.Selected())
	}

	if !model.Selected().Mark(2) {
		t.Error("InputNumber in mark mode didn't set the second mark", model.Selected())
	}

	model.MoveSelectionRight(false)
	if !model.MarkMode() {
		t.Error("Moving selection right DID exit mark mode.")
	}

	//Make sure that enter mark mode doesn't happen if the cell is locked or filled.

	model.MoveSelectionRight(false)
	model.Selected().Lock()

	model.ToggleSelectedMark(1)

	if model.Selected().Mark(1) {
		t.Error("Toggled mark on locked cell")
	}

	if model.consoleMessage != MARKS_MODE_FAIL_LOCKED {
		t.Error("Trying to mark locked cell didn't console message")
	}

	model.Selected().Unlock()
	model.Selected().SetNumber(1)
	model.ToggleSelectedMark(1)

	if model.Selected().Mark(1) {
		t.Error("Toggled mark on filled cell")
	}

	if model.consoleMessage != MARKS_MODE_FAIL_NUMBER {
		t.Error("Trying to mark filled cell didn't console message")
	}

}

func TestCommandMode(t *testing.T) {
	model := newController()
	//Add empty grid.
	model.SetGrid(sudoku.NewGrid())

	model.EnterMode(MODE_COMMAND)

	if model.mode != MODE_COMMAND {
		t.Error("Trying to enter state command failed")
	}

	if model.StatusLine() != STATUS_COMMAND {
		t.Error("got wrong status line for command state", model.StatusLine())
	}

	sendCharEvent(model, 'q')

	if model.mode != MODE_CONFIRM {
		t.Error("'q' in command mode didn't got to confirm state")
	}

	sendCharEvent(model, 'y')

	if !model.exitNow {
		t.Error("In command state, 'q' confirmed with 'y' didn't tell us to quit")
	}

	model.exitNow = false

	model.EnterMode(MODE_COMMAND)

	gridBefore := model.Grid()

	sendCharEvent(model, 'n')

	if model.mode != MODE_CONFIRM {
		t.Error("'n' didn't go to confirm state")
	}

	//confirmState has its own tests, so just pretend the user accepted.
	sendCharEvent(model, 'y')

	if model.Grid() == gridBefore {
		t.Error("'n' in command status didn't create new grid.")
	}

	if model.mode != MODE_DEFAULT {
		t.Error("'n' in command mode didn't go back to default mode when it was done")
	}

	model.EnterMode(MODE_COMMAND)

	sendKeyEvent(model, termbox.KeyEsc)
	if model.mode != MODE_DEFAULT {
		t.Error("'Esc' in command state didn't go back to default mode")

	}
}

func TestConfirmMode(t *testing.T) {
	model := newController()

	channel := make(chan bool, 1)

	model.enterConfirmMode("TEST", DEFAULT_YES,
		func() {
			channel <- true
			model.EnterMode(MODE_DEFAULT)
		},
		func() {
			channel <- false
			model.EnterMode(MODE_DEFAULT)
		},
	)

	if model.mode != MODE_CONFIRM {
		t.Error("enterConfirmState didn't lead to being in confirm state")
	}

	if model.StatusLine() != "TEST  {Y}/{n}" {
		t.Error("Default yes confirm state had wrong status line:", model.StatusLine())
	}

	sendCharEvent(model, 'y')

	select {
	case val := <-channel:
		if !val {
			t.Error("Got a value for confirm, but it was from the no action")
		}
		//If val is true, that's good.
	default:
		t.Error("No action filed after confirm state decided.")

	}

	if model.mode != MODE_DEFAULT {
		t.Error("After confirm accepted, not ending up in default mode.")
	}

	//TODO: test behavior with different DEFAULT_* , and with enter, y, n.
}

func TestLoadMode(t *testing.T) {
	c := newController()

	sendCharEvent(c, 'c')
	sendCharEvent(c, 'l')

	if c.mode != MODE_FILE_INPUT {
		t.Error("c then l didn't get us into load mode")
	}

	if MODE_FILE_INPUT.cursorOffset != 0 {
		t.Error("Cursor offset wasn't 0 to start")
	}

	sendCharEvent(c, 'm')

	if MODE_FILE_INPUT.cursorOffset != 1 {
		t.Error("Typing a character didn't move the cursor")
	}

	if MODE_FILE_INPUT.input != "m" {
		t.Error("Typing m in load mode didn't type m in the input")
	}

	cursorX := c.mode.cursorLocation(c)
	if cursorX != len(STATUS_LOAD)+len(MODE_FILE_INPUT.input) {
		t.Error("Cursor in wrong location, should be at end of input")
	}

	sendKeyEvent(c, termbox.KeyArrowRight)

	if MODE_FILE_INPUT.cursorOffset != 1 {
		t.Error("Moving right at end of input moved off end")
	}

	sendKeyEvent(c, termbox.KeyArrowLeft)

	if MODE_FILE_INPUT.cursorOffset != 0 {
		t.Error("Moving left didn't move cursor back to beginning")
	}

	sendKeyEvent(c, termbox.KeyArrowLeft)

	if MODE_FILE_INPUT.cursorOffset != 0 {
		t.Error("Moving left at front of input went off end")
	}

	sendKeyEvent(c, termbox.KeyArrowRight)
	sendKeyEvent(c, termbox.KeyBackspace2)

	if MODE_FILE_INPUT.cursorOffset != 0 {
		t.Error("Backspace at end of input didn't remove it")
	}

	if MODE_FILE_INPUT.input != "" {
		t.Error("Backspace at end of input didn't make it zero length")
	}

	sendKeyEvent(c, termbox.KeyEsc)

	if c.mode != MODE_DEFAULT {
		t.Error("Esc in load mode didn't go back to default.")
	}

	sendCharEvent(c, 'c')
	sendCharEvent(c, 'l')

	if MODE_FILE_INPUT.input != "" {
		t.Error("Going back into load mode didn't reset the input")
	}

	if MODE_FILE_INPUT.cursorOffset != 0 {
		t.Error("Cursor offset wasn't reset back to 0")
	}

	sendCharEvent(c, 'a')
	sendCharEvent(c, 'b')
	sendCharEvent(c, 'c')

	sendKeyEvent(c, termbox.KeyArrowLeft)
	sendCharEvent(c, 'd')

	if MODE_FILE_INPUT.input != "abdc" {
		t.Error("Adding a character in middle of string came out wrong:", MODE_FILE_INPUT.input)
	}

	if MODE_FILE_INPUT.cursorOffset != 3 {
		t.Error("After adding a character in middle of string, cursor was in wrong loc", MODE_FILE_INPUT.cursorOffset)
	}

	sendKeyEvent(c, termbox.KeyBackspace2)

	if MODE_FILE_INPUT.input != "abc" {
		t.Error("Backspace in middle had wrong result: ", MODE_FILE_INPUT.input)
	}

	if MODE_FILE_INPUT.cursorOffset != 2 {
		t.Error("Backspace in middle had wrong result: ", MODE_FILE_INPUT.cursorOffset)
	}

	//move cursor to end
	sendKeyEvent(c, termbox.KeyArrowRight)

	//Try to delete FOUR characters
	sendKeyEvent(c, termbox.KeyBackspace2)
	sendKeyEvent(c, termbox.KeyBackspace2)
	sendKeyEvent(c, termbox.KeyBackspace2)
	sendKeyEvent(c, termbox.KeyBackspace2)

	if MODE_FILE_INPUT.input != "" {
		t.Error("Removing all input still left some")
	}

	currentGrid := c.Grid().DataString()

	//Try loading for real

	//Try an invalid file name

	for _, ch := range "INVALIDPUZZLEFILE" {
		sendCharEvent(c, ch)
	}
	sendKeyEvent(c, termbox.KeyEnter)
	if c.consoleMessage == "" {
		t.Error("Trying to load invalid puzzle didn't show error message")
	}
	if c.Grid().DataString() != currentGrid {
		t.Error("Trying to load an invalid puzzle mutated grid")
	}
	if c.mode != MODE_DEFAULT {
		t.Error("Trying to load invalid puzzle didn't go back to default mode")
	}

	sendCharEvent(c, 'c')
	sendCharEvent(c, 'l')

	//Try an invalid puzzle
	for _, ch := range "test_puzzles/invalid_sdk_too_short.sdk" {
		sendCharEvent(c, ch)
	}
	sendKeyEvent(c, termbox.KeyEnter)
	if c.consoleMessage == "" {
		t.Error("Trying to load invalid puzzle didn't show error message")
	}
	if c.Grid().DataString() != currentGrid {
		t.Error("Trying to load an invalid puzzle mutated grid")
	}
	if c.mode != MODE_DEFAULT {
		t.Error("Trying to load invalid puzzle didn't go back to default mode")
	}

	//Try a good puzzle

	sendCharEvent(c, 'c')
	sendCharEvent(c, 'l')

	for _, ch := range "test_puzzles/converter_one.sdk" {
		sendCharEvent(c, ch)
	}
	sendKeyEvent(c, termbox.KeyEnter)
	if c.consoleMessage != GRID_LOADED_MESSAGE {
		t.Error("Trying to load valid puzzle didn't show load message:", c.consoleMessage)
	}
	if c.Grid().DataString() == currentGrid {
		t.Error("Trying to load a valid puzzle didn't mutate grid")
	}
	if c.mode != MODE_DEFAULT {
		t.Error("Trying to load valid puzzle didn't go back to default mode")
	}

	//Tab completion
	sendCharEvent(c, 'c')
	sendCharEvent(c, 'l')
	sendCharEvent(c, 't')
	sendKeyEvent(c, termbox.KeyTab)

	if MODE_FILE_INPUT.input != "test_puzzles/" {
		t.Error("tab on 'p' didn't complete to puzzles")
	}

	sendKeyEvent(c, termbox.KeyTab)

	if MODE_FILE_INPUT.input != "test_puzzles/" {
		t.Error("Tab complete on a thing with no obvious fill did something")
	}

	if c.consoleMessage != "Possible completions\n{converter_copy.sdk}\n{converter_one.sdk}\n{invalid_sdk_too_short.sdk}" {
		t.Error("Got wrong console message on ambiguous tab:", c.consoleMessage)
	}

	sendCharEvent(c, 'i')

	sendKeyEvent(c, termbox.KeyTab)

	if MODE_FILE_INPUT.input != "test_puzzles/invalid_sdk_too_short.sdk" {
		t.Error("Second valid autocomplete filled wrong thing")
	}

	if MODE_FILE_INPUT.cursorOffset != len(MODE_FILE_INPUT.input) {
		t.Error("Tab complete didn't move to end of input.")
	}

	sendKeyEvent(c, termbox.KeyEsc)

	//Try completion with a prefix
	sendCharEvent(c, 'c')
	sendCharEvent(c, 'l')
	sendCharEvent(c, 't')
	sendKeyEvent(c, termbox.KeyTab)
	//autocompleted to 'puzzles/'
	sendCharEvent(c, 'c')
	sendKeyEvent(c, termbox.KeyTab)

	if MODE_FILE_INPUT.input != "test_puzzles/converter_" {
		t.Error("Tab on prefix didn't fill out to end of longested common prefix")
	}

	//Test ctrl-a

	sendKeyEvent(c, termbox.KeyCtrlA)

	if MODE_FILE_INPUT.cursorOffset != 0 {
		t.Error("Ctrl-A didn't move cursor to front")
	}

	//Test ctrl-e

	sendKeyEvent(c, termbox.KeyCtrlE)

	if MODE_FILE_INPUT.cursorOffset != len(MODE_FILE_INPUT.input) {
		t.Error("Ctrl-E didn't move to end of line")
	}

	//Test ctrl-k

	currentInput := MODE_FILE_INPUT.input

	sendKeyEvent(c, termbox.KeyArrowLeft)
	sendKeyEvent(c, termbox.KeyArrowLeft)

	sendKeyEvent(c, termbox.KeyCtrlK)

	expectedInput := currentInput[:len(currentInput)-2]

	if MODE_FILE_INPUT.input != expectedInput {
		t.Error("Ctrl-K in middle didn't remove expected characters")
	}

}

func TestLongestCommonPrefix(t *testing.T) {
	//This is actually kind of a dumb way to encode test cases, because we
	//can't have two tests with the same expected result.
	testCases := map[string][]string{
		"123":  {"123", "1234"},
		"":     {},
		"abc":  {"abc", "abc"},
		"abcd": {"abcd"},
	}

	for expected, testSet := range testCases {
		result := longestCommonPrefix(testSet)
		if result != expected {
			t.Error("Got wrong result, got", result, "expected", expected, "for", testSet)
		}
	}
}

//TODO: test fast move mode
