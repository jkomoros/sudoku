package sudoku

import (
	"testing"
)

func TestDokugen(t *testing.T) {
	//TODO: test that neighbors are alerted correctly about SetNumbers happening.
	grid := NewGrid()
	row, col, num := 3, 3, 3
	target := grid.MutableCell(row, col)
	target.SetNumber(num)
	for _, cell := range grid.Col(col) {
		if cell == target {
			continue
		}
		if cell.Possible(num) {
			t.Log("Neighbors in the same column did not have their possibles updated.")
			t.Fail()
		}
	}
	for _, cell := range grid.Row(col) {
		if cell == target {
			continue
		}
		if cell.Possible(num) {
			t.Log("Neighbors in the same row did not have their possibles updated.")
			t.Fail()
		}
	}
	for _, cell := range grid.Block(target.Block()) {
		if cell == target {
			continue
		}
		if cell.Possible(num) {
			t.Log("Neighbors in the same row did not have their possibles updated.")
			t.Fail()
		}
	}

	if grid.impl().fillSimpleCells() != 0 {
		t.Log("We filled more than 0 cells even though there aren't any cells to obviously fill!")
		t.Fail()
	}

	//TODO: test that fillSimpleCells does work for cases where it should be obvious.
}
