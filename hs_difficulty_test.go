package sudoku

import (
	"math"
	"reflect"
	"testing"
)

var sampleSolveDirections SolveDirections

func init() {
	sampleSolveDirections = SolveDirections{
		&SolveStep{
			techniquesByName["Necessary In Row"],
			nil,
			nil,
			nil,
			nil,
		},
		&SolveStep{
			techniquesByName["Guess"],
			nil,
			nil,
			nil,
			nil,
		},
		&SolveStep{
			techniquesByName["Naked Pair Block"],
			nil,
			nil,
			nil,
			nil,
		},
		&SolveStep{
			techniquesByName["Guess"],
			nil,
			nil,
			nil,
			nil,
		},
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

	for _, technique := range AllTechniques {
		golden[technique.Name()+" Count"] = 0.0
		golden[technique.Name()+" Percentage"] = 0.0
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

	for _, technique := range AllTechniques {
		golden[technique.Name()+" Count"] = 0.0
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

	for _, technique := range AllTechniques {
		golden[technique.Name()+" Percentage"] = 0.0
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
