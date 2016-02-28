package sudoku

import (
	"math"
	"testing"
)

func TestTweakChainedStepsWeights(t *testing.T) {

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

	results := tweakChainedStepsWeights(possibilities, lastStep.TargetCells)

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
