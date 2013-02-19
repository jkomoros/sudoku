package sudoku

import (
	"testing"
)

func TestBasicCellList(t *testing.T) {
	grid := NewGrid()
	grid.Load(SOLVED_TEST_GRID)
	row := CellList(grid.Row(2))
	if !row.SameRow() {
		t.Log("The items of a row were not all of the same row.")
		t.Fail()
	}
}
