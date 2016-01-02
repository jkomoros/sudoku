package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"strconv"
	"strings"
)

var (
	//TODO: create a confirmationState, which takes a default (Y, N) and a state object whose commit it should call if accepted.
	STATE_DEFAULT     = &defaultState{}
	STATE_ENTER_MARKS = &enterMarkState{}
	STATE_COMMAND     = &commandState{}
)

type InputState interface {
	//TODO: doesn't it feel weird that every method takes a main model?
	handleInput(m *mainModel, evt termbox.Event) (doQuit bool)
	shouldEnter(m *mainModel) bool
	statusLine(m *mainModel) string
	newCellSelected(m *mainModel)
}

type baseState struct{}

func (s *baseState) handleInput(m *mainModel, evt termbox.Event) (doQuit bool) {
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyCtrlC:
			return true
		}
	}
	return false
}

func (s *baseState) statusLine(m *mainModel) string {
	//TODO: in StatusLine, the keyboard shortcuts should be in bold.
	//Perhaps make it so at open parens set to bold, at close parens set
	//to normal.

	return STATUS_DEFAULT

}

func (s *baseState) newCellSelected(m *mainModel) {
	//Do nothing by default.
}

func (s *baseState) shouldEnter(m *mainModel) bool {
	return true
}

type defaultState struct {
	baseState
}

func (s *defaultState) handleInput(m *mainModel, evt termbox.Event) (doQuit bool) {

	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyArrowDown:
			m.MoveSelectionDown()
		case termbox.KeyArrowLeft:
			m.MoveSelectionLeft()
		case termbox.KeyArrowRight:
			m.MoveSelectionRight()
		case termbox.KeyArrowUp:
			m.MoveSelectionUp()
		default:
			handled = false
		}
		switch evt.Ch {
		case 'c':
			m.EnterState(STATE_COMMAND)
		case 'm':
			//TODO: ideally Ctrl+Num would work to put in one mark. But termbox doesn't appear to let that work.
			m.EnterState(STATE_ENTER_MARKS)
		//TODO: do this in a more general way related to DIM
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			//TODO: this is a seriously gross way of converting a rune to a string.
			num, err := strconv.Atoi(strings.Replace(strconv.QuoteRuneToASCII(evt.Ch), "'", "", -1))
			if err != nil {
				panic(err)
			}
			m.SetSelectedNumber(num)
		default:
			if !handled {
				//neither handler handled it; defer to base.
				return s.baseState.handleInput(m, evt)
			}
		}
	}
	return false
}

type enterMarkState struct {
	baseState
	marksToInput []int
}

func (s *enterMarkState) handleInput(m *mainModel, evt termbox.Event) (doQuit bool) {
	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyEnter:
			s.commitMarks(m)
		case termbox.KeyEsc:
			m.EnterState(STATE_DEFAULT)
		default:
			handled = false
		}
		switch evt.Ch {
		//TODO: do this in a more general way related to DIM
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			//TODO: this is a seriously gross way of converting a rune to a string.
			num, err := strconv.Atoi(strings.Replace(strconv.QuoteRuneToASCII(evt.Ch), "'", "", -1))
			if err != nil {
				panic(err)
			}
			s.numberInput(num)
		default:
			if !handled {
				//Neither of us handled it so defer to base.
				return s.baseState.handleInput(m, evt)
			}
		}
	}
	return false
}

func (s *enterMarkState) numberInput(num int) {
	s.marksToInput = append(s.marksToInput, num)
}

func (s *enterMarkState) commitMarks(m *mainModel) {
	for _, num := range s.marksToInput {
		m.ToggleSelectedMark(num)
	}
	s.marksToInput = nil
	m.EnterState(STATE_DEFAULT)
}

func (s *enterMarkState) shouldEnter(m *mainModel) bool {
	selected := m.Selected()
	if selected != nil {
		if selected.Number() != 0 || selected.Locked() {
			//Dion't enter mark mode.
			return false
		}
	}
	s.marksToInput = make([]int, 0)
	return true
}

func (s *enterMarkState) statusLine(m *mainModel) string {
	return STATUS_MARKING + fmt.Sprint(s.marksToInput) + STATUS_MARKING_POSTFIX
}

func (s *enterMarkState) newCellSelected(m *mainModel) {
	m.EnterState(STATE_DEFAULT)
}

type commandState struct {
	baseState
}

func (s *commandState) handleInput(m *mainModel, evt termbox.Event) (doQuit bool) {
	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyEsc:
			m.EnterState(STATE_DEFAULT)
		default:
			handled = false
		}
		switch evt.Ch {
		case 'q':
			return true
		case 'n':
			//TODO: since this is a destructive action, require a confirmation
			m.NewGrid()
			m.EnterState(STATE_DEFAULT)
		default:
			if !handled {
				//Neither of us handled it so defer to base.
				return s.baseState.handleInput(m, evt)
			}
		}
	}
	return false
}

func (s *commandState) statusLine(m *mainModel) string {
	return STATUS_COMMAND
}
