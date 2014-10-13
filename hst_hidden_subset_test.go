package sudoku

import (
	"testing"
)

func TestSubsetCellsWithNUniquePossibilities(t *testing.T) {
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath("hiddenpair1_filled.sdk")) {
		t.Log("Failed to load hiddenpair1_filled.sdk")
		t.Fail()
	}
	cells, nums := subsetCellsWithNUniquePossibilities(2, grid.Row(4))
	if len(cells) != 1 {
		t.Log("Didn't get right number of subset cells unique with n possibilities: ", len(cells))
		t.FailNow()
	}
	cellList := cells[0]
	numList := nums[0]
	if len(cellList) != 2 {
		t.Log("Number of subset cells did not match k: ", len(cellList))
		t.Fail()
	}
	if cellList[0].Row != 4 || cellList[0].Col != 7 || cellList[1].Row != 4 || cellList[1].Col != 8 {
		t.Log("Subset cells unique came back with wrong cells: ", cellList)
		t.Fail()
	}
	if !numList.SameContentAs(IntSlice([]int{3, 5})) {
		t.Error("Subset cells unique came back with wrong numbers: ", numList)
	}

	grid.Done()
}

func TestHiddenPairRow(t *testing.T) {
	grid := NewGrid()
	if !grid.LoadFromFile(puzzlePath("hiddenpair1_filled.sdk")) {
		t.Log("Failed to load hiddenpair1_filled.sdk")
		t.Fail()
	}

	techniqueName := "Hidden Pair Row"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

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
	if len(step.TargetNums) != 3 || !step.TargetNums.SameContentAs([]int{7, 8, 2}) {
		t.Log("Hidden pair row found the wrong numbers: ", step.TargetNums)
		t.Fail()
	}
	if len(step.PointerNums) != 2 || !step.PointerNums.SameContentAs([]int{3, 5}) {
		t.Error("Hidden pair row had the wrong pointer numbers: ", step.PointerNums)
	}

	description := solver.Description(step)
	if description != "3 and 5 are only possible in (4,7) and (4,8) within row 4, which means that only those numbers could be in those cells" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	step.Apply(grid)
	firstNum := step.TargetNums[0]
	secondNum := step.TargetNums[1]
	thirdNum := step.TargetNums[2]
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

	techniqueName := "Hidden Pair Col"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

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
	if len(step.TargetNums) != 3 || !step.TargetNums.SameContentAs([]int{7, 8, 2}) {
		t.Log("Hidden pair col found the wrong numbers: ", step.TargetNums)
		t.Fail()
	}
	if len(step.PointerNums) != 2 || !step.PointerNums.SameContentAs([]int{3, 5}) {
		t.Error("Hidden pair col had the wrong pointer numbers: ", step.PointerNums)
	}

	description := solver.Description(step)
	if description != "3 and 5 are only possible in (7,4) and (8,4) within column 4, which means that only those numbers could be in those cells" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	step.Apply(grid)
	firstNum := step.TargetNums[0]
	secondNum := step.TargetNums[1]
	thirdNum := step.TargetNums[2]
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
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{4, 7}, {4, 8}},
		pointerCells: []cellRef{{4, 7}, {4, 8}},
		//Yes, in this case we want them to be the same row.
		targetSame:  GROUP_ROW,
		targetGroup: 4,
		targetNums:  IntSlice([]int{7, 8, 2}),
		pointerNums: IntSlice([]int{3, 5}),
		description: "3 and 5 are only possible in (4,7) and (4,8) within block 5, which means that only those numbers could be in those cells",
	}
	humanSolveTechniqueTestHelper(t, "hiddenpair1_filled.sdk", "Hidden Pair Block", options)

}
