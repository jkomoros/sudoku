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

	nums := row.CollectNums(func(cell *Cell) int {
		return cell.Row
	})

	if !nums.Same() {
		t.Log("Collecting rows gave us different numbers/.")
		t.Fail()
	}
}

func TestIntList(t *testing.T) {
	numArr := [...]int{1, 1, 1}
	if !intList(numArr[:]).Same() {
		t.Log("We didn't think that a num list with all of the same ints was the same.")
		t.Fail()
	}
	differentNumArr := [...]int{1, 2, 1}
	if intList(differentNumArr[:]).Same() {
		t.Log("We thought a list of different ints were the same")
		t.Fail()
	}
}
