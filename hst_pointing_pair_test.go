package sudoku

import (
	"testing"
)

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
	if len(step.TargetNums) != 1 || step.TargetNums[0] != 7 {
		t.Log("Pointing pair col technique gave the wrong number")
		t.Fail()
	}

	description := solver.Description(step)
	if description != "7 is only possible in column 1 of block 0, which means it can't be in any other cell in that column not in that block" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	step.Apply(grid)
	num := step.TargetNums[0]
	for _, cell := range step.TargetCells {
		if cell.Possible(num) {
			t.Log("The pointing pairs col technique was not applied correclty")
			t.Fail()
		}
	}

	grid.Done()
}

func TestPointingPairRow(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}},
		pointerCells: []cellRef{{1, 0}, {1, 2}},
		targetSame:   GROUP_ROW,
		targetGroup:  1,
		targetNums:   IntSlice([]int{7}),
		description:  "7 is only possible in row 1 of block 0, which means it can't be in any other cell in that row not in that block",
	}
	humanSolveTechniqueTestHelper(t, "pointingpairrow1.sdk", "Pointing Pair Row", options)

}
