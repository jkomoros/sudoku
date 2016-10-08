package sudoku

import (
	"testing"
)

func TestObviousInCollectionRow(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells: []CellReference{{2, 3}},
		targetSame:  _GROUP_ROW,
		targetGroup: 2,
		targetNums:  IntSlice([]int{7}),
		description: "(2,3) is the only cell in row 2 that is unfilled, and it must be 7",
	}
	humanSolveTechniqueTestHelper(t, "obviousrow.sdk", "Obvious In Row", options)
	techniqueVariantsTestHelper(t, "Obvious In Row")

}

func TestObviousInCollectionCol(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		transpose:   true,
		targetCells: []CellReference{{3, 2}},
		targetSame:  _GROUP_COL,
		targetGroup: 2,
		targetNums:  IntSlice([]int{7}),
		description: "(3,2) is the only cell in column 2 that is unfilled, and it must be 7",
	}
	humanSolveTechniqueTestHelper(t, "obviousrow.sdk", "Obvious In Col", options)
	techniqueVariantsTestHelper(t, "Obvious In Col")

}

func TestObviousInCollectionBlock(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells: []CellReference{{4, 1}},
		targetSame:  _GROUP_BLOCK,
		targetGroup: 3,
		targetNums:  IntSlice([]int{9}),
		description: "(4,1) is the only cell in block 3 that is unfilled, and it must be 9",
	}
	humanSolveTechniqueTestHelper(t, "obviousblock.sdk", "Obvious In Block", options)
	techniqueVariantsTestHelper(t, "Obvious In Block")

}

func TestSolveOnlyLegalNumber(t *testing.T) {

	techniqueVariantsTestHelper(t, "Only Legal Number")

	grid := MutableLoadSDK(SOLVED_TEST_GRID)
	cell := grid.MutableCell(3, 3)
	num := cell.Number()
	cell.SetNumber(0)

	//Now that cell should be filled by this technique.

	techniqueName := "Only Legal Number"
	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	results := make(chan *SolveStep, DIM*DIM)
	done := make(chan bool)

	coordinator := &channelFindCoordinator{
		results: results,
		done:    done,
	}

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	solver.find(grid, coordinator)

	//TODO: test that Find exits early when done is closed. (or maybe just doesn't send after done is closed)
	close(done)

	var step *SolveStep

	//TODO: test cases where we expectmultipel results...
	select {
	case step = <-results:
	default:
		t.Fatal(techniqueName, " didn't find a cell it should have.")
	}

	description := solver.Description(step)
	if description != "3 is the only remaining valid number for that cell" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col() != 3 || cellFromStep.Row() != 3 {
		t.Log("The only legal number technique identified the wrong cell.")
		t.Fail()
	}
	numFromStep := step.TargetNums[0]

	if numFromStep != num {
		t.Log("The only legal number technique identified the wrong number.")
		t.Fail()
	}
	if grid.Solved() {
		t.Log("The only legal number technique did actually mutate the grid.")
		t.Fail()
	}
}

//TODO: use the test solve helper func for these three tests.
func TestNecessaryInRow(t *testing.T) {
	//Load up a solved grid
	grid := MutableLoad(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.MutableRow(3) {
		cellI := cell.(*mutableCellImpl)
		cellI.number = 0
		copy(cellI.impossibles[:], impossibles)
	}

	cell := grid.MutableCell(3, 3)
	cellI := cell.(*mutableCellImpl)
	//This is the only cell where DIM will be allowed.
	cellI.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	techniqueName := "Necessary In Row"
	solver := techniquesByName[techniqueName]

	techniqueVariantsTestHelper(t, "Necessary In Row")

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	results := make(chan *SolveStep, DIM*DIM)
	done := make(chan bool)

	coordinator := &channelFindCoordinator{
		results: results,
		done:    done,
	}

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	solver.find(grid, coordinator)

	//TODO: test that Find exits early when done is closed. (or maybe just doesn't send after done is closed)
	close(done)

	var step *SolveStep

	//TODO: test cases where we expectmultipel results...
	select {
	case step = <-results:
	default:
		t.Fatal(techniqueName, " didn't find a cell it should have.")
	}

	description := solver.Description(step)
	if description != "9 is required in the 3 row, and 3 is the only column it fits" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col() != 3 || cellFromStep.Row() != 3 {
		t.Log("The necessary in row technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.TargetNums[0]

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

	//Load up a solved grid
	grid := MutableLoadSDK(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.MutableCol(3) {
		cellI := cell.(*mutableCellImpl)
		cellI.number = 0
		copy(cellI.impossibles[:], impossibles)
	}

	cell := grid.MutableCell(3, 3)
	cellI := cell.(*mutableCellImpl)
	//This is the only cell where DIM will be allowed.
	cellI.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	techniqueName := "Necessary In Col"
	solver := techniquesByName[techniqueName]

	techniqueVariantsTestHelper(t, "Necessary In Col")

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	results := make(chan *SolveStep, DIM*DIM)
	done := make(chan bool)

	coordinator := &channelFindCoordinator{
		results: results,
		done:    done,
	}

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	solver.find(grid, coordinator)

	//TODO: test that Find exits early when done is closed. (or maybe just doesn't send after done is closed)
	close(done)

	var step *SolveStep

	//TODO: test cases where we expectmultipel results...
	select {
	case step = <-results:
	default:
		t.Fatal(techniqueName, " didn't find a cell it should have.")
	}

	description := solver.Description(step)
	if description != "9 is required in the 3 column, and 3 is the only row it fits" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col() != 3 || cellFromStep.Row() != 3 {
		t.Log("The necessary in col technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.TargetNums[0]

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

	//Load up a solved grid
	grid := MutableLoadSDK(SOLVED_TEST_GRID)

	//We're going to cheat an set up an unrealistic grid.

	impossibles := make([]int, DIM)

	for i := 0; i < DIM-1; i++ {
		impossibles[i] = 0
	}
	impossibles[DIM-1] = 1

	//SetNumber will affect the other cells in row, so do it first.
	for _, cell := range grid.MutableBlock(4) {
		cellI := cell.(*mutableCellImpl)
		cellI.number = 0
		copy(cellI.impossibles[:], impossibles)
	}

	cell := grid.MutableCell(3, 3)
	cellI := cell.(*mutableCellImpl)
	//This is the only cell where DIM will be allowed.
	cellI.impossibles[DIM-1] = 0

	//Now that cell should be filled by this technique.

	techniqueName := "Necessary In Block"
	solver := techniquesByName[techniqueName]

	techniqueVariantsTestHelper(t, "Necessary In Block")

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	results := make(chan *SolveStep, DIM*DIM)
	done := make(chan bool)

	coordinator := &channelFindCoordinator{
		results: results,
		done:    done,
	}

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	solver.find(grid, coordinator)

	//TODO: test that Find exits early when done is closed. (or maybe just doesn't send after done is closed)
	close(done)

	var step *SolveStep

	//TODO: test cases where we expectmultipel results...
	select {
	case step = <-results:
	default:
		t.Fatal(techniqueName, " didn't find a cell it should have.")
	}

	description := solver.Description(step)
	if description != "9 is required in the 4 block, and (3,3) is the only cell it fits" {
		t.Error("Wrong description for ", techniqueName, ": ", description)
	}

	cellFromStep := step.TargetCells[0]

	if cellFromStep.Col() != 3 || cellFromStep.Row() != 3 {
		t.Log("The necessary in block technique identified the wrong cell.")
		t.Fail()
	}

	numFromStep := step.TargetNums[0]

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
