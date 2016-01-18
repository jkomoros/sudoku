package main

import (
	"github.com/jkomoros/sudoku"
	"github.com/nsf/termbox-go"
	"strconv"
	"testing"
	"unicode/utf8"
)

func sendKeyEvent(m *mainController, k termbox.Key) {
	evt := termbox.Event{
		Type: termbox.EventKey,
		Key:  k,
	}
	m.mode.handleInput(m, evt)
}

func sendNumberEvent(m *mainController, num int) {
	ch, _ := utf8.DecodeRuneInString(strconv.Itoa(num))
	evt := termbox.Event{
		Type: termbox.EventKey,
		Ch:   ch,
	}
	m.mode.handleInput(m, evt)
}

//TODO: use sendCharEvent to verify that chars in all states do what they should.
func sendCharEvent(m *mainController, ch rune) {
	evt := termbox.Event{
		Type: termbox.EventKey,
		Ch:   ch,
	}
	m.mode.handleInput(m, evt)
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

	model.enterConfirmMode("TEST", DEFAULT_YES, func() { channel <- true }, func() { channel <- false })

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

//TODO: test fast move mode
