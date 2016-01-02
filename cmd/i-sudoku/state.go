package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"strconv"
	"strings"
)

var (
	STATE_DEFAULT     = &defaultState{}
	STATE_ENTER_MARKS = &enterMarkState{}
	STATE_COMMAND     = &commandState{}
	STATE_CONFIRM     = &confirmState{}
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
		//TODO: 'h' should give a hint
		//TODO: '?' should print help to console
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
		//TODO: '+' should set marks to add all Possible values that are not currently added
		//TODO: '-' should set marks list to remove any things that are not possible.
		case 'q':
			//TODO: this should use a confirmState, too.
			//...Although it's hard to do this given that we have to return whether or not to quit right now.
			return true
		case 'n':
			m.enterConfirmState("Replace grid with a new one? This is a destructive action.",
				DEFAULT_NO,
				func() {
					m.NewGrid()
				},
				func() {},
			)
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

type defaultOption int

const (
	DEFAULT_YES defaultOption = iota
	DEFAULT_NO
	DEFAULT_NONE
)

type confirmState struct {
	msg           string
	defaultAction defaultOption
	yesAction     func()
	noAction      func()
	baseState
}

func (s *confirmState) handleInput(m *mainModel, evt termbox.Event) (doQuit bool) {
	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyEnter:
			switch s.defaultAction {
			case DEFAULT_YES:
				s.yesAction()
				m.EnterState(STATE_DEFAULT)
			case DEFAULT_NO:
				s.noAction()
				m.EnterState(STATE_DEFAULT)
			case DEFAULT_NONE:
				//Don't do anything
			}
		default:
			handled = false
		}
		switch evt.Ch {
		case 'y':
			s.yesAction()
			m.EnterState(STATE_DEFAULT)
		case 'n':
			s.noAction()
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

func (s *confirmState) statusLine(m *mainModel) string {
	confirmMsg := "(y) or (n)"
	if s.defaultAction == DEFAULT_YES {
		confirmMsg = "(Y) or (n)"
	} else if s.defaultAction == DEFAULT_NO {
		confirmMsg = "(y) or (N)"
	}
	return s.msg + "  " + confirmMsg
}
