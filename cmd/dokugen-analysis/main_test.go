package main

import (
	"github.com/jkomoros/sudoku"
	"math"
	"reflect"
	"testing"
)

/*
type puzzle struct {
	id                     int
	userRelativeDifficulty float64
	difficultyRating       int
	name                   string
	puzzle                 string
}

*/

func defaultPuzzleSet() []*puzzle {
	return []*puzzle{
		&puzzle{
			userRelativeDifficulty: 0.001,
		},
		&puzzle{
			userRelativeDifficulty: 0.01,
		},
		&puzzle{
			userRelativeDifficulty: 0.1,
		},
		&puzzle{
			userRelativeDifficulty: 0.9,
		},
	}
}

func TestSkewAmount(t *testing.T) {
	puzzles := defaultPuzzleSet()
	tests := []struct {
		pow      float64
		expected float64
	}{
		{10.0, -0.04281493061016942},
		{5.0, -0.039248270833675214},
		{3.0, 0.2088930756577995},
	}

	for _, tt := range tests {
		actual := skewAmount(puzzles, tt.pow)
		const epsilon = 1e-10
		if math.Abs(actual-tt.expected) > epsilon {
			t.Error("SkewAmount(", tt.pow, "): expected", tt.expected, ", got:", actual)
		}
	}
}

func TestTrimTails(t *testing.T) {

	puzzles := []*puzzle{
		&puzzle{
			userRelativeDifficulty: 0.1,
		},
		&puzzle{
			userRelativeDifficulty: 0.2,
		},
		&puzzle{
			userRelativeDifficulty: 0.3,
		},
		&puzzle{
			userRelativeDifficulty: 0.4,
		},
		&puzzle{
			userRelativeDifficulty: 0.5,
		},
		&puzzle{
			userRelativeDifficulty: 0.6,
		},
		&puzzle{
			userRelativeDifficulty: 0.7,
		},
		&puzzle{
			userRelativeDifficulty: 0.8,
		},
		&puzzle{
			userRelativeDifficulty: 0.9,
		},
		&puzzle{
			userRelativeDifficulty: 1.0,
		},
	}

	expected := []*puzzle{
		&puzzle{
			userRelativeDifficulty: 0.2,
		},
		&puzzle{
			userRelativeDifficulty: 0.2,
		},
		&puzzle{
			userRelativeDifficulty: 0.3,
		},
		&puzzle{
			userRelativeDifficulty: 0.4,
		},
		&puzzle{
			userRelativeDifficulty: 0.5,
		},
		&puzzle{
			userRelativeDifficulty: 0.6,
		},
		&puzzle{
			userRelativeDifficulty: 0.7,
		},
		&puzzle{
			userRelativeDifficulty: 0.8,
		},
		&puzzle{
			userRelativeDifficulty: 0.9,
		},
		&puzzle{
			userRelativeDifficulty: 0.9,
		},
	}

	trimTails(puzzles, 0.20)

	for i, puzz := range puzzles {
		diff := puzz.userRelativeDifficulty
		expectedDiff := expected[i].userRelativeDifficulty
		if diff != expectedDiff {
			t.Error("TrimTails discrepancy found at", i, "expected", expectedDiff, "got", diff)
		}
	}

}

func TestBisectPower(t *testing.T) {
	puzzles := defaultPuzzleSet()
	actual := bisectPower(puzzles)
	expected := 3.9013671875
	if actual != expected {
		t.Error("BisectPower wrong. expected", expected, "got", actual)
	}
}

func TestPuzzleConversion(t *testing.T) {

	//TODO: fix up this test now that convertPuzzleString no longer exists.
	//... We might need to move everything over to user sdkconverter.

	// 	config := [][]string{
	// 		{"7!,2,8!,4!,6!,9,1,5,3;9,1,5!,7,3,8!,6,4,2!;6,4,3,5!,2,1,9!,7,8;8!,7,9,2!,4!,6,5!,3!,1;5!,6!,2,9,1,3!,7,8,4;1,3!,4,8,5,7!,2!,6,9!;2!,5!,6!,3,9,4!,8!,1,7!;3,9!,7,1!,8,5!,4!,2!,6;4!,8,1,6!,7,2!,3,9!,5!",
	// 			`7.846....
	// ..5..8..2
	// ...5..9..
	// 8..24.53.
	// 56...3...
	// .3...72.9
	// 256..48.7
	// .9.1.542.
	// 4..6.2.95`},
	// 	}

	// 	for i, line := range config {
	// 		if convertPuzzleString(line[0]) != line[1] {
	// 			t.Error("For row ", i, " expected ", line[1], " got ", line[0])
	// 		}
	// 	}

}

func TestRemoveZeroedFloats(t *testing.T) {
	input := [][][]float64{
		{
			{0.0, 1.0, 1.0, 0.0},
			{1.0, 0.0, 0.0, 0.0},
			{0.0, 0.0, 0.0, 0.0},
		},
		{
			{1.0, 0.5, 1.0},
			{0.0, 0.0, 0.0},
		},
		{
			{1.0, 0.0, 0.5},
			{0.0, 0.0, 0.0},
		},
		{
			{1.0, 0.0, 2.0, 0.0, 3.0},
			{3.0, 0.0, 2.0, 0.0, 1.0},
		},
		{
			{1.0, 0.0, 2.0, 0.0, 3.0},
			{3.0, 0.0, 2.0, 0.0, 1.0},
		},
	}
	expected := [][][]float64{
		{
			{0.0, 1.0, 1.0},
			{1.0, 0.0, 0.0},
			{0.0, 0.0, 0.0},
		},
		{
			{1.0, 0.5, 1.0},
			{0.0, 0.0, 0.0},
		},
		{
			{1.0, 0.5},
			{0.0, 0.0},
		},
		{
			{1.0, 2.0, 3.0},
			{3.0, 2.0, 1.0},
		},
		{
			{1.0, 2.0, 0.0, 3.0},
			{3.0, 2.0, 0.0, 1.0},
		},
	}
	safeIndexes := [][]int{
		nil,
		nil,
		nil,
		nil,
		{3},
	}
	keptIndexes := [][]int{
		{0, 1, 2},
		{0, 1, 2},
		{0, 2},
		{0, 2, 4},
		{0, 2, 3, 4},
	}
	for i, test := range input {
		expect := expected[i]
		safe := safeIndexes[i]
		expectedIndexes := sudoku.IntSlice(keptIndexes[i])
		result, indexes := removeZeroedColumns(test, safe)
		if !reflect.DeepEqual(expect, result) {
			t.Error("Didn't equal:", result)
		}
		if !expectedIndexes.SameContentAs(sudoku.IntSlice(indexes)) {
			t.Error("Wrong indexes: ", indexes)
		}
	}

}
