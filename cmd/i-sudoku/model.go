package main

import (
	"github.com/jkomoros/sudoku"
)

type mainModel struct {
	grid     *sudoku.Grid
	selected *sudoku.Cell
	state    InputState
}

func newModel() *mainModel {
	model := &mainModel{
		state: STATE_DEFAULT,
	}
	model.EnsureSelected()
	return model
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

func (m *mainModel) MoveSelectionLeft() {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	c--
	if c < 0 {
		c = 0
	}
	m.SetSelected(m.grid.Cell(r, c))
}

func (m *mainModel) MoveSelectionRight() {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	c++
	if c >= sudoku.DIM {
		c = sudoku.DIM - 1
	}
	m.SetSelected(m.grid.Cell(r, c))
}

func (m *mainModel) MoveSelectionUp() {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	r--
	if r < 0 {
		r = 0
	}
	m.SetSelected(m.grid.Cell(r, c))
}

func (m *mainModel) MoveSelectionDown() {
	m.EnsureSelected()
	r := m.Selected().Row()
	c := m.Selected().Col()
	r++
	if r >= sudoku.DIM {
		r = sudoku.DIM - 1
	}
	m.SetSelected(m.grid.Cell(r, c))
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
		return
	}

	if m.Selected().Number() != num {
		m.Selected().SetNumber(num)
	} else {
		//If the number to set is already set, then empty the cell instead.
		m.Selected().SetNumber(0)
	}
}

func (m *mainModel) ToggleSelectedMark(num int) {
	m.EnsureSelected()
	if m.Selected().Locked() {
		return
	}
	m.Selected().SetMark(num, !m.Selected().Mark(num))
}
