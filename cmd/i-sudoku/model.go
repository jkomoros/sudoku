package main

import (
	"github.com/jkomoros/sudoku"
	"github.com/mitchellh/go-wordwrap"
	"github.com/nsf/termbox-go"
)

type mainModel struct {
	grid     *sudoku.Grid
	selected *sudoku.Cell
	state    InputState
	//The size of the console output. Not used for much.
	outputWidth    int
	lastShownHint  *sudoku.SolveDirections
	consoleMessage string
	//if true, will zero out console message on turn of event loop.
	consoleMessageShort bool
	//If exitNow is flipped to true, we will quit at next turn of event loop.
	exitNow bool
	toggles []toggle
}

const (
	TOGGLE_SOLVED = iota
	TOGGLE_INVALID
	TOGGLE_MARK_MODE
	TOGGLE_FAST_MODE
)

type toggle struct {
	Value     func() bool
	Toggle    func()
	OnText    string
	OffText   string
	GridColor termbox.Attribute
}

func newModel() *mainModel {
	model := &mainModel{
		state: STATE_DEFAULT,
	}
	model.setUpToggles()
	model.EnsureSelected()
	return model
}

func (m *mainModel) setUpToggles() {

	//State variable for the closure
	var fastMode bool
	var markMode bool

	m.toggles = []toggle{
		//Solved
		{
			func() bool {
				return m.grid.Solved()
			},
			func() {
				//Do nothing; read only
			},
			"  SOLVED  ",
			" UNSOLVED ",
			termbox.ColorYellow,
		},
		//invalid
		{
			func() bool {
				return m.grid.Invalid()
			},
			func() {
				//Read only
			},
			" INVALID ",
			"  VALID  ",
			termbox.ColorRed,
		},
		//Mark mode
		{
			func() bool {
				return markMode
			},
			func() {
				markMode = !markMode
			},
			" MARKING ",
			"         ",
			termbox.ColorCyan,
		},
		//Fast mode
		{
			func() bool {
				return fastMode
			},
			func() {
				fastMode = !fastMode
			},
			"  FAST MODE  ",
			"             ",
			termbox.ColorBlue,
		},
	}
}

//EnterState attempts to set the model to the given state. The state object is
//given a chance to do initalization and potentially cancel the transition,
//leaving the model in the same state as before.
func (m *mainModel) EnterState(state InputState) {
	//SetState doesn't do much, it just makes it feel less weird than
	//STATE.enter(m) (which feels backward)

	if state.shouldEnter(m) {
		m.state = state
	}
}

//enterConfirmState is a special state to set
func (m *mainModel) enterConfirmState(msg string, defaultAction defaultOption, yesAction func(), noAction func()) {
	STATE_CONFIRM.msg = msg
	STATE_CONFIRM.defaultAction = defaultAction
	STATE_CONFIRM.yesAction = yesAction
	STATE_CONFIRM.noAction = noAction
	m.EnterState(STATE_CONFIRM)
}

func (m *mainModel) SetConsoleMessage(msg string, shortLived bool) {

	if m.outputWidth != 0 {
		//Wrap to fit in given size
		msg = wordwrap.WrapString(msg, uint(m.outputWidth))
	}

	m.consoleMessage = msg
	m.consoleMessageShort = shortLived
	m.lastShownHint = nil
}

func (m *mainModel) EndOfEventLoop() {
	if m.consoleMessageShort {
		m.ClearConsole()
	}
}

func (m *mainModel) ClearConsole() {
	m.consoleMessage = ""
	m.consoleMessageShort = false
	m.lastShownHint = nil
}

func (m *mainModel) StatusLine() string {
	return m.state.statusLine(m)
}

func (m *mainModel) Selected() *sudoku.Cell {
	return m.selected
}

func (m *mainModel) SetSelected(cell *sudoku.Cell) {
	if cell == m.selected {
		//Already done
		return
	}
	m.selected = cell
	m.state.newCellSelected(m)
}

func (m *mainModel) EnsureSelected() {
	m.EnsureGrid()
	//Ensures that at least one cell is selected.
	if m.Selected() == nil {
		m.SetSelected(m.grid.Cell(0, 0))
	}
}

