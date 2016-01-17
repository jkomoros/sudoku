package main

import (
	"github.com/jkomoros/sudoku"
	"reflect"
	"testing"
)

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
