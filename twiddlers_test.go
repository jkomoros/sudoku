package sudoku

import (
	"math"
	"testing"
)

func TestTwiddleCommonNumbers(t *testing.T) {

	grid := NewGrid()
	grid.LoadSDK(TEST_GRID)

	//Fill all of the 4's
	solvedGrid := NewGrid()
	solvedGrid.LoadSDK(SOLVED_TEST_GRID)

	for _, cell := range solvedGrid.Cells() {
		if cell.Number() == 4 {
			otherCell := cell.InGrid(grid)
			otherCell.SetNumber(4)
		}
	}

	possibilities := []*SolveStep{
		//Step with 2 filled
		{
			techniquesByName["Only Legal Number"],
			cellRefsToCells([]cellRef{{1, 0}}, grid),
			IntSlice{8},
			nil,
			nil,
			nil,
		},
		//Step with non-fill technique
		{
			techniquesByName["Hidden Pair Block"],
			cellRefsToCells([]cellRef{{1, 0}}, grid),
			IntSlice{8},
			nil,
			nil,
			nil,
		},
		//High valued 1
		{
			techniquesByName["Only Legal Number"],
			cellRefsToCells([]cellRef{{0, 4}}, grid),
			IntSlice{5},
			nil,
			nil,
			nil,
		},
		//Already-filled number
		{
			techniquesByName["Only Legal Number"],
			cellRefsToCells([]cellRef{{0, 4}}, grid),
			IntSlice{4},
			nil,
			nil,
			nil,
		},
	}

	expected := []probabilityTweak{
		2.0,
		1.0,
		5.0,
		1.0,
	}

	for i, step := range possibilities {
		result := twiddleCommonNumbers(step, nil, nil, grid)
		if result != expected[i] {
			t.Error("Twiddle Common Numbers wrong for", i, "got", result, "expected", expected[i])
		}
	}

}

func TestTwiddleChainedSteps(t *testing.T) {
	//TODO: test other, harder cases as well.
	grid := NewGrid()
	lastStep := []*SolveStep{
		{
			nil,
			cellRefsToCells([]cellRef{
				{0, 0},
			}, grid),
			nil,
			nil,
			nil,
			nil,
		},
	}

	possibilities := []*SolveStep{
		{
			nil,
			cellRefsToCells([]cellRef{
				{1, 0},
			}, grid),
			nil,
			nil,
			nil,
			nil,
		},
		{
			nil,
			cellRefsToCells([]cellRef{
				{2, 2},
			}, grid),
			nil,
			nil,
			nil,
			nil,
		},
		{
			nil,
			cellRefsToCells([]cellRef{
				{7, 7},
			}, grid),
			nil,
			nil,
			nil,
			nil,
		},
	}

	expected := []probabilityTweak{
		3.727593720314952e+08,
		517947.4679231202,
		1.0,
	}

	lastResult := probabilityTweak(math.MaxFloat64)

	for i, step := range possibilities {
		result := twiddleChainedSteps(step, lastStep, nil, grid)
		expectedResult := expected[i]

		if result >= lastResult {
			t.Error("Tweak Chained Steps Weights didn't tweak things in the right direction: ", result, "at", i)
		}
		lastResult = result

		if math.Abs(float64(expectedResult-result)) > 0.00001 {
			t.Error("Twiddle chained steps at", i, "got", result, "expected", expectedResult)
		}

	}

}
