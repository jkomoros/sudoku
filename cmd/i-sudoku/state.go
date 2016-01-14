package main

import (
	"github.com/nsf/termbox-go"
	"strconv"
	"strings"
)

var (
	STATE_DEFAULT = &defaultState{}
	STATE_COMMAND = &commandState{}
	STATE_CONFIRM = &confirmState{}
)

const (
	MARKS_MODE_FAIL_LOCKED         = "Can't enter mark mode on a cell that's locked."
	MARKS_MODE_FAIL_NUMBER         = "Can't enter mark mode on a cell that has a filled number."
	DEFAULT_MODE_FAIL_LOCKED       = "Can't enter a number in a locked cell."
	FAST_MODE_NO_OPEN_CELLS        = "Can't fast move: no more open cells in that direction"
	SINGLE_FILL_MORE_THAN_ONE_MARK = "The cell does not have precisely one mark set."
	HELP_MESSAGE                   = `The following commands are also available on this screen:
{c} to enter command mode to do things like quit and load a new puzzle
{h} to get a hint
{+} or {=} to set the selected cell's marks to all legal marks
{-} to remove all invalid marks from the selected cell
{<enter>} to set a cell to the number that is the only current mark
{m} to enter mark mode, so all numbers entered will toggle marks
{f} to toggle fast move mode, allowing you to skip over filled cells`
	STATUS_DEFAULT = "{→,←,↓,↑} to move cells, {0-9} to enter number, {Shift + 0-9} to toggle marks, {?} to list other commands"
	STATUS_COMMAND = "COMMAND: {n}ew puzzle, {q}uit, {ESC} cancel"
)

func runeIsNum(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func runeIsShiftedNum(ch rune) bool {
	if runeIsNum(ch) {
		return false
	}
	return runeIsNum(shiftedNumRuneToNum(ch))
}

func shiftedNumRuneToNum(ch rune) rune {
	//Note: this assumes an american keyboard layout
	//TODO: make this resilient to other keyboard layouts, perhaps with a way
	//for the user to 'train' it.
	switch ch {
	case '!':
		return '1'
	case '@':
		return '2'
	case '#':
		return '3'
	case '$':
		return '4'
	case '%':
		return '5'
	case '^':
		return '6'
	case '&':
		return '7'
	case '*':
		return '8'
	case '(':
		return '9'
	case ')':
		return '0'
	}
	return ch
}

type InputState interface {
	//TODO: doesn't it feel weird that every method takes a main model?
	handleInput(m *mainModel, evt termbox.Event)
	shouldEnter(m *mainModel) bool
	statusLine(m *mainModel) string
	newCellSelected(m *mainModel)
}

type baseState struct{}

func (s *baseState) handleInput(m *mainModel, evt termbox.Event) {
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyCtrlC:
			confirmQuit(m)
		}
	}
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

func showHint(m *mainModel) {

	//TODO: shouldn't this be a method on model?  The rule of thumb is no
	//modifying state in model except in model methods.
	hint := m.grid.Hint(nil)

	if len(hint.Steps) == 0 {
		m.SetConsoleMessage("No hint to give.", true)
		return
	}
	m.SetConsoleMessage("{Hint}\n"+strings.Join(hint.Description(), "\n")+"\n\n"+"{ENTER} to accept, {ESC} to ignore", false)
	//This hast to be after setting console message, since SetConsoleMessage clears the last hint.
	m.lastShownHint = hint
	lastStep := hint.Steps[len(hint.Steps)-1]
	m.SetSelected(lastStep.TargetCells[0].InGrid(m.grid))
}

func (s *defaultState) enterHint(m *mainModel) {
	if m.lastShownHint == nil {
		return
	}
	lastStep := m.lastShownHint.Steps[len(m.lastShownHint.Steps)-1]
	cell := lastStep.TargetCells[0]
	num := lastStep.TargetNums[0]

	m.SetSelected(cell.InGrid(m.grid))
	m.SetSelectedNumber(num)

	m.ClearConsole()
}

func (s *defaultState) handleInput(m *mainModel, evt termbox.Event) {

	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyArrowDown:
			m.MoveSelectionDown(m.FastMode())
		case termbox.KeyArrowLeft:
			m.MoveSelectionLeft(m.FastMode())
		case termbox.KeyArrowRight:
			m.MoveSelectionRight(m.FastMode())
		case termbox.KeyArrowUp:
			m.MoveSelectionUp(m.FastMode())
		case termbox.KeyEsc:
			m.ClearConsole()
		case termbox.KeyEnter:
			if m.lastShownHint != nil {
				s.enterHint(m)
			} else {
				m.SetSelectedToOnlyMark()
			}
		default:
			handled = false
		}
		switch {
		case evt.Ch == 'h':
			showHint(m)
		case evt.Ch == 'f':
			m.ToggleFastMode()
		case evt.Ch == '?', evt.Ch == '/':
			m.SetConsoleMessage(HELP_MESSAGE, true)
		case evt.Ch == '+', evt.Ch == '=':
			m.FillSelectedWithLegalMarks()
		case evt.Ch == '-':
			m.RemoveInvalidMarksFromSelected()
		case evt.Ch == 'c':
			m.EnterState(STATE_COMMAND)
		case evt.Ch == 'm':
			m.ToggleMarkMode()
		case runeIsShiftedNum(evt.Ch):
			//TODO: ideally Ctrl+Num would work to put in one mark. But termbox doesn't appear to let that work.
			num, err := strconv.Atoi(strings.Replace(strconv.QuoteRuneToASCII(shiftedNumRuneToNum(evt.Ch)), "'", "", -1))
			if err != nil {
				panic(err)
			}
			m.ToggleSelectedMark(num)
		case runeIsNum(evt.Ch):
			//TODO: this is a seriously gross way of converting a rune to a string.
			num, err := strconv.Atoi(strings.Replace(strconv.QuoteRuneToASCII(evt.Ch), "'", "", -1))
			if err != nil {
				panic(err)
			}
			if m.MarkMode() {
				m.ToggleSelectedMark(num)
			} else {
				m.SetSelectedNumber(num)
			}
		default:
			if !handled {
				//neither handler handled it; defer to base.
				s.baseState.handleInput(m, evt)
			}
		}
	}
}

type commandState struct {
	baseState
}

func confirmQuit(m *mainModel) {
	m.enterConfirmState("Quit? Your progress will be lost.",
		DEFAULT_NO,
		func() {
			m.exitNow = true
		},
		func() {},
	)
}

func (s *commandState) handleInput(m *mainModel, evt termbox.Event) {
	handled := true
	switch evt.Type {
	case termbox.EventKey:
		switch evt.Key {
		case termbox.KeyEsc:
			m.EnterState(STATE_DEFAULT)
		default:
			handled = false
		}
		switch {
		case evt.Ch == 'q':
			confirmQuit(m)
		case evt.Ch == 'n':
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
				s.baseState.handleInput(m, evt)
			}
		}
	}
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

func (s *confirmState) handleInput(m *mainModel, evt termbox.Event) {
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
				s.baseState.handleInput(m, evt)
			}
		}
	}
}

func (s *confirmState) statusLine(m *mainModel) string {
	confirmMsg := "{y}/{n}"
	if s.defaultAction == DEFAULT_YES {
		confirmMsg = "{Y}/{n}"
	} else if s.defaultAction == DEFAULT_NO {
		confirmMsg = "{y}/{N}"
	}
	return s.msg + "  " + confirmMsg
}
