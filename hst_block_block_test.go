package sudoku

import (
	"testing"
)

//TODO: test a few more block/block puzzles to make sure I'm handling all cases right.

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

func TestBlockBlockInteraction(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{6, 3}, {6, 4}, {7, 3}, {7, 4}, {8, 3}, {8, 4}},
		pointerCells: []cellRef{{2, 3}, {2, 4}, {3, 4}, {5, 3}},
		targetSame:   GROUP_BLOCK,
		targetGroup:  7,
		targetNums:   IntSlice([]int{3}),
		description:  "3 can only be in two different columns in blocks 1 and 4, which means that 3 can't be in any other cells in those columns that aren't in blocks 1 and 4",
	}
	humanSolveTechniqueTestHelper(t, "blockblocktest.sdk", "Block Block Interactions", options)

}

func TestBlockBlockInteractionAgain(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		targetCells:  []cellRef{{3, 3}, {3, 4}, {3, 5}, {5, 3}, {5, 4}, {5, 5}},
		pointerCells: []cellRef{{3, 0}, {3, 1}, {3, 6}, {3, 7}, {5, 0}, {5, 1}, {5, 6}, {5, 7}},
		targetSame:   GROUP_BLOCK,
		targetGroup:  4,
		targetNums:   IntSlice([]int{2}),
		description:  "2 can only be in two different rows in blocks 3 and 5, which means that 2 can't be in any other cells in those rows that aren't in blocks 3 and 5",
	}
	humanSolveTechniqueTestHelper(t, "blockblocktest1.sdk", "Block Block Interactions", options)
}

func TestBlockInteractionFlipped(t *testing.T) {
	options := solveTechniqueTestHelperOptions{
		transpose:    true,
		targetCells:  []cellRef{{3, 6}, {4, 6}, {3, 7}, {4, 7}, {3, 8}, {4, 8}},
		pointerCells: []cellRef{{3, 2}, {4, 2}, {4, 3}, {3, 5}},
		targetSame:   GROUP_BLOCK,
		targetGroup:  5,
		targetNums:   IntSlice([]int{3}),
		description:  "3 can only be in two different rows in blocks 3 and 4, which means that 3 can't be in any other cells in those rows that aren't in blocks 3 and 4",
	}
	humanSolveTechniqueTestHelper(t, "blockblocktest.sdk", "Block Block Interactions", options)
}
