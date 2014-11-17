package sudoku

import (
	"testing"
)

func TestPairwiseBlocks(t *testing.T) {
	grid := NewGrid()
	result := pairwiseBlocks(grid)

	//TODO: should this be encoded as DIM x DIM?
	if len(result) != 18 {
		t.Error("Pairwise blocks had wrong length")
	}

	for i, pair := range result {
		if len(pair) != 2 {
			t.Error(i, "th pair was wrong size: ", pair)
		}
		if pair[0] == pair[1] {
			t.Error(i, "th pair was not two unique cells", pair)
		}
	}
}
