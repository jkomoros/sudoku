package sudoku

import (
	"reflect"
	"testing"
)

var sampleSolveDirections SolveDirections

func init() {
	sampleSolveDirections = SolveDirections{
		&SolveStep{
			nil,
			nil,
			nil,
			nil,
			techniquesByName["Necessary In Row"],
		},
		&SolveStep{
			nil,
			nil,
			nil,
			nil,
			techniquesByName["Guess"],
		},
		&SolveStep{
			nil,
			nil,
			nil,
			nil,
			techniquesByName["Naked Pair Block"],
		},
		&SolveStep{
			nil,
			nil,
			nil,
			nil,
			techniquesByName["Guess"],
		},
	}
}

//TODO: the other solvedirections tests should be in this file.

func TestDifficultySignals(t *testing.T) {
	signals := DifficultySignals{"a": 1.0, "b": 5.0}
	other := DifficultySignals{"a": 3.2, "c": 6.0}

	golden := DifficultySignals{"a": 3.2, "b": 5.0, "c": 6.0}
	signals.Add(other)

	if !reflect.DeepEqual(signals, golden) {
		t.Error("Signals when added didn't have right values. Got", signals, "expected", golden)
	}
}

func TestSolveDirectionsSignals(t *testing.T) {
	result := sampleSolveDirections.Signals()
	golden := DifficultySignals{}

	for _, technique := range AllTechniques {
		golden[technique.Name()+" Count"] = 0.0
	}

	golden["Constant"] = 1.0
	golden["Guess Count"] = 2.0
	golden["Necessary In Row Count"] = 1.0
	golden["Naked Pair Block Count"] = 1.0
	golden["Number of Steps"] = 4.0
	golden["Percentage Fill Steps"] = 0.75

	if !reflect.DeepEqual(result, golden) {
		t.Error("SolveDirections.Signals on sampleSolveDirections didn't return right value. Got: ", result, " expected: ", golden)
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

func TestConstantSignal(t *testing.T) {
	result := signalConstant(sampleSolveDirections)
	golden := DifficultySignals{
		"Constant": 1.0,
	}
	if !reflect.DeepEqual(result, golden) {
		t.Error("Constant signal didn't work as expected. Got", result, "expected", golden)
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
