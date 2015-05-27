package sudoku

import (
	"testing"
)

func TestSubsetCellsWithNPossibilities(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
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
	if result[0].Row() != 6 || result[0].Col() != 8 || result[1].Row() != 7 || result[1].Col() != 8 {
		t.Log("Subset cells came back with wrong cells: ", result)
		t.Fail()
	}
}

func TestNakedPairCol(t *testing.T) {

	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{0, 8}, {1, 8}, {2, 8}, {3, 8}, {4, 8}, {5, 8}, {8, 8}},
		pointerCells: []cellRef{{6, 8}, {7, 8}},
		targetSame:   _GROUP_COL,
		targetGroup:  8,
		targetNums:   IntSlice([]int{2, 3}),
		description:  "2 and 3 are only possible in (6,8) and (7,8), which means that they can't be in any other cell in column 8",
	}
	humanSolveTechniqueTestHelper(t, "nakedpair3.sdk", "Naked Pair Col", options)
	techniqueVariantsTestHelper(t, "Naked Pair Col")

}

func TestNakedPairRow(t *testing.T) {

	options := solveTechniqueTestHelperOptions{
		transpose:    true,
		targetCells:  []cellRef{{8, 0}, {8, 1}, {8, 2}, {8, 3}, {8, 4}, {8, 5}, {8, 8}},
		pointerCells: []cellRef{{8, 6}, {8, 7}},
		targetSame:   _GROUP_ROW,
		targetGroup:  8,
		targetNums:   IntSlice([]int{2, 3}),
		description:  "2 and 3 are only possible in (8,6) and (8,7), which means that they can't be in any other cell in row 8",
	}
	humanSolveTechniqueTestHelper(t, "nakedpair3.sdk", "Naked Pair Row", options)
	techniqueVariantsTestHelper(t, "Naked Pair Row")

}

func TestNakedPairBlock(t *testing.T) {

	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		pointerCells: []cellRef{{0, 0}, {0, 1}},
		targetSame:   _GROUP_BLOCK,
		targetGroup:  0,
		targetNums:   IntSlice([]int{1, 2}),
		description:  "1 and 2 are only possible in (0,0) and (0,1), which means that they can't be in any other cell in block 0",
	}
	humanSolveTechniqueTestHelper(t, "nakedpairblock1.sdk", "Naked Pair Block", options)
	techniqueVariantsTestHelper(t, "Naked Pair Block")

}

func TestNakedTriple(t *testing.T) {
	//TODO: test for col and block as well

	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{4, 0}, {4, 1}, {4, 2}, {4, 6}, {4, 7}, {4, 8}},
		pointerCells: []cellRef{{4, 3}, {4, 4}, {4, 5}},
		targetSame:   _GROUP_ROW,
		targetGroup:  4,
		targetNums:   IntSlice([]int{3, 5, 8}),
		description:  "3, 5, and 8 are only possible in (4,3), (4,4), and (4,5), which means that they can't be in any other cell in row 4",
	}
	humanSolveTechniqueTestHelper(t, "nakedtriplet2.sdk", "Naked Triple Row", options)
	techniqueVariantsTestHelper(t, "Naked Triple Row")
}

//TODO: test naked quad techniques. (We don't have an easy one that requires it off hand.)
