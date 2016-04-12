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
		{"weka", Weka},
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
		{Weka, "weka"},
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

	options.rawStart = "weka"
	options.rawEnd = "histogram"

	if err := options.fixUp(); err != nil {
		t.Error("Got non-nil error on basic options", err)
	}

	if options.start != Weka {
		t.Error("Expected options.start to be weka, got", options.start)
	}

	if options.end != Histogram {
		t.Error("Expected options.end to be weka, got", options.end)
	}

	options = getDefaultOptions()

	options.rawStart = "histogram"
	options.rawEnd = "weka"

	if err := options.fixUp(); err == nil {
		t.Error("Didn't get error for a start phase that's after an end phase")
	}

	options = getDefaultOptions()

	options.rawStart = "foo"

	if err := options.fixUp(); err == nil {
		t.Error("Didn't get an error for an invalid start phase")
	}

}

//Callers should call fixUpOptions after receiving this.
func getDefaultOptions() *appOptions {
	options := newAppOptions(flag.NewFlagSet("main", flag.ExitOnError))
	options.flagSet.Parse([]string{})
	return options
}
