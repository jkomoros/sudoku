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
	golden := DifficultySignals{
		"Constant":               1.0,
		"Guess Count":            2.0,
		"Naked Pair Block Count": 1.0,
	}
	if !reflect.DeepEqual(result, golden) {
		t.Error("SolveDirections.Signals on sampleSolveDirections didn't return right value. Got: ", result, " expected: ", golden)
	}
}

func TestTechniqueSignal(t *testing.T) {

	result := signalTechnique(sampleSolveDirections)

	golden := DifficultySignals{
		"Guess Count":            2.0,
		"Naked Pair Block Count": 1.0,
	}

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
