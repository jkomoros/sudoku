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

const (
	MARKS_MODE_FAIL_LOCKED = "Can't enter mark mode on a cell that's locked."
	MARKS_MODE_FAIL_NUMBER = "Can't enter mark mode on a cell that has a filled number."
)

func runeIsNum(ch rune) bool {
	return ch >= '0' && ch <= '9'
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
	hint := m.grid.Hint(nil)
	m.SetConsoleMessage(strings.Join(hint.Description(), "\n")+"\n\n"+"To accept this hint, type {ENTER}\nTo clear this message, type {ESC}", false)
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
			m.MoveSelectionDown()
		case termbox.KeyArrowLeft:
			m.MoveSelectionLeft()
		case termbox.KeyArrowRight:
			m.MoveSelectionRight()
		case termbox.KeyArrowUp:
			m.MoveSelectionUp()
		case termbox.KeyEsc:
			m.ClearConsole()
		case termbox.KeyEnter:
			s.enterHint(m)
		default:
			handled = false
		}
		switch {
		case evt.Ch == 'c':
			m.EnterState(STATE_COMMAND)
		case evt.Ch == 'm':
			//TODO: ideally Ctrl+Num would work to put in one mark. But termbox doesn't appear to let that work.
			m.EnterState(STATE_ENTER_MARKS)
		case runeIsNum(evt.Ch):
			//TODO: this is a seriously gross way of converting a rune to a string.
			num, err := strconv.Atoi(strings.Replace(strconv.QuoteRuneToASCII(evt.Ch), "'", "", -1))
			if err != nil {
				panic(err)
			}
			m.SetSelectedNumber(num)
		default:
			if !handled {
				//neither handler handled it; defer to base.
				s.baseState.handleInput(m, evt)
			}
		}
	}
}

type enterMarkState struct {
	baseState
	marksToInput []int
}

func (s *enterMarkState) handleInput(m *mainModel, evt termbox.Event) {
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
		switch {
		case runeIsNum(evt.Ch):
			//TODO: this is a seriously gross way of converting a rune to a string.
			num, err := strconv.Atoi(strings.Replace(strconv.QuoteRuneToASCII(evt.Ch), "'", "", -1))
			if err != nil {
				panic(err)
			}
			s.numberInput(num)
		default:
			if !handled {
				//Neither of us handled it so defer to base.
				s.baseState.handleInput(m, evt)
			}
		}
	}
}

func (s *enterMarkState) numberInput(num int) {
	s.marksToInput = append(s.marksToInput, num)

	s.marksToInput = cleanMarkList(s.marksToInput)
}

func cleanMarkList(nums []int) []int {
	//Now, go through and remove duplicates
	numCount := make(map[int]int)
	for _, num := range nums {
		numCount[num] += 1
	}
	//Now we'll reconstruct the slice. For each num, if it was see an odd
	//number of times, we'll include it--but only the first time (so we'll
	//keep track of if we've output the number yet in numsIncluded).
	numsIncluded := make(map[int]bool)

	var result []int
	for _, num := range nums {
		if numCount[num]%2 == 1 {
			//It's odd, so output it if we haven't already.
			if !numsIncluded[num] {
				//We haven't output it yet, so output it.
				result = append(result, num)
				//Keep track of that we already output it.s
				numsIncluded[num] = true
			}
		}
	}

	if result == nil {
		result = []int{}
	}
	return result
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
		if selected.Locked() {
			m.SetConsoleMessage(MARKS_MODE_FAIL_LOCKED, true)
			return false
		}
		if selected.Number() != 0 {
			m.SetConsoleMessage(MARKS_MODE_FAIL_NUMBER, true)
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
		case evt.Ch == 'h':
			showHint(m)
			m.EnterState(STATE_DEFAULT)
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
