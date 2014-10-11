package sudoku

import (
	"testing"
)

func TestHiddenPairRow(t *testing.T) {
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath("hiddenpair1_filled.sdk")) {
		t.Log("Failed to load hiddenpair1_filled.sdk")
		t.Fail()
	}

	solver := &hiddenPairRow{}
	steps := solver.Find(grid)
	if len(steps) == 0 {
		t.Log("The hidden pair row didn't find a cell it should have.")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != 2 {
		t.Log("The hidden pair row had the wrong number of target cells: ", len(step.TargetCells))
		t.FailNow()
	}
	if len(step.PointerCells) != 2 {
		t.Log("The hidden pair row had the wrong number of pointer cells: ", len(step.PointerCells))
		t.FailNow()
	}
	if step.TargetCells[0] != step.PointerCells[0] || step.TargetCells[1] != step.PointerCells[1] {
		t.Error("Hidden Pair Row did not have the same target and pointer cells")
	}
	if !step.TargetCells.SameRow() || step.TargetCells.Row() != 4 {
		t.Log("The target cells in the hidden pair row were wrong row")
		t.Fail()
	}
	if len(step.Nums) != 3 || !step.Nums.SameContentAs([]int{7, 8, 2}) {
		t.Log("Hidden pair row found the wrong numbers: ", step.Nums)
		t.Fail()
	}
	step.Apply(grid)
	firstNum := step.Nums[0]
	secondNum := step.Nums[1]
	thirdNum := step.Nums[2]
	for _, cell := range step.TargetCells {

		for i := 1; i <= DIM; i++ {
			if i == firstNum || i == secondNum || i == thirdNum {
				if cell.Possible(i) {
					t.Error("Hidden Pair Row was not applied correctly; it did not clear the right numbers.")
				}
			}
		}
	}

	grid.Done()
}

func TestHiddenPairCol(t *testing.T) {
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath("hiddenpair1_filled.sdk")) {
		t.Log("Failed to load hiddenpair1_filled.sdk")
		t.Fail()
	}

	grid = grid.transpose()

	solver := &hiddenPairCol{}
	steps := solver.Find(grid)
	if len(steps) == 0 {
		t.Log("The hidden pair col didn't find a cell it should have.")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != 2 {
		t.Log("The hidden pair col had the wrong number of target cells: ", len(step.TargetCells))
		t.FailNow()
	}
	if len(step.PointerCells) != 2 {
		t.Log("The hidden pair col had the wrong number of pointer cells: ", len(step.PointerCells))
		t.FailNow()
	}
	if step.TargetCells[0] != step.PointerCells[0] || step.TargetCells[1] != step.PointerCells[1] {
		t.Error("Hidden Pair col did not have the same target and pointer cells")
	}
	if !step.TargetCells.SameCol() || step.TargetCells.Col() != 4 {
		t.Log("The target cells in the hidden pair col were wrong row")
		t.Fail()
	}
	if len(step.Nums) != 3 || !step.Nums.SameContentAs([]int{7, 8, 2}) {
		t.Log("Hidden pair col found the wrong numbers: ", step.Nums)
		t.Fail()
	}
	step.Apply(grid)
	firstNum := step.Nums[0]
	secondNum := step.Nums[1]
	thirdNum := step.Nums[2]
	for _, cell := range step.TargetCells {

		for i := 1; i <= DIM; i++ {
			if i == firstNum || i == secondNum || i == thirdNum {
				if cell.Possible(i) {
					t.Error("Hidden Pair col was not applied correctly; it did not clear the right numbers.")
				}
			}
		}
	}

	grid.Done()
}

func TestHiddenPairBlock(t *testing.T) {
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath("hiddenpair1_filled.sdk")) {
		t.Log("Failed to load hiddenpair1_filled.sdk")
		t.Fail()
	}

	solver := &hiddenPairBlock{}
	steps := solver.Find(grid)
	if len(steps) == 0 {
		t.Log("The hidden pair block didn't find a cell it should have.")
		t.FailNow()
	}

	step := steps[0]

	if len(step.TargetCells) != 2 {
		t.Log("The hidden pair block had the wrong number of target cells: ", len(step.TargetCells))
		t.FailNow()
	}
	if len(step.PointerCells) != 2 {
		t.Log("The hidden pair block had the wrong number of pointer cells: ", len(step.PointerCells))
		t.FailNow()
	}
	if step.TargetCells[0] != step.PointerCells[0] || step.TargetCells[1] != step.PointerCells[1] {
		t.Error("Hidden Pair block did not have the same target and pointer cells")
	}
	if !step.TargetCells.SameRow() || step.TargetCells.Row() != 4 {
		t.Log("The target cells in the hidden pair block were wrong row")
		t.Fail()
	}
	if len(step.Nums) != 3 || !step.Nums.SameContentAs([]int{7, 8, 2}) {
		t.Log("Hidden pair block found the wrong numbers: ", step.Nums)
		t.Fail()
	}
	step.Apply(grid)
	firstNum := step.Nums[0]
	secondNum := step.Nums[1]
	thirdNum := step.Nums[2]
	for _, cell := range step.TargetCells {

		for i := 1; i <= DIM; i++ {
			if i == firstNum || i == secondNum || i == thirdNum {
				if cell.Possible(i) {
					t.Error("Hidden Pair block was not applied correctly; it did not clear the right numbers.")
				}
			}
		}
	}

	grid.Done()
}
