package sudoku

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

var sampleSolveDirections SolveDirections

func init() {
	sampleSolveDirections = SolveDirections{
		nil,
		[]*CompoundSolveStep{
			&CompoundSolveStep{
				FillStep: &SolveStep{
					techniquesByName["Necessary In Row"],
					nil,
					nil,
					nil,
					nil,
					nil,
				},
			},
			&CompoundSolveStep{
				FillStep: &SolveStep{
					techniquesByName["Guess"],
					nil,
					nil,
					nil,
					nil,
					nil,
				},
			},
			&CompoundSolveStep{
				PrecursorSteps: []*SolveStep{
					&SolveStep{
						techniquesByName["Naked Pair Block"],
						nil,
						nil,
						nil,
						nil,
						nil,
					},
				},
				FillStep: &SolveStep{
					techniquesByName["Guess"],
					nil,
					nil,
					nil,
					nil,
					nil,
				},
			},
		},
	}
}

func TestSetDifficultyModel(t *testing.T) {
	currentModel := difficultySignalWeights
	currentModelHash := DifficultyModelHash()
	newModel := map[string]float64{
		"a": 0.5,
		"b": 0.6,
	}
	newModelExpectedHash := "2EE69110CFF4112E325A099D3B83DA0F7B2DC737"

	LoadDifficultyModel(newModel)
	if !reflect.DeepEqual(difficultySignalWeights, newModel) {
		t.Error("LoadDifficultyModel didn't set model")
	}
	if currentModelHash == DifficultyModelHash() {
		t.Error("Model hash didn't change when setting new model")
	}
	if DifficultyModelHash() != newModelExpectedHash {
		t.Error("Model hash wasn't what was expected. Got", DifficultyModelHash(), "expected", newModelExpectedHash)
	}

	LoadDifficultyModel(currentModel)

	if DifficultyModelHash() != currentModelHash {
		t.Error("Setting back to old model didn't set back to old hash")
	}
}

func TestHintDirections(t *testing.T) {

	grid := NewGrid()

	shortSolveDirections := SolveDirections{
		grid,
		[]*CompoundSolveStep{
			{
				FillStep: &SolveStep{
					techniquesByName["Necessary In Row"],
					[]*Cell{
						grid.Cell(2, 3),
					},
					IntSlice{4},
					nil,
					nil,
					nil,
				},
			},
		},
	}

	descriptions := strings.Join(shortSolveDirections.Description(), " ")

	shortGolden := "Based on the other numbers you've entered, (2,3) can only be a 4. How do we know that? We put 4 in cell (2,3) because 4 is required in the 2 row, and 3 is the only column it fits."

	if descriptions != shortGolden {
		t.Error("Got wrong description for hint. Got", descriptions, "wanted", shortGolden)
	}

	multiStepSolveDirections := SolveDirections{
		grid,
		[]*CompoundSolveStep{
			{
				PrecursorSteps: []*SolveStep{
					{
						techniquesByName["Naked Pair Block"],
						[]*Cell{
							grid.Cell(2, 3),
							grid.Cell(2, 3),
						},
						IntSlice{4, 5},
						[]*Cell{
							grid.Cell(3, 4),
							grid.Cell(4, 5),
						},
						IntSlice{2, 3},
						nil,
					},
					{
						techniquesByName["Naked Pair Block"],
						[]*Cell{
							grid.Cell(3, 2),
							grid.Cell(3, 2),
						},
						IntSlice{4, 5},
						[]*Cell{
							grid.Cell(5, 5),
							grid.Cell(4, 6),
						},
						IntSlice{2, 3},
						nil,
					},
				},
				FillStep: &SolveStep{
					techniquesByName["Necessary In Row"],
					[]*Cell{
						grid.Cell(2, 3),
					},
					IntSlice{4},
					nil,
					nil,
					nil,
				},
			},
		},
	}

	descriptions = strings.Join(multiStepSolveDirections.Description(), " ")

	longGolden := "Based on the other numbers you've entered, (2,3) can only be a 4. How do we know that? We can't fill any cells right away so first we need to cull some possibilities. First, we remove the possibilities 4 and 5 from cells (2,3) and (2,3) because 4 and 5 are only possible in (3,4) and (4,5), which means that they can't be in any other cell in block 1. Next, we remove the possibilities 4 and 5 from cells (3,2) and (3,2) because 4 and 5 are only possible in (5,5) and (4,6), which means that they can't be in any other cell in block 3. Finally, we put 4 in cell (2,3) because 4 is required in the 2 row, and 3 is the only column it fits."

	if descriptions != longGolden {
		t.Error("Got wrong description for hint. Got", descriptions, "wanted", longGolden)
	}
}

//TODO: the other solvedirections tests should be in this file.

func TestDifficultySignals(t *testing.T) {
	signals := DifficultySignals{"a": 1.0, "b": 5.0}
	other := DifficultySignals{"a": 3.2, "c": 6.0}

	golden := DifficultySignals{"a": 3.2, "b": 5.0, "c": 6.0}
	signals.add(other)

	if !reflect.DeepEqual(signals, golden) {
		t.Error("Signals when added didn't have right values. Got", signals, "expected", golden)
	}
}

