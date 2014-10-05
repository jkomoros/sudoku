package main

import (
	"testing"
)

func TestPuzzleConversion(t *testing.T) {

	config := [][]string{
		{"abc", "abc"},
		{"def", "def"},
	}

	for i, line := range config {
		if convertPuzzleString(line[0]) != line[1] {
			t.Error("For row ", i, " expected ", line[1], " got ", line[0])
		}
	}

}
