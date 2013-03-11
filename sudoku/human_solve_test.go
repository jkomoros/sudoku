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

//The following example comes from http://www.sadmansoftware.com/sudoku/nakedsubset.htm
const NAKED_PAIR_COL_GRID = `3|.|5|.|.|.|7|.|9
.|9|.|3|7|.|.|.|6
.|7|.|.|.|1|.|3|.
6|.|4|.|.|.|3|.|.
.|.|.|4|.|5|.|.|.
.|.|9|.|.|.|.|.|4
.|4|.|1|.|.|9|8|.
9|.|.|.|8|7|.|.|.
8|.|2|.|.|.|1|7|5`

func TestSolveOnlyLegalNumber(t *testing.T) {
	grid := NewGrid()
	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)
	cell := grid.Cell(3, 3)
	num := cell.Number()
	cell.SetNumber(0)

	//Now that cell should be filled by this technique.

	solver := &nakedSingleTechnique{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The only legal number technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The only legal number technique identified the wrong cell.")
		t.Fail()
	}
	numFromStep := step.Nums[0]

	if numFromStep != num {
		t.Log("The only legal number technique identified the wrong number.")
		t.Fail()
	}
	if grid.Solved() {
		t.Log("The only legal number technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestNecessaryInRow(t *testing.T) {
	grid := NewGrid()
	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.Row(3) {
		cell.number = 0
		copy(cell.impossibles[:], impossibles)
	}

	cell := grid.Cell(3, 3)
	//This is the only cell where DIM will be allowed.
	cell.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	solver := &hiddenSingleInRow{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in row technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in row technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.Nums[0]

	if numFromStep != DIM {
		t.Log("The necessary in row technique identified the wrong number.")
		t.Fail()
	}
	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in row technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestNecessaryInCol(t *testing.T) {
	grid := NewGrid()
	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.Col(3) {
		cell.number = 0
		copy(cell.impossibles[:], impossibles)
	}

	cell := grid.Cell(3, 3)
	//This is the only cell where DIM will be allowed.
	cell.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	solver := &hiddenSingleInCol{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in col technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in col technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.Nums[0]

	if numFromStep != DIM {
		t.Log("The necessary in col technique identified the wrong number.")
		t.Fail()
	}
	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in col technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestNecessaryInBlock(t *testing.T) {
	grid := NewGrid()
	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.Block(4) {
		cell.number = 0
		copy(cell.impossibles[:], impossibles)
	}

	cell := grid.Cell(3, 3)
	//This is the only cell where DIM will be allowed.
	cell.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	solver := &hiddenSingleInBlock{}

	step := solver.Find(grid)

	if step == nil {
		t.Log("The necessary in block technique did not solve a puzzle it should have.")
		t.FailNow()
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col != 3 || cellFromStep.Row != 3 {
		t.Log("The necessary in block technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.Nums[0]

	if numFromStep != DIM {
		t.Log("The necessary in block technique identified the wrong number.")
		t.Fail()
	}
	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in block technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestPointingPairCol(t *testing.T) {
	grid := NewGrid()
	grid.Load(POINTING_PAIR_COL_GRID)
	solver := &pointingPairCol{}
	step := solver.Find(grid)
	if step == nil {
		t.Log("The pointing pair col didn't find a cell it should have")
		t.FailNow()
	}
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
}

func TestPointingPairRow(t *testing.T) {
	grid := NewGrid()
	grid.Load(POINTING_PAIR_ROW_GRID)
	solver := &pointingPairRow{}
	step := solver.Find(grid)
	if step == nil {
		t.Log("The pointing pair row didn't find a cell it should have")
		t.FailNow()
	}
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
}

func TestNakedPairCol(t *testing.T) {
	grid := NewGrid()
	grid.Load(NAKED_PAIR_COL_GRID)
	solver := &nakedPairCol{}
	step := solver.Find(grid)
	if step == nil {
		t.Log("The naked pair col didn't find a cell it should have.")
		t.FailNow()
	}
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
	if len(step.Nums) != 2 || step.Nums[0] != 2 || step.Nums[1] != 3 {
		t.Log("Naked pair col found the wrong numbers")
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
}

func TestHumanSolve(t *testing.T) {
	grid := NewGrid()
	grid.Load(TEST_GRID)
	steps := grid.HumanSolve()
	//TODO: test to make sure that we use a wealth of different techniques. This will require a cooked random for testing.
	if steps == nil {
		t.Log("Human solve returned 0 techniques")
		t.Fail()
	}
	if !grid.Solved() {
		t.Log("Human solve failed to solve the simple grid.")
		t.Fail()
	}
}
