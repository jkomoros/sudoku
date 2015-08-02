package sudoku

import (
	"testing"
)

func TestObviousInCollectionRow(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells: []cellRef{{2, 3}},
		targetSame:  _GROUP_ROW,
		targetGroup: 2,
		targetNums:  IntSlice([]int{7}),
		pointerCells: []cellRef{
			{2, 0},
			{2, 1},
			{2, 2},
			{2, 4},
			{2, 5},
			{2, 6},
			{2, 7},
			{2, 8},
		},
		pointerNums: IntSlice{
			1,
			2,
			3,
			4,
			5,
			6,
			8,
			9,
		},
		description:    "(2,3) is the only cell in row 2 that is unfilled, and it must be 7",
		checkIsImplied: true,
	}
	humanSolveTechniqueTestHelper(t, "obviousrow.sdk", "Obvious In Row", options)
	techniqueVariantsTestHelper(t, "Obvious In Row")

}

func TestObviousInCollectionCol(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		transpose:   true,
		targetCells: []cellRef{{3, 2}},
		targetSame:  _GROUP_COL,
		targetGroup: 2,
		targetNums:  IntSlice([]int{7}),
		pointerCells: []cellRef{
			{0, 2},
			{1, 2},
			{2, 2},
			{4, 2},
			{5, 2},
			{6, 2},
			{7, 2},
			{8, 2},
		},
		pointerNums: IntSlice{
			1,
			2,
			3,
			4,
			5,
			6,
			8,
			9,
		},
		description:    "(3,2) is the only cell in column 2 that is unfilled, and it must be 7",
		checkIsImplied: true,
	}
	humanSolveTechniqueTestHelper(t, "obviousrow.sdk", "Obvious In Col", options)
	techniqueVariantsTestHelper(t, "Obvious In Col")

}

func TestObviousInCollectionBlock(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells: []cellRef{{4, 1}},
		targetSame:  _GROUP_BLOCK,
		targetGroup: 3,
		targetNums:  IntSlice([]int{9}),
		pointerCells: []cellRef{
			{3, 0},
			{3, 1},
			{3, 2},
			{4, 0},
			{4, 2},
			{5, 0},
			{5, 1},
			{5, 2},
		},
		pointerNums: IntSlice{
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
		},
		description:    "(4,1) is the only cell in block 3 that is unfilled, and it must be 9",
		checkIsImplied: true,
	}
	humanSolveTechniqueTestHelper(t, "obviousblock.sdk", "Obvious In Block", options)
	techniqueVariantsTestHelper(t, "Obvious In Block")

}

func TestSolveOnlyLegalNumber(t *testing.T) {

	techniqueVariantsTestHelper(t, "Only Legal Number")

	grid := NewGrid()
	defer grid.Done()
	//Load up a solved grid
	grid.Load(SOLVED_TEST_GRID)
	cell := grid.Cell(3, 3)
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

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	solver.Find(grid, results, done)

	//TODO: test that Find exits early when done is closed. (or maybe just doesn't send after done is closed)
	close(done)

	var step *SolveStep

	//TODO: test cases where we expectmultipel results...
	select {
	case step = <-results:
	default:
		t.Fatal(techniqueName, " didn't find a cell it should have.")
	}

	if !step.IsImplied(grid) {
		t.Error("Only legal number step that we just found wasn't valid right afterwards.")
	}

	//TODO: test that if the grid is a bizarroGrid that isImplied is false.

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

	if !step.PointerNums.SameContentAs(IntSlice{1, 2, 4, 5, 6, 7, 8, 9}) {
		t.Error("Pointer nums set wrong")
	}

	refs := []cellRef{
		{3, 0},
		{3, 1},
		{3, 6},
		{3, 8},
		{4, 3},
		{4, 4},
		{6, 3},
		{5, 3},
		{2, 3},
		{7, 3},
		{8, 3},
		{5, 5},
		{3, 2},
		{3, 4},
		{0, 3},
		{1, 3},
		{3, 5},
		{3, 7},
		{4, 5},
		{5, 4},
	}

	if !step.PointerCells.sameAsRefs(refs) {
		t.Error("Got wrong pointer cells. Got: ", step.PointerCells, len(step.PointerCells), len(refs))
	}

	if grid.Solved() {
		t.Log("The only legal number technique did actually mutate the grid.")
		t.Fail()
	}
}

//TODO: use the test solve helper func for these three tests.
func TestNecessaryInRow(t *testing.T) {
	grid := NewGrid()

	//We DON'T call grid.done because we will have poked some unrealistic values into the cells.

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

	techniqueName := "Necessary In Row"
	solver := techniquesByName[techniqueName]

	techniqueVariantsTestHelper(t, "Necessary In Row")

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	results := make(chan *SolveStep, DIM*DIM)
	done := make(chan bool)

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	solver.Find(grid, results, done)

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

	refs := []cellRef{
		{3, 0},
		{3, 1},
		{3, 2},
		{3, 4},
		{3, 5},
		{3, 6},
		{3, 7},
		{3, 8},
	}

	if !step.PointerCells.sameAsRefs(refs) {
		t.Error("Necessary in row didn't have right pointer cells: ", step.PointerCells)
	}

	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in row technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestNecessaryInCol(t *testing.T) {
	grid := NewGrid()

	//We DON'T call grid.done because we will have poked some unrealistic values into the cells.

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

	techniqueName := "Necessary In Col"
	solver := techniquesByName[techniqueName]

	techniqueVariantsTestHelper(t, "Necessary In Col")

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	results := make(chan *SolveStep, DIM*DIM)
	done := make(chan bool)

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	solver.Find(grid, results, done)

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

	refs := []cellRef{
		{0, 3},
		{1, 3},
		{2, 3},
		{4, 3},
		{5, 3},
		{6, 3},
		{7, 3},
		{8, 3},
	}

	if !step.PointerCells.sameAsRefs(refs) {
		t.Error("Necessary in col didn't have right pointer cells: ", step.PointerCells)
	}

	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in col technique did actually mutate the grid.")
		t.Fail()
	}
}

func TestNecessaryInBlock(t *testing.T) {
	grid := NewGrid()

	//We DON'T call grid.done because we will have poked some unrealistic values into the cells.

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

	techniqueName := "Necessary In Block"
	solver := techniquesByName[techniqueName]

	techniqueVariantsTestHelper(t, "Necessary In Block")

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	results := make(chan *SolveStep, DIM*DIM)
	done := make(chan bool)

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	solver.Find(grid, results, done)

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

	refs := []cellRef{
		{3, 4},
		{3, 5},
		{4, 3},
		{4, 4},
		{4, 5},
		{5, 3},
		{5, 4},
		{5, 5},
	}

	if !step.PointerCells.sameAsRefs(refs) {
		t.Error("Necessary in row didn't have right pointer cells: ", step.PointerCells)
	}

	//Can't check if grid is solved because we un-set all the other cells in the row.
	if cell.Number() != 0 {
		t.Log("The necessary in block technique did actually mutate the grid.")
		t.Fail()
	}

}
