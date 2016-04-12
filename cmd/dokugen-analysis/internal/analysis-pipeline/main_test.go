package main

import (
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
