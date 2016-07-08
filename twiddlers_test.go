package sudoku

import (
	"math"
	"testing"
)

func TestTwiddleTwiddle(t *testing.T) {

	twiddlerFMaker := func(returnValue float64) probabilityTwiddler {
		return func(proposedStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, previousGrid *Grid) probabilityTweak {
			return probabilityTweak(returnValue)
		}
	}

	tests := []struct {
		item        *probabilityTwiddlerItem
		expected    probabilityTweak
		description string
	}{
		{
			&probabilityTwiddlerItem{
				f:      twiddlerFMaker(10.0),
				weight: -1.0,
			},
			10.0,
			"Large value negative weight",
		},
		{
			&probabilityTwiddlerItem{
				f:      twiddlerFMaker(0.0),
				weight: 4.0,
			},
			0.0,
			"Normal with 0 return and normal weight",
		},
		{
			&probabilityTwiddlerItem{
				f:      twiddlerFMaker(0.5),
				weight: 4.0,
			},
			2.0,
			"Normal with 0.5 return 4.0 weight",
		},
	}

	for i, test := range tests {
		result := test.item.Twiddle(nil, nil, nil, nil)
		if result != test.expected {
			t.Error("Got wrong result for item", i, test.description, "Got", result, "Expected", test.expected)
		}
	}
}

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
			0.000001,
			"Full pointer/cell overlap",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			0.010000000000000018,
			"Full target/target overlap",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: CellSlice{grid.Cell(0, 0)},
			},
			0.7704938271604939,
			"Single cell out of 9",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0).Intersection(grid.Block(0)),
			},
			&SolveStep{
				TargetCells: CellSlice{grid.Cell(0, 0)},
			},
			0.4011111111111111,
			"Single cell out of three",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: grid.Row(7),
			},
			1.0,
			"Two rows no overlap",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0).Intersection(grid.Block(0)),
			},
			&SolveStep{
				TargetCells: grid.Row(DIM - 1).Intersection(grid.Block(DIM - 1)),
			},
			1.0,
			"Two three-cell rows no overlap",
		},
		{
			&SolveStep{
				TargetCells: CellSlice{grid.Cell(0, 0)},
			},
			&SolveStep{
				TargetCells: CellSlice{grid.Cell(0, 0)},
			},
			0.010000000000000018,
			"Two individual cells overlapping",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: grid.Col(0),
			},
			0.8747750865051903,
			"Row and col intersecting at one point",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0),
			},
			&SolveStep{
				TargetCells: grid.Block(0),
			},
			0.6084,
			"First row and first block overlapping",
		},
		{
			&SolveStep{
				TargetCells: grid.Row(0).Intersection(grid.Block(0)),
			},
			&SolveStep{
				TargetCells: grid.Block(0),
			},
			0.4011111111111111,
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
		0.7777777777777778,
		0.0,
		0.4444444444444444,
		0.0,
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
		0.12316184623065915,
		0.22758459260747887,
		1.0,
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

func TestTwiddlePreferFilledGroups(t *testing.T) {
	grid := NewGrid()

	keyCell := grid.Cell(0, 0)

	step := &SolveStep{
		TargetCells: CellSlice{
			keyCell,
		},
		Technique: techniquesByName["Only Legal Number"],
	}

	//TODO: instead of testing for the exact values, just make sure that every
	//value is lower than the one before it.

	//TODO: more exhaustive tests

	testHelper := func(expected probabilityTweak, description string) {
		result := twiddlePreferFilledGroups(step, nil, nil, grid)
		if result != expected {
			t.Error("Got wrong result for", description, "Got", result, "Expected", expected)
		}
	}

	testHelper(0.7043478260869563, "Completely empty grid")

	//Fill the rest of the block
	for _, cell := range grid.Block(0).RemoveCells(CellSlice{keyCell}) {
		cell.SetNumber(1)
	}

	testHelper(0.28695652173913044, "Full block, empty everything else")

	//Fill the rest of the row, too
	for _, cell := range grid.Row(0).RemoveCells(CellSlice{keyCell}) {
		cell.SetNumber(1)
	}

	testHelper(0.11086956521739132, "Full block and row, otherwise empty col")

	//Fill the rest of the col, too.

	for _, cell := range grid.Col(0).RemoveCells(CellSlice{keyCell}) {
		cell.SetNumber(1)
	}

	testHelper(0.0782608695652174, "Full block, row, col")
}
