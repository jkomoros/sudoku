package sudoku

import (
	"testing"
)

func TestSubsetCellsWithNPossibilities(t *testing.T) {
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath("nakedpair3.sdk")) {
		t.Log("Failed to load nakedpair3.sdk")
		t.Fail()
	}
	results := subsetCellsWithNPossibilities(2, grid.Col(DIM-1))
	if len(results) != 1 {
		t.Log("Didn't get right number of subset cells with n possibilities: ", len(results))
		t.FailNow()
	}
	result := results[0]
	if len(result) != 2 {
		t.Log("Number of subset cells did not match k: ", len(result))
		t.Fail()
	}
	if result[0].Row != 6 || result[0].Col != 8 || result[1].Row != 7 || result[1].Col != 8 {
		t.Log("Subset cells came back with wrong cells: ", result)
		t.Fail()
	}

	grid.Done()
}

func TestNakedPairCol(t *testing.T) {
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath("nakedpair3.sdk")) {
		t.Log("Failed to load nakedpair3.sdk")
		t.Fail()
	}

	solver := &nakedPairCol{}
	steps := solver.Find(grid)
	if len(steps) == 0 {
		t.Log("The naked pair col didn't find a cell it should have.")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != DIM-2 {
		t.Log("The naked pair col had the wrong number of target cells")
		t.Fail()
	}
	if len(step.PointerCells) != 2 {
		t.Log("The naked pair col had the wrong number of pointer clles")
		t.Fail()
	}
	if !step.TargetCells.SameCol() || step.TargetCells.Col() != 8 {
		t.Log("The target cells in the naked pair col were wrong col")
		t.Fail()
	}
	if len(step.Nums) != 2 || !step.Nums.SameContentAs([]int{2, 3}) {
		t.Log("Naked pair col found the wrong numbers: ", step.Nums)
		t.Fail()
	}
	step.Apply(grid)
	firstNum := step.Nums[0]
	secondNum := step.Nums[1]
	for _, cell := range step.TargetCells {
		if cell.Possible(firstNum) || cell.Possible(secondNum) {
			t.Log("Naked Pair col found was not appleid correctly")
			t.Fail()
		}
	}

	grid.Done()
}

func TestNakedPairRow(t *testing.T) {
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath("nakedpair3.sdk")) {
		t.Log("Failed to load nakedpair3.sdk")
		t.Fail()
	}
	grid = grid.transpose()
	solver := &nakedPairRow{}
	steps := solver.Find(grid)
	if len(steps) == 0 {
		t.Log("The naked pair row didn't find a cell it should have.")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != DIM-2 {
		t.Log("The naked pair row had the wrong number of target cells")
		t.Fail()
	}
	if len(step.PointerCells) != 2 {
		t.Log("The naked pair row had the wrong number of pointer clles")
		t.Fail()
	}
	if !step.TargetCells.SameRow() || step.TargetCells.Row() != 8 {
		t.Log("The target cells in the naked pair row were wrong row")
		t.Fail()
	}
	if len(step.Nums) != 2 || !step.Nums.SameContentAs([]int{2, 3}) {
		t.Log("Naked pair row found the wrong numbers: ", step.Nums)
		t.Fail()
	}
	step.Apply(grid)
	firstNum := step.Nums[0]
	secondNum := step.Nums[1]
	for _, cell := range step.TargetCells {
		if cell.Possible(firstNum) || cell.Possible(secondNum) {
			t.Log("Naked Pair row found was not appleid correctly")
			t.Fail()
		}
	}

	grid.Done()
}

func TestNakedPairBlock(t *testing.T) {
	grid := NewGrid()
	grid.Load(NAKED_PAIR_BLOCK_GRID)
	solver := &nakedPairBlock{}
	steps := solver.Find(grid)
	if len(steps) == 0 {
		t.Log("The naked pair block didn't find a cell it should have.")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != DIM-2 {
		t.Log("The naked pair block had the wrong number of target cells")
		t.Fail()
	}
	if len(step.PointerCells) != 2 {
		t.Log("The naked pair block had the wrong number of pointer clles")
		t.Fail()
	}
	if !step.TargetCells.SameBlock() || step.TargetCells.Block() != 0 {
		t.Log("The target cells in the naked pair block were wrong block")
		t.Fail()
	}
	if len(step.Nums) != 2 || !step.Nums.SameContentAs([]int{1, 2}) {
		t.Log("Naked pair block found the wrong numbers: ", step.Nums)
		t.Fail()
	}
	step.Apply(grid)
	firstNum := step.Nums[0]
	secondNum := step.Nums[1]
	for _, cell := range step.TargetCells {
		if cell.Possible(firstNum) || cell.Possible(secondNum) {
			t.Log("Naked Pair block found was not appleid correctly")
			t.Fail()
		}
	}

	grid.Done()
}

func TestNakedTriple(t *testing.T) {
	//TODO: test for col and block as well
	grid := NewGrid()
	grid.LoadFromFile(puzzlePath("nakedtriplet2.sdk"))
	solver := &nakedTripleRow{}
	steps := solver.Find(grid)

	if len(steps) == 0 {
		t.Log("The naked triple row didn't find a cell it should have.")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != DIM-3 {
		t.Log("The naked triple row had the wrong number of target cells")
		t.Fail()
	}
	if len(step.PointerCells) != 3 {
		t.Log("The naked triple row had the wrong number of pointer clles")
		t.Fail()
	}
	if !step.TargetCells.SameRow() || step.TargetCells.Row() != 4 {
		t.Log("The target cells in the naked triple row were wrong row")
		t.Fail()
	}
	if len(step.Nums) != 3 || !step.Nums.SameContentAs([]int{3, 5, 8}) {
		t.Log("Naked triple row found the wrong numbers: ", step.Nums)
		t.Fail()
	}
	step.Apply(grid)
	firstNum := step.Nums[0]
	secondNum := step.Nums[1]
	thirdNum := step.Nums[2]
	for _, cell := range step.TargetCells {
		if cell.Possible(firstNum) || cell.Possible(secondNum) || cell.Possible(thirdNum) {
			t.Log("Naked triple row found was not appleid correctly")
			t.Fail()
		}
	}

	grid.Done()
}
