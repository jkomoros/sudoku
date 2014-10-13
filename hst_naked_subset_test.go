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

	techniqueName := "Naked Pair Col"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

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
	if len(step.TargetNums) != 2 || !step.TargetNums.SameContentAs([]int{2, 3}) {
		t.Log("Naked pair col found the wrong numbers: ", step.TargetNums)
		t.Fail()
	}

	description := solver.Description(step)
	if description != "2 and 3 are only possible in (6,8) and (7,8), which means that they can't be in any other cell in column 8" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	step.Apply(grid)
	firstNum := step.TargetNums[0]
	secondNum := step.TargetNums[1]
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

	techniqueName := "Naked Pair Row"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

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
	if len(step.TargetNums) != 2 || !step.TargetNums.SameContentAs([]int{2, 3}) {
		t.Log("Naked pair row found the wrong numbers: ", step.TargetNums)
		t.Fail()
	}

	description := solver.Description(step)
	if description != "2 and 3 are only possible in (8,6) and (8,7), which means that they can't be in any other cell in row 8" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	step.Apply(grid)
	firstNum := step.TargetNums[0]
	secondNum := step.TargetNums[1]
	for _, cell := range step.TargetCells {
		if cell.Possible(firstNum) || cell.Possible(secondNum) {
			t.Log("Naked Pair row found was not appleid correctly")
			t.Fail()
		}
	}

	grid.Done()
}

func TestNakedPairBlock(t *testing.T) {

	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		pointerCells: []cellRef{{0, 0}, {0, 1}},
		targetSame:   GROUP_BLOCK,
		targetGroup:  0,
		targetNums:   IntSlice([]int{1, 2}),
		description:  "1 and 2 are only possible in (0,0) and (0,1), which means that they can't be in any other cell in block 0",
	}
	humanSolveTechniqueTestHelper(t, "nakedpairblock1.sdk", "Naked Pair Block", options)

}

func TestNakedTriple(t *testing.T) {
	//TODO: test for col and block as well

	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{4, 0}, {4, 1}, {4, 2}, {4, 6}, {4, 7}, {4, 8}},
		pointerCells: []cellRef{{4, 3}, {4, 4}, {4, 5}},
		targetSame:   GROUP_ROW,
		targetGroup:  4,
		targetNums:   IntSlice([]int{3, 5, 8}),
		description:  "3, 5, and 8 are only possible in (4,3), (4,4), and (4,5), which means that they can't be in any other cell in row 4",
	}
	humanSolveTechniqueTestHelper(t, "nakedtriplet2.sdk", "Naked Triple Row", options)
}
