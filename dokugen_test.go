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

	consumeCells(grid.queue, DIM*DIM, "After reset to base", t)

	target.SetNumber(num)

	consumeCells(grid.queue, (DIM-1)*3-(BLOCK_DIM-1)*2, "After setting one number", t)

	if grid.fillSimpleCells() != 0 {
		t.Log("We filled more than 0 cells even though there aren't any cells to obviously fill!")
		t.Fail()
	}

	//TODO: test that fillSimpleCells does work for cases where it should be obvious.
}

func consumeCells(queue *FiniteQueue, expected int, msg string, t *testing.T) {
	count := 0
	queuedCell := queue.Get()
	for queuedCell != nil {
		count++
		queuedCell = queue.Get()
	}
	if count != expected {
		t.Log("The grid's queue didn't have the correct number of items in it (", msg, "). Expected ", expected, " but actually got ", count)
		t.Fail()
	}
}
