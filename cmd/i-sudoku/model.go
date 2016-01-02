package main

import (
	"fmt"
	"github.com/jkomoros/sudoku"
)

type mainModel struct {
	grid         *sudoku.Grid
	selected     *sudoku.Cell
	marksToInput []int
}

func newModel() *mainModel {
	model := &mainModel{}
	model.EnsureSelected()
	return model
}

func (m *mainModel) StatusLine() string {
	//TODO: return something dynamic depending on mode.

	//TODO: in StatusLine, the keyboard shortcuts should be in bold.
	//Perhaps make it so at open parens set to bold, at close parens set
	//to normal.

	if m.marksToInput == nil {
		return STATUS_DEFAULT
	} else {
		//Marks mode
		return STATUS_MARKING + fmt.Sprint(m.marksToInput) + STATUS_MARKING_POSTFIX
	}
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
	m.ModeCancelMarkMode()
}

func (m *mainModel) ModeInputEsc() (quit bool) {
	if m.marksToInput != nil {
		m.ModeCancelMarkMode()
		return false
	}
	return true
}

func (m *mainModel) ModeEnterMarkMode() {
	//TODO: if already in Mark mode, ignore.
	selected := m.Selected()
	if selected != nil {
		if selected.Number() != 0 || selected.Locked() {
			//Dion't enter mark mode.
			return
		}
	}
	m.marksToInput = make([]int, 0)
}

func (m *mainModel) ModeCommitMarkMode() {
	for _, num := range m.marksToInput {
		m.ToggleSelectedMark(num)
	}
	m.marksToInput = nil
}

func (m *mainModel) ModeCancelMarkMode() {
	m.marksToInput = nil
}

func (m *mainModel) ModeInputNumber(num int) {
	if m.marksToInput == nil {
		m.SetSelectedNumber(num)
	} else {
		m.marksToInput = append(m.marksToInput, num)
	}
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
