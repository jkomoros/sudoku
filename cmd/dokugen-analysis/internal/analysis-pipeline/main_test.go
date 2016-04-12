package main

import (
	"flag"
	"testing"
)

func TestPhaseToString(t *testing.T) {
	var strToPhaseTests = []struct {
		str      string
		expected Phase
	}{
		{"DiFFiculties", Difficulties},
		{"Difficulties", Difficulties},
		{"Foo", -1},
		{"solves", Solves},
		{"analysis", Analysis},
		{"histogram", Histogram},
	}

	for _, test := range strToPhaseTests {
		result := StringToPhase(test.str)
		if result != test.expected {
			t.Error("Test got back wrong StringToPhase.Got", result, "expected:", test.expected)
		}
	}

	var phasetoStringTests = []struct {
		p        Phase
		expected string
	}{
		{Difficulties, "difficulties"},
		{Solves, "solves"},
		{Analysis, "analysis"},
		{Histogram, "histogram"},
		{Phase(len(phaseMap)), ""},
		{-1, ""},
	}

	for _, test := range phasetoStringTests {
		result := test.p.String()
		if result != test.expected {
			t.Error("Test got back wrong phasetoString. Got", result, "Expected:", test.expected)
		}
	}
}

func TestAppOptionsPhase(t *testing.T) {

	options := getDefaultOptions()

	if err := options.fixUp(); err != nil {
		t.Error("Got non-nil error on basic options", err)
	}

	if options.start != Solves || options.end != Analysis {
		t.Error("Start or end defaulted to wrong things when empty:", options.start, options.end)
	}

	options = getDefaultOptions()

	options.rawStart = "analysis"
	options.rawEnd = "histogram"

	if err := options.fixUp(); err != nil {
		t.Error("Got non-nil error on basic options", err)
	}

	if options.start != Analysis {
		t.Error("Expected options.start to be analysis, got", options.start)
	}

	if options.end != Histogram {
		t.Error("Expected options.end to be histogram, got", options.end)
	}

	options = getDefaultOptions()

	options.rawStart = "histogram"
	options.rawEnd = "analysis"

	if err := options.fixUp(); err == nil {
		t.Error("Didn't get error for a start phase that's after an end phase")
	}

	options = getDefaultOptions()

	options.rawStart = "foo"

	if err := options.fixUp(); err == nil {
		t.Error("Didn't get an error for an invalid start phase")
	}

	options = getDefaultOptions()

	options.rawStart = "histogram"

	if err := options.fixUp(); err != nil {
		t.Error("Got unexpected error during fixup: ", err)
	}

	if options.start != Histogram && options.end != Histogram {
		t.Error("Expected start and end to be histogram")
	}

}

//Callers should call fixUpOptions after receiving this.
func getDefaultOptions() *appOptions {
	options := newAppOptions(flag.NewFlagSet("main", flag.ExitOnError))
	options.flagSet.Parse([]string{})
	return options
}
