package sudoku

import (
	"math"
	"reflect"
	"testing"
)

func TestDefaultProbabilityDistributionTweak(t *testing.T) {
	result := defaultProbabilityDistributionTweak(3)
	expected := probabilityDistributionTweak{
		1.0,
		1.0,
		1.0,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Error("Got wrong result from defaultProbabilityDistributionTweak. Got", result, "expected", expected)
	}

}

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

	expected := probabilityDistributionTweak{
		2.0,
		1.0,
		5.0,
		1.0,
	}
	result := twiddleCommonNumbers(possibilities, grid, nil)

	if !reflect.DeepEqual(result, expected) {
		t.Error("Got wrong result. Wanted:", expected, "got", result)
	}

}

func TestTwiddleChainedSteps(t *testing.T) {
	//TODO: test other, harder cases as well.
	grid := NewGrid()
	lastStep := &SolveStep{
		nil,
		cellRefsToCells([]cellRef{
			{0, 0},
		}, grid),
		nil,
		nil,
		nil,
		nil,
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

	expected := []float64{
		3.727593720314952e+08,
		517947.4679231202,
		1.0,
	}

	results := twiddleChainedSteps(possibilities, grid, lastStep.TargetCells)

	lastWeight := math.MaxFloat64
	for i, weight := range results {
		if weight >= lastWeight {
			t.Error("Tweak Chained Steps Weights didn't tweak things in the right direction: ", results, "at", i)
		}
		lastWeight = weight
	}

	for i, weight := range results {
		if math.Abs(expected[i]-weight) > 0.00001 {
			t.Error("Index", i, "was different than expected. Got", weight, "wanted", expected[i])
		}
	}
}
