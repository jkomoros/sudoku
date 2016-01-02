package main

import (
	"github.com/jkomoros/sudoku"
	"github.com/nsf/termbox-go"
	"strconv"
	"testing"
	"unicode/utf8"
)

func sendKeyEvent(m *mainModel, k termbox.Key) bool {
	evt := termbox.Event{
		Type: termbox.EventKey,
		Key:  k,
	}
	return m.state.handleInput(m, evt)
}

func sendNumberEvent(m *mainModel, num int) {
	ch, _ := utf8.DecodeRuneInString(strconv.Itoa(num))
	evt := termbox.Event{
		Type: termbox.EventKey,
		Ch:   ch,
	}
	m.state.handleInput(m, evt)
}

//TODO: use sendCharEvent to verify that chars in all states do what they should.
func sendCharEvent(m *mainModel, ch rune) bool {
	evt := termbox.Event{
		Type: termbox.EventKey,
		Ch:   ch,
	}
	return m.state.handleInput(m, evt)
}

func TestDefaultState(t *testing.T) {
	model := newModel()
	//Add empty grid.
	model.grid = sudoku.NewGrid()
	model.SetSelected(nil)

	if model.state != STATE_DEFAULT {
		t.Error("model didn't start in default state")
	}

	if STATE_DEFAULT.statusLine(model) != STATUS_DEFAULT {
		t.Error("Didn't get default status line in default mode.")
	}

	sendNumberEvent(model, 1)

	if model.Selected().Number() != 1 {
		t.Error("InputNumber in default mode didn't add a number")
	}

	if sendKeyEvent(model, termbox.KeyEsc) {
		t.Error("ModeInputEsc in DEFAULT_STATE did tell us to quit, but it shouldn't")
	}
}

func TestEnterMarksState(t *testing.T) {
	model := newModel()
	//Add empty grid.
	model.grid = sudoku.NewGrid()
	model.SetSelected(nil)
	model.EnsureSelected()

	model.EnterState(STATE_ENTER_MARKS)

	if model.state != STATE_ENTER_MARKS {
		t.Error("Failed to enter marks state")
	}
	if STATE_ENTER_MARKS.statusLine(model) != STATUS_MARKING+"[]"+STATUS_MARKING_POSTFIX {
		t.Error("In mark mode with no marks, didn't get expected", model.StatusLine())
	}
	sendNumberEvent(model, 1)
	sendNumberEvent(model, 2)
	if STATE_ENTER_MARKS.statusLine(model) != STATUS_MARKING+"[1 2]"+STATUS_MARKING_POSTFIX {
		t.Error("In makr mode with two marks, didn't get expected", model.StatusLine())
	}
	STATE_ENTER_MARKS.commitMarks(model)

	if model.state != STATE_DEFAULT {
		t.Error("Didn't go back to default state after commiting marks.")
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

	model.EnterState(STATE_ENTER_MARKS)
	sendNumberEvent(model, 1)
	sendNumberEvent(model, 2)
	if sendKeyEvent(model, termbox.KeyEsc) {
		t.Error("ModeInputEsc in mark enter state DID tell us to quit")
	}

	if model.state != STATE_DEFAULT {
		t.Error("Hitting esc in enter marks state didn't go back to esc")
	}

	if model.Selected().Mark(1) || model.Selected().Mark(2) {
		t.Error("InputNumber in canceled mark mode still set marks")
	}

	model.EnterState(STATE_ENTER_MARKS)
	model.MoveSelectionRight()
	if model.state == STATE_ENTER_MARKS {
		t.Error("Moving selection right didn't exit mark mode.")
	}

	//Make sure that enter mark mode doesn't happen if the cell is locked or filled.

	model.MoveSelectionRight()
	model.Selected().Lock()
	model.EnterState(STATE_ENTER_MARKS)

	if model.state == STATE_ENTER_MARKS {
		t.Error("Were allowed to enter mark mode even though cell was locked.")
	}

	model.Selected().Unlock()
	model.Selected().SetNumber(1)
	model.EnterState(STATE_ENTER_MARKS)

	if model.state == STATE_ENTER_MARKS {
		t.Error("We were allowed to enter mark mode even though cell had a number in it.")
	}

}

func TestCommandState(t *testing.T) {
	model := newModel()
	//Add empty grid.
	model.grid = sudoku.NewGrid()
	model.SetSelected(nil)
	model.EnsureSelected()

	model.EnterState(STATE_COMMAND)

	if model.state != STATE_COMMAND {
		t.Error("Trying to enter state command failed")
	}

	if model.StatusLine() != STATUS_COMMAND {
		t.Error("got wrong status line for command state", model.StatusLine())
	}

	if !sendCharEvent(model, 'q') {
		t.Error("In command state, 'q' didn't tell us to quit")
	}

	gridBefore := model.grid

	sendCharEvent(model, 'n')

	if model.state != STATE_CONFIRM {
		t.Error("'n' didn't go to confirm state")
	}

	//confirmState has its own tests, so just pretend the user accepted.
	sendCharEvent(model, 'y')

	if model.grid == gridBefore {
		t.Error("'n' in command status didn't create new grid.")
	}

	if model.state != STATE_DEFAULT {
		t.Error("'n' in command mode didn't go back to default mode when it was done")
	}

	model.EnterState(STATE_COMMAND)

	sendKeyEvent(model, termbox.KeyEsc)
	if model.state != STATE_DEFAULT {
		t.Error("'Esc' in command state didn't go back to default mode")

	}

}

func TestConfirmState(t *testing.T) {
	model := newModel()

	channel := make(chan bool, 1)

	model.enterConfirmState("TEST", DEFAULT_YES, func() { channel <- true }, func() { channel <- false })

	if model.state != STATE_CONFIRM {
		t.Error("enterConfirmState didn't lead to being in confirm state")
	}

	if model.StatusLine() != "TEST  (Y) or (n)" {
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

	if model.state != STATE_DEFAULT {
		t.Error("After confirm accepted, not ending up in default mode.")
	}

	//TODO: test behavior with different DEFAULT_* , and with enter, y, n.
}
