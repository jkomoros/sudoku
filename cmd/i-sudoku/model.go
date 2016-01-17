package main

import (
	"github.com/jkomoros/sudoku"
	"github.com/mitchellh/go-wordwrap"
	"github.com/nsf/termbox-go"
)

type mainController struct {
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

func newController() *mainController {
	model := &mainController{
		state: STATE_DEFAULT,
	}
	model.setUpToggles()
	model.EnsureSelected()
	return model
}

func (c *mainController) setUpToggles() {

	//State variable for the closure
	var fastMode bool
	var markMode bool

	c.toggles = []toggle{
		//Solved
		{
			func() bool {
				return c.grid.Solved()
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
				return c.grid.Invalid()
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
func (c *mainController) EnterState(state InputState) {
	//SetState doesn't do much, it just makes it feel less weird than
	//STATE.enter(m) (which feels backward)

	if state.shouldEnter(c) {
		c.state = state
	}
}

//enterConfirmState is a special state to set
func (c *mainController) enterConfirmState(msg string, defaultAction defaultOption, yesAction func(), noAction func()) {
	STATE_CONFIRM.msg = msg
	STATE_CONFIRM.defaultAction = defaultAction
	STATE_CONFIRM.yesAction = yesAction
	STATE_CONFIRM.noAction = noAction
	c.EnterState(STATE_CONFIRM)
}

func (c *mainController) SetConsoleMessage(msg string, shortLived bool) {

	if c.outputWidth != 0 {
		//Wrap to fit in given size
		msg = wordwrap.WrapString(msg, uint(c.outputWidth))
	}

	c.consoleMessage = msg
	c.consoleMessageShort = shortLived
	c.lastShownHint = nil
}

//WillProcessEvent is cleared right before we call handleInput on the current
//state--that is, right before we process an event. That is a convenient time
//to clear state and prepare for the next state. This is *not* called before a
//timer/display tick.
func (c *mainController) WillProcessEvent() {
	if c.consoleMessageShort {
		c.ClearConsole()
	}
}

func (c *mainController) ClearConsole() {
	c.consoleMessage = ""
	c.consoleMessageShort = false
	c.lastShownHint = nil
}

func (c *mainController) StatusLine() string {
	return c.state.statusLine(c)
}

func (c *mainController) Selected() *sudoku.Cell {
	return c.selected
}

func (c *mainController) SetSelected(cell *sudoku.Cell) {
	if cell == c.selected {
		//Already done
		return
	}
	c.selected = cell
	c.state.newCellSelected(c)
}

func (c *mainController) EnsureSelected() {
	c.EnsureGrid()
	//Ensures that at least one cell is selected.
	if c.Selected() == nil {
		c.SetSelected(c.grid.Cell(0, 0))
	}
}

func (c *mainController) MoveSelectionLeft(fast bool) {
	c.EnsureSelected()
	row := c.Selected().Row()
	col := c.Selected().Col()
	for {
		col--
		if col < 0 {
			col = 0
		}
		if fast && c.grid.Cell(row, col).Number() != 0 {
			if col == 0 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				c.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		c.SetSelected(c.grid.Cell(row, col))
		break
	}
}

func (c *mainController) MoveSelectionRight(fast bool) {
	c.EnsureSelected()
	row := c.Selected().Row()
	col := c.Selected().Col()
	for {
		col++
		if col >= sudoku.DIM {
			col = sudoku.DIM - 1
		}
		if fast && c.grid.Cell(row, col).Number() != 0 {
			if col == sudoku.DIM-1 {
				//We're at the end and didn't find anything.
				//guess there's nothing to find.
				c.SetConsoleMessage(FAST_MODE_NO_OPEN_CELLS, true)
				return
			}
			continue
		}
		c.SetSelected(c.grid.Cell(row, col))
		break
	}
}

func (m *mainController) MoveSelectionUp(fast bool) {
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

func (m *mainController) MoveSelectionDown(fast bool) {
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

func (m *mainController) FastMode() bool {
	return m.toggles[TOGGLE_FAST_MODE].Value()
}

func (m *mainController) ToggleFastMode() {
	m.toggles[TOGGLE_FAST_MODE].Toggle()
}

func (m *mainController) MarkMode() bool {
	return m.toggles[TOGGLE_MARK_MODE].Value()
}

func (m *mainController) ToggleMarkMode() {
	m.toggles[TOGGLE_MARK_MODE].Toggle()
}

func (m *mainController) EnsureGrid() {
	if m.grid == nil {
		m.NewGrid()
	}
}

func (c *mainController) NewGrid() {
	oldCell := c.Selected()

	c.grid = sudoku.GenerateGrid(nil)
	//The currently selected cell is tied to the grid, so we need to fix it up.
	if oldCell != nil {
		c.SetSelected(oldCell.InGrid(c.grid))
	}
	c.grid.LockFilledCells()
}

func (c *mainController) ResetGrid() {
	c.grid.ResetUnlockedCells()
}

//If the selected cell has only one mark, fill it.
func (c *mainController) SetSelectedToOnlyMark() {
	c.EnsureSelected()
	marks := c.Selected().Marks()
	if len(marks) != 1 {
		c.SetConsoleMessage(SINGLE_FILL_MORE_THAN_ONE_MARK, true)
		return
	}
	//Rely on SetSelectedNumber to barf if it's not allowed for some other reason.
	c.SetSelectedNumber(marks[0])
}

func (c *mainController) SetSelectedNumber(num int) {
	c.EnsureSelected()
	if c.Selected().Locked() {
		c.SetConsoleMessage(DEFAULT_MODE_FAIL_LOCKED, true)
		return
	}

	if c.Selected().Number() != num {
		c.Selected().SetNumber(num)
	} else {
		//If the number to set is already set, then empty the cell instead.
		c.Selected().SetNumber(0)
	}

	c.checkHintDone()
}

func (c *mainController) checkHintDone() {
	if c.lastShownHint == nil {
		return
	}
	lastStep := c.lastShownHint.Steps[len(c.lastShownHint.Steps)-1]
	num := lastStep.TargetNums[0]
	cell := lastStep.TargetCells[0]
	if cell.InGrid(c.grid).Number() == num {
		c.ClearConsole()
	}
}

func (c *mainController) ToggleSelectedMark(num int) {
	c.EnsureSelected()
	if c.Selected().Locked() {
		c.SetConsoleMessage(MARKS_MODE_FAIL_LOCKED, true)
		return
	}
	if c.Selected().Number() != 0 {
		c.SetConsoleMessage(MARKS_MODE_FAIL_NUMBER, true)
		return
	}
	c.Selected().SetMark(num, !c.Selected().Mark(num))
}

func (c *mainController) FillSelectedWithLegalMarks() {
	c.EnsureSelected()
	c.Selected().ResetMarks()
	for _, num := range c.Selected().Possibilities() {
		c.Selected().SetMark(num, true)
	}
}

func (c *mainController) RemoveInvalidMarksFromSelected() {
	c.EnsureSelected()
	for _, num := range c.Selected().Marks() {
		if !c.Selected().Possible(num) {
			c.Selected().SetMark(num, false)
		}
	}
}
