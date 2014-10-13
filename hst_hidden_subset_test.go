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
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{4, 7}, {4, 8}},
		pointerCells: []cellRef{{4, 7}, {4, 8}},
		//Yes, in this case we want them to be the same row.
		targetSame:  GROUP_ROW,
		targetGroup: 4,
		targetNums:  IntSlice([]int{7, 8, 2}),
		pointerNums: IntSlice([]int{3, 5}),
		description: "3 and 5 are only possible in (4,7) and (4,8) within row 4, which means that only those numbers could be in those cells",
	}
	humanSolveTechniqueTestHelper(t, "hiddenpair1_filled.sdk", "Hidden Pair Row", options)

}

func TestHiddenPairCol(t *testing.T) {

	options := solveTechniqueTestHelperOptions{
		transpose:    true,
		targetCells:  []cellRef{{7, 4}, {8, 4}},
		pointerCells: []cellRef{{7, 4}, {8, 4}},
		//Yes, in this case we want them to be the same row.
		targetSame:  GROUP_COL,
		targetGroup: 4,
		targetNums:  IntSlice([]int{7, 8, 2}),
		pointerNums: IntSlice([]int{3, 5}),
		description: "3 and 5 are only possible in (7,4) and (8,4) within column 4, which means that only those numbers could be in those cells",
	}
	humanSolveTechniqueTestHelper(t, "hiddenpair1_filled.sdk", "Hidden Pair Col", options)

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
