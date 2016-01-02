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

func TestMode(t *testing.T) {
	model := newModel()

	//TODO: refactor these tests, make them be oriented around state structs.

	//Add empty grid.
	model.grid = sudoku.NewGrid()
	model.SetSelected(nil)

	if model.StatusLine() != STATUS_DEFAULT {
		t.Error("Didn't get default status line in default mode.")
	}

	sendNumberEvent(model, 1)

	if model.Selected().Number() != 1 {
		t.Error("InputNumber in default mode didn't add a number")
	}

	model.MoveSelectionRight()

	STATE_ENTER_MARKS.enter(model)
	if model.StatusLine() != STATUS_MARKING+"[]"+STATUS_MARKING_POSTFIX {
		t.Error("In mark mode with no marks, didn't get expected", model.StatusLine())
	}
	sendNumberEvent(model, 1)
	sendNumberEvent(model, 2)
	if model.StatusLine() != STATUS_MARKING+"[1 2]"+STATUS_MARKING_POSTFIX {
		t.Error("In makr mode with two marks, didn't get expected", model.StatusLine())
	}
	STATE_ENTER_MARKS.commitMarks(model)
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

	STATE_ENTER_MARKS.enter(model)
	sendNumberEvent(model, 1)
	sendNumberEvent(model, 2)
	sendKeyEvent(model, termbox.KeyEsc)

	if model.StatusLine() != STATUS_DEFAULT {
		t.Error("After canceling mark mode, status didn't go back to default.", model.StatusLine())
	}

	if model.Selected().Mark(1) || model.Selected().Mark(2) {
		t.Error("InputNumber in canceled mark mode still set marks")
	}

	model.MoveSelectionRight()

	sendNumberEvent(model, 1)

	if model.Selected().Number() != 1 {
		t.Error("InputNumber after cancled mark and another InputNum didn't set num", model.Selected())
	}

	if !sendKeyEvent(model, termbox.KeyEsc) {
		t.Error("ModeInputEsc not in mark enter mode didn't tell us to quit.")
	}

	model.MoveSelectionRight()

	STATE_ENTER_MARKS.enter(model)
	if sendKeyEvent(model, termbox.KeyEsc) {
		t.Error("ModeInputEsc in mark enter mode DID tell us to quit")
	}
	if model.state == STATE_ENTER_MARKS {
		t.Error("ModeInputEsc in mark enter mode didn't exit mark enter mode")
	}

	STATE_ENTER_MARKS.enter(model)
	model.MoveSelectionRight()
	if model.state == STATE_ENTER_MARKS {
		t.Error("Moving selection right didn't exit mark mode.")
	}

}

func TestNoMarkModeWhenLocked(t *testing.T) {
	model := newModel()
	model.grid = sudoku.NewGrid()
	model.SetSelected(nil)
	model.EnsureSelected()

	model.Selected().Lock()
	STATE_ENTER_MARKS.enter(model)

	if model.state == STATE_ENTER_MARKS {
		t.Error("Were allowed to enter mark mode even though cell was locked.")
	}

	model.Selected().Unlock()
	model.Selected().SetNumber(1)
	STATE_ENTER_MARKS.enter(model)

	if model.state == STATE_ENTER_MARKS {
		t.Error("We were allowed to enter mark mode even though cell had a number in it.")
	}

}