func (m *mainModel) MoveSelectionLeft(fast bool) {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	for {
		c--
		if c < 0 {
			c = 0
		}
		if fast && m.grid.Cell(r, c).Number() != 0 {
			if c == 0 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				m.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		m.SetSelected(m.grid.Cell(r, c))
		break
	}
}

func (m *mainModel) MoveSelectionRight(fast bool) {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	for {
		c++
		if c >= sudoku.DIM {
			c = sudoku.DIM - 1
		}
		if fast && m.grid.Cell(r, c).Number() != 0 {
			if c == sudoku.DIM-1 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				m.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		m.SetSelected(m.grid.Cell(r, c))
		break
	}
}

func (m *mainModel) MoveSelectionUp(fast bool) {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	for {
		r--
		if r < 0 {
			r = 0
		}
		if fast && m.grid.Cell(r, c).Number() != 0 {
			if r == 0 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				m.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		m.SetSelected(m.grid.Cell(r, c))
		break
	}
}

func (m *mainModel) MoveSelectionDown(fast bool) {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	for {
		r++
		if r >= sudoku.DIM {
			r = sudoku.DIM - 1
		}
		if fast && m.grid.Cell(r, c).Number() != 0 {
			if r == sudoku.DIM-1 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				m.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		m.SetSelected(m.grid.Cell(r, c))
		break
	}
}

func (m *mainModel) FastMode() bool {
	return m.toggles[TOGGLE_FAST_MODE].Value()
}

func (m *mainModel) ToggleFastMode() {
	m.toggles[TOGGLE_FAST_MODE].Toggle()
}

func (m *mainModel) MarkMode() bool {
	return m.toggles[TOGGLE_MARK_MODE].Value()
}

func (m *mainModel) ToggleMarkMode() {
	m.toggles[TOGGLE_MARK_MODE].Toggle()
}

func (m *mainModel) EnsureGrid() {
	if m.grid == nil {
		m.NewGrid()
	}
}

func (m *mainModel) NewGrid() {
	oldCell := m.Selected()

	m.grid = sudoku.GenerateGrid(nil)
	//The currently selected cell is tied to the grid, so we need to fix it up.
	if oldCell != nil {
		m.SetSelected(oldCell.InGrid(m.grid))
	}
	m.grid.LockFilledCells()
}

func (m *mainModel) SetSelectedNumber(num int) {
	m.EnsureSelected()
	if m.Selected().Locked() {
		m.SetConsoleMessage(DEFAULT_MODE_FAIL_LOCKED, true)
		return
	}

	if m.Selected().Number() != num {
		m.Selected().SetNumber(num)
	} else {
		//If the number to set is already set, then empty the cell instead.
		m.Selected().SetNumber(0)
	}

	m.checkHintDone()
}

func (m *mainModel) checkHintDone() {
	if m.lastShownHint == nil {
		return
	}
	lastStep := m.lastShownHint.Steps[len(m.lastShownHint.Steps)-1]
	num := lastStep.TargetNums[0]
	cell := lastStep.TargetCells[0]
	if cell.InGrid(m.grid).Number() == num {
		m.ClearConsole()
	}
}

func (m *mainModel) ToggleSelectedMark(num int) {
	m.EnsureSelected()
	if m.Selected().Locked() {
		m.SetConsoleMessage(MARKS_MODE_FAIL_LOCKED, true)
		return
	}
	if m.Selected().Number() != 0 {
		m.SetConsoleMessage(MARKS_MODE_FAIL_NUMBER, true)
		return
	}
	m.Selected().SetMark(num, !m.Selected().Mark(num))
}

func (m *mainModel) FillSelectedWithLegalMarks() {
	m.EnsureSelected()
	m.Selected().ResetMarks()
	for _, num := range m.Selected().Possibilities() {
		m.Selected().SetMark(num, true)
	}
}

func (m *mainModel) RemoveInvalidMarksFromSelected() {
	m.EnsureSelected()
	for _, num := range m.Selected().Marks() {
		if !m.Selected().Possible(num) {
			m.Selected().SetMark(num, false)
		}
	}
}
