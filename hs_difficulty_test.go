package sudoku

import (
	"reflect"
	"testing"
)

func TestDifficultySignals(t *testing.T) {
	signals := DifficultySignals{"a": 1.0, "b": 5.0}
	other := DifficultySignals{"a": 3.2, "c": 6.0}

	golden := DifficultySignals{"a": 3.2, "b": 5.0, "c": 6.0}
	signals.Add(other)

	if !reflect.DeepEqual(signals, golden) {
		t.Error("Signals when added didn't have right values. Got", signals, "expected", golden)
	}
}
