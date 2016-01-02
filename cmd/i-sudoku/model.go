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
	m.grid = sudoku.GenerateGrid(nil)
	m.grid.LockFilledCells()
}

func (m *mainModel) SetSelectedNumber(num int) {
	//TODO: if the number already has that number set, set 0.
	m.EnsureSelected()
	if m.Selected().Locked() {
		return
	}
	m.Selected().SetNumber(num)
}

func (m *mainModel) ToggleSelectedMark(num int) {
	m.EnsureSelected()
	if m.Selected().Locked() {
		return
	}
	m.Selected().SetMark(num, !m.Selected().Mark(num))
}