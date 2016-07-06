package sudoku

import (
	"math"
	"testing"
)

func TestTwiddleHumanLikelihood(t *testing.T) {
	possibilities := []*SolveStep{
		{
			Technique: techniquesByName["Only Legal Number"],
		},
		{
			Technique: techniquesByName["Guess"],
		},
	}

	for i, step := range possibilities {
		result := twiddleHumanLikelihood(step, nil, nil, nil)
		expected := probabilityTweak(step.HumanLikelihood())

		if result != expected {
			t.Error("Got wrong twiddle for human likelihood at index", i, "got", result, "expected", expected)
		}
	}
}

func TestTwiddlePointingTargetOverlap(t *testing.T) {
	grid := NewGrid()
	tests := []struct {
		lastStep    *SolveStep
		currentStep *SolveStep
		expected    probabilityTweak
		description string
	}{
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				PointerCells: grid.Row(0),
			},
			1.000001,
			"Full pointer/cell overlap",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			1.01,
			"Full target/target overlap",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: CellSlice{grid.Cell(0, 0)},
			},
			1.770493827160494,
			"Single cell out of 9",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0).Intersection(grid.Block(0)),
			},
			&SolveStep{
				TargetCells: CellSlice{grid.Cell(0, 0)},
			},
			1.4011111111111112,
			"Single cell out of three",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: grid.Row(7),
			},
			2.0,
			"Two rows no overlap",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0).Intersection(grid.Block(0)),
			},
			&SolveStep{
				TargetCells: grid.Row(DIM - 1).Intersection(grid.Block(DIM - 1)),
			},
			2.0,
			"Two three-cell rows no overlap",
		},
		{
			&SolveStep{
				TargetCells: CellSlice{grid.Cell(0, 0)},
			},
			&SolveStep{
				TargetCells: CellSlice{grid.Cell(0, 0)},
			},
			1.01,
			"Two individual cells overlapping",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: grid.Col(0),
			},
			1.8747750865051902,
			"Row and col intersecting at one point",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: grid.Block(0),
			},
			1.6084,
			"First row and first block overlapping",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0).Intersection(grid.Block(0)),
			},
			&SolveStep{
				TargetCells: grid.Block(0),
			},
			1.4011111111111112,
			"First three cells and first block overlapping",
		},
	}

	for i, test := range tests {
		result := twiddlePointingTargetOverlap(test.currentStep, []*SolveStep{test.lastStep}, nil, grid)
		if math.IsNaN(float64(result)) {
			t.Error("Got NaN on test", i, test.description)
		}
		if math.Abs(float64(result-test.expected)) > 0.000001 {
			t.Error("Test", i, "got wrong result. Got", result, "expected", test.expected, test.description)
		}
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

	expected := []probabilityTweak{
		7.0,
		1.0,
		4.0,
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
		1.3894954943731375,
		2.6826957952797263,
		10.0,
	}

	lastResult := probabilityTweak(math.SmallestNonzeroFloat64)

	for i, step := range possibilities {
		result := twiddleChainedSteps(step, lastStep, nil, grid)
		expectedResult := expected[i]

		if result <= lastResult {
			t.Error("Tweak Chained Steps Weights didn't tweak things in the right direction: ", result, "at", i)
		}
		lastResult = result

		if math.Abs(float64(expectedResult-result)) > 0.00001 {
			t.Error("Twiddle chained steps at", i, "got", result, "expected", expectedResult)
		}

	}

}
