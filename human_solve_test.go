package sudoku

import (
	"testing"
	"time"
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

const NAKED_PAIR_BLOCK_GRID = `.|.|3|.|7|8|9|.|.
4|5|6|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.
.|.|.|.|.|.|.|.|.`

func TestSubsetIndexes(t *testing.T) {
	result := subsetIndexes(3, 1)
	expectedResult := [][]int{[]int{0}, []int{1}, []int{2}}
	subsetIndexHelper(t, result, expectedResult)

	result = subsetIndexes(3, 2)
	expectedResult = [][]int{[]int{0, 1}, []int{0, 2}, []int{1, 2}}
	subsetIndexHelper(t, result, expectedResult)

	result = subsetIndexes(5, 3)
	expectedResult = [][]int{[]int{0, 1, 2}, []int{0, 1, 3}, []int{0, 1, 4}, []int{0, 2, 3}, []int{0, 2, 4}, []int{0, 3, 4}, []int{1, 2, 3}, []int{1, 2, 4}, []int{1, 3, 4}, []int{2, 3, 4}}
	subsetIndexHelper(t, result, expectedResult)

	if subsetIndexes(1, 2) != nil {
		t.Log("Subset indexes returned a subset where the length is greater than the len")
		t.Fail()
	}

}

func subsetIndexHelper(t *testing.T, result [][]int, expectedResult [][]int) {
	if len(result) != len(expectedResult) {
		t.Log("subset indexes returned wrong number of results for: ", result, " :", expectedResult)
		t.FailNow()
	}
	for i, item := range result {
		if len(item) != len(expectedResult[0]) {
			t.Log("subset indexes returned a result with wrong numbrer of items ", i, " : ", result, " : ", expectedResult)
			t.FailNow()
		}
		for j, value := range item {
			if value != expectedResult[i][j] {
				t.Log("Subset indexes had wrong number at ", i, ",", j, " : ", result, " : ", expectedResult)
				t.Fail()
			}
		}
	}
}

func BenchmarkHumanSolve(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grid := NewGrid()
		grid.Load(TEST_GRID)
		grid.HumanSolve()
		grid.Done()
	}
}

func TestHumanSolve(t *testing.T) {
	grid := NewGrid()
	grid.Load(TEST_GRID)

	steps := grid.HumanSolution()

	if steps == nil {
		t.Log("Human solution returned 0 techniques.")
		t.Fail()
	}

	if grid.Solved() {
		t.Log("Human Solutions mutated the grid.")
		t.Fail()
	}

	steps = grid.HumanSolve()
	//TODO: test to make sure that we use a wealth of different techniques. This will require a cooked random for testing.
	if steps == nil {
		t.Log("Human solve returned 0 techniques")
		t.Fail()
	}
	if !grid.Solved() {
		t.Log("Human solve failed to solve the simple grid.")
		t.Fail()
	}

	grid.Done()

}

func TestStepsDescription(t *testing.T) {

	grid := NewGrid()

	steps := SolveDirections{
		&SolveStep{
			CellList{
				grid.Cell(0, 0),
			},
			nil,
			IntSlice{1},
			Techniques[3],
		},
		&SolveStep{
			CellList{
				grid.Cell(1, 0),
				grid.Cell(1, 1),
			},
			CellList{
				grid.Cell(1, 3),
				grid.Cell(1, 4),
			},
			IntSlice{1, 2},
			Techniques[5],
		},
		&SolveStep{
			CellList{
				grid.Cell(2, 0),
			},
			nil,
			IntSlice{2},
			Techniques[3],
		},
	}

	descriptions := steps.Description()

	GOLDEN_DESCRIPTIONS := []string{
		"First, we put 1 in cell (0,0) because 1 is the only remaining valid number for that cell.",
		"Next, we remove the possibilities 1 and 2 from cells (1,0) and (1,1) because 1 is only possible in column 0 of block 1, which means it can't be in any other cell in that column not in that block.",
		"Finally, we put 2 in cell (2,0) because 2 is the only remaining valid number for that cell.",
	}

	for i := 0; i < len(GOLDEN_DESCRIPTIONS); i++ {
		if descriptions[i] != GOLDEN_DESCRIPTIONS[i] {
			t.Log("Got wrong human solve description: ", descriptions[i])
			t.Fail()
		}
	}
}

func TestPuzzleDifficulty(t *testing.T) {
	grid := NewGrid()
	grid.Load(TEST_GRID)

	difficulty := grid.Difficulty()

	if grid.Solved() {
		t.Log("Difficulty shouldn't have changed the underlying grid, but it did.")
		t.Fail()
	}

	if difficulty < 0.0 || difficulty > 1.0 {
		t.Log("The grid's difficulty was outside of allowed bounds.")
		t.Fail()
	}

	grid.Done()

	puzzleFilenames := []string{"harddifficulty.sdk", "harddifficulty2.sdk"}

	for _, filename := range puzzleFilenames {
		puzzleDifficultyHelper(filename, t)
	}
}

func puzzleDifficultyHelper(filename string, t *testing.T) {
	otherGrid := NewGrid()
	if !otherGrid.LoadFromFile(puzzlePath(filename)) {
		t.Log("Whoops, couldn't load the file to test:", filename)
		t.Fail()
	}

	after := time.After(time.Second * 5)

	done := make(chan bool)

	go func() {
		_ = otherGrid.Difficulty()
		done <- true
	}()

	select {
	case <-done:
		//totally fine.
	case <-after:
		//Uh oh.
		t.Log("We never finished solving the hard difficulty puzzle: ", filename)
		t.Fail()
	}
}
