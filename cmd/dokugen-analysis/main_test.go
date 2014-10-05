package main

import (
	"testing"
)

func TestPuzzleConversion(t *testing.T) {

	config := [][]string{
		{"7!,2,8!,4!,6!,9,1,5,3;9,1,5!,7,3,8!,6,4,2!;6,4,3,5!,2,1,9!,7,8;8!,7,9,2!,4!,6,5!,3!,1;5!,6!,2,9,1,3!,7,8,4;1,3!,4,8,5,7!,2!,6,9!;2!,5!,6!,3,9,4!,8!,1,7!;3,9!,7,1!,8,5!,4!,2!,6;4!,8,1,6!,7,2!,3,9!,5!",
			`7.846....
..5..8..2
...5..9..
8..24.53.
56...3...
.3...72.9
256..48.7
.9.1.542.
4..6.2.95`},
	}

	for i, line := range config {
		if convertPuzzleString(line[0]) != line[1] {
			t.Error("For row ", i, " expected ", line[1], " got ", line[0])
		}
	}

}