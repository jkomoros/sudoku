package dokugen

import (
	"testing"
)

func TestDokugen(t *testing.T) {
	//TODO: test that neighbors are alerted correctly about SetNumbers happening.
	grid := NewGrid()
	row, col, num := 3, 3, 3
	target := grid.Cell(row, col)
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
	for _, cell := range grid.Block(target.Block) {
		if cell == target {
			continue
		}
		if cell.Possible(num) {
			t.Log("Neighbors in the same row did not have their possibles updated.")
			t.Fail()
		}
	}

	//Reset the grid to starting conditions.
	target.SetNumber(0)

	count := 0
	queuedCell := grid.queue.Get()
	for queuedCell != nil {
		count++
		queuedCell = grid.queue.Get()
	}
	if count != DIM*DIM {
		t.Log("The grid's queue didn't have the correct number of items in it: ", count)
		t.Fail()
	}
}