func TestSumDifficultySignals(t *testing.T) {
	signals := DifficultySignals{
		"a": 0.5,
		"b": 1.0,
		"c": 0.6,
	}
	other := DifficultySignals{
		"b": 2.0,
		"d": 1.0,
	}
	golden := DifficultySignals{
		"a": 0.5,
		"b": 3.0,
		"c": 0.6,
		"d": 1.0,
	}
	signals.sum(other)
	if !reflect.DeepEqual(signals, golden) {
		t.Error("Signas when summed didn't have right values. Got", signals, "expected", golden)
	}
}

func TestSolveDirectionsSignals(t *testing.T) {
	result := sampleSolveDirections.Signals()
	golden := DifficultySignals{}

	for _, techniqueName := range AllTechniqueVariants {
		golden[techniqueName+" Count"] = 0.0
		golden[techniqueName+" Percentage"] = 0.0
	}
	golden["Guess Count"] = 2.0
	golden["Necessary In Row Count"] = 1.0
	golden["Naked Pair Block Count"] = 1.0
	golden["Number of Steps"] = 4.0
	golden["Percentage Fill Steps"] = 0.75
	golden["Number Unfilled Cells"] = 3.0
	golden["Steps Until Nonfill"] = 2.0
	golden["Guess Percentage"] = 0.5
	golden["Necessary In Row Percentage"] = 0.25
	golden["Naked Pair Block Percentage"] = 0.25

	if !reflect.DeepEqual(result, golden) {
		t.Error("SolveDirections.Signals on sampleSolveDirections didn't return right value. Got: ", result, " expected: ", golden)
	}

	//We're going to swap out the real difficulty signal weights for the test.
	realWeights := difficultySignalWeights
	defer func() {
		difficultySignalWeights = realWeights
	}()

	difficultySignalWeights = map[string]float64{
		"Constant":               0.5,
		"Guess Count":            -0.09,
		"Necessary In Row Count": -0.07,
		"Naked Pair Block Count": 0.13,
		"Number of Steps":        0.11,
	}

	difficulty := result.difficulty()
	expectedDifficulty := 0.82

	if math.Abs(difficulty-expectedDifficulty) > 0.00000000001 {
		t.Error("Got wrong difficulty from baked signals: ", difficulty, "expected", 0.82)
	}

}

func TestTechniqueSignal(t *testing.T) {

	result := signalTechnique(sampleSolveDirections)

	golden := DifficultySignals{}

	for _, techniqueName := range AllTechniqueVariants {
		golden[techniqueName+" Count"] = 0.0
	}

	golden["Guess Count"] = 2.0
	golden["Necessary In Row Count"] = 1.0
	golden["Naked Pair Block Count"] = 1.0

	if !reflect.DeepEqual(result, golden) {
		t.Error("Technique signal didn't work as expected. Got", result, "expected", golden)
	}
}

func TestTechniqueSignalPercentage(t *testing.T) {

	result := signalTechniquePercentage(sampleSolveDirections)

	golden := DifficultySignals{}

	for _, techniqueName := range AllTechniqueVariants {
		golden[techniqueName+" Percentage"] = 0.0
	}

	golden["Guess Percentage"] = 0.5
	golden["Necessary In Row Percentage"] = 0.25
	golden["Naked Pair Block Percentage"] = 0.25

	if !reflect.DeepEqual(result, golden) {
		t.Error("Technique signal percentage didn't work as expected. Got", result, "expected", golden)
	}
}

func TestSignalNumberOfSteps(t *testing.T) {
	result := signalNumberOfSteps(sampleSolveDirections)
	golden := DifficultySignals{
		"Number of Steps": 4.0,
	}
	if !reflect.DeepEqual(result, golden) {
		t.Error("Number of steps signal didn't work as expected. Got ", result, "expected", golden)
	}
}

func TestSignalPercentageFillSteps(t *testing.T) {
	result := signalPercentageFilledSteps(sampleSolveDirections)
	golden := DifficultySignals{
		"Percentage Fill Steps": 0.75,
	}
	if !reflect.DeepEqual(result, golden) {
		t.Error("Percentage fill steps didn't work as expected. Got ", result, " expected ", golden)
	}
}

func TestSignalNumberUnfilled(t *testing.T) {
	result := signalNumberUnfilled(sampleSolveDirections)
	golden := DifficultySignals{
		"Number Unfilled Cells": 3.0,
	}
	if !reflect.DeepEqual(result, golden) {
		t.Error("Number unfilled cells didn't work as expected. Got", result, "expected ", golden)
	}

}

func TestSignalStepsUntilNonFill(t *testing.T) {
	result := signalStepsUntilNonFill(sampleSolveDirections)
	golden := DifficultySignals{
		"Steps Until Nonfill": 2.0,
	}
	if !reflect.DeepEqual(result, golden) {
		t.Error("Steps until nonfill didn't work as expected. Got ", result, "expected", golden)
	}
}
