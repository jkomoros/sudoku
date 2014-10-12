package sudoku

import (
	"testing"
)

const POINTING_PAIR_ROW_GRID = `3|.|6|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
4|.|5|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|7|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.`
const POINTING_PAIR_COL_GRID = `3|.|6|.|.|.|.|.|.
.|.|.|.|7|.|.|.|.
4|.|5|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.`

func TestPointingPairCol(t *testing.T) {
	grid := NewGrid()
	grid.Load(POINTING_PAIR_COL_GRID)

	techniqueName := "Pointing Pair Col"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := solver.Find(grid)
	if len(steps) == 0 {
		t.Log("The pointing pair col didn't find a cell it should have")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != BLOCK_DIM*2 {
		t.Log("The pointing pair col gave back the wrong number of target cells")
		t.Fail()
	}
	if len(step.PointerCells) != BLOCK_DIM-1 {
		t.Log("The pointing pair col gave back the wrong number of pointer cells")
		t.Fail()
	}
	if !step.TargetCells.SameCol() || step.TargetCells.Col() != 1 {
		t.Log("The target cells in the pointing pair col technique were wrong col")
		t.Fail()
	}
	if len(step.Nums) != 1 || step.Nums[0] != 7 {
		t.Log("Pointing pair col technique gave the wrong number")
		t.Fail()
	}
	step.Apply(grid)
	num := step.Nums[0]
	for _, cell := range step.TargetCells {
		if cell.Possible(num) {
			t.Log("The pointing pairs col technique was not applied correclty")
			t.Fail()
		}
	}

	grid.Done()
}

func TestPointingPairRow(t *testing.T) {
	grid := NewGrid()
	grid.Load(POINTING_PAIR_ROW_GRID)

	techniqueName := "Pointing Pair Row"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := solver.Find(grid)
	if len(steps) == 0 {
		t.Log("The pointing pair row didn't find a cell it should have")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != BLOCK_DIM*2 {
		t.Log("The pointing pair row gave back the wrong number of target cells")
		t.Fail()
	}
	if len(step.PointerCells) != BLOCK_DIM-1 {
		t.Log("The pointing pair row gave back the wrong number of pointer cells")
		t.Fail()
	}
	if !step.TargetCells.SameRow() || step.TargetCells.Row() != 1 {
		t.Log("The target cells in the pointing pair row technique were wrong row")
		t.Fail()
	}
	if len(step.Nums) != 1 || step.Nums[0] != 7 {
		t.Log("Pointing pair row technique gave the wrong number")
		t.Fail()
	}
	step.Apply(grid)
	num := step.Nums[0]
	for _, cell := range step.TargetCells {
		if cell.Possible(num) {
			t.Log("The pointing pairs row technique was not applied correclty")
			t.Fail()
		}
	}

	grid.Done()
}
