package sudoku

import (
	"fmt"
	"math/rand"
)

type nakedPairCol struct {
	basicSolveTechnique
}

type nakedPairRow struct {
	basicSolveTechnique
}

type nakedPairBlock struct {
	basicSolveTechnique
}

type nakedTripleCol struct {
	basicSolveTechnique
}

type nakedTripleRow struct {
	basicSolveTechnique
}

type nakedTripleBlock struct {
	basicSolveTechnique
}

//TODO: Factor out as much as possible about Description and Find for a class.

func (self nakedPairCol) Description(step *SolveStep) string {
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that they can't be in any other cell in column %d", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Col())
}

func (self nakedPairCol) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will find multiple if they exist.
	colGetter := func(i int) CellList {
		return grid.Col(i)
	}
	return nakedSubset(grid, self, 2, colGetter)
}

func (self nakedPairRow) Description(step *SolveStep) string {
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that they can't be in any other cell in row %d", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Row())
}

func (self nakedPairRow) Find(grid *Grid) []*SolveStep {
	//TODO: test we find multiple if they exist.
	rowGetter := func(i int) CellList {
		return grid.Row(i)
	}
	return nakedSubset(grid, self, 2, rowGetter)
}

func (self nakedPairBlock) Description(step *SolveStep) string {
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that they can't be in any other cell in block %d", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Block())
}

func (self nakedPairBlock) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will return multiple if they exist.
	blockGetter := func(i int) CellList {
		return grid.Block(i)
	}
	return nakedSubset(grid, self, 2, blockGetter)
}

func (self nakedTripleCol) Description(step *SolveStep) string {
	if len(step.Nums) < 3 || len(step.PointerCells) < 3 {
		return ""
	}
	return fmt.Sprintf("%d, %d, and %d are only possible in (%d,%d), (%d,%d) and (%d,%d), which means that they can't be in any other cell in column %d", step.Nums[0], step.Nums[1], step.Nums[2], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.PointerCells[2].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Col())
}

func (self nakedTripleCol) Find(grid *Grid) []*SolveStep {
	//TODO: test we find multiple if they exist.
	colGetter := func(i int) CellList {
		return grid.Col(i)
	}
	return nakedSubset(grid, self, 3, colGetter)
}

func (self nakedTripleRow) Description(step *SolveStep) string {
	if len(step.Nums) < 3 || len(step.PointerCells) < 3 {
		return ""
	}
	return fmt.Sprintf("%d, %d, and %d are only possible in (%d,%d), (%d, %d) and (%d,%d), which means that they can't be in any other cell in row %d", step.Nums[0], step.Nums[1], step.Nums[2], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.PointerCells[2].Row+1, step.PointerCells[2].Col+1, step.TargetCells.Row())
}

func (self nakedTripleRow) Find(grid *Grid) []*SolveStep {
	//TODO: test that if there are multiple we find them.
	rowGetter := func(i int) CellList {
		return grid.Row(i)
	}
	return nakedSubset(grid, self, 3, rowGetter)
}

func (self nakedTripleBlock) Description(step *SolveStep) string {
	if len(step.Nums) < 3 || len(step.PointerCells) < 3 {
		return ""
	}
	return fmt.Sprintf("%d, %d and %d are only possible in (%d,%d), (%d,%d) and (%d,%d), which means that they can't be in any other cell in block %d", step.Nums[0], step.Nums[1], step.Nums[2], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.PointerCells[2].Row+1, step.PointerCells[2].Col+1, step.TargetCells.Block())
}

func (self nakedTripleBlock) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will find multiple ones if they exist.
	blockGetter := func(i int) CellList {
		return grid.Block(i)
	}
	return nakedSubset(grid, self, 3, blockGetter)
}

func nakedSubset(grid *Grid, technique SolveTechnique, k int, collectionGetter func(int) CellList) []*SolveStep {
	//NOTE: very similar implemenation in hiddenSubset.
	//TODO: randomize order we visit things.
	var results []*SolveStep
	for _, i := range rand.Perm(DIM) {

		groups := subsetCellsWithNPossibilities(k, collectionGetter(i))

		if len(groups) > 0 {

			for _, groupIndex := range rand.Perm(len(groups)) {

				group := groups[groupIndex]

				result := &SolveStep{collectionGetter(i).RemoveCells(group), group, group.PossibilitiesUnion(), technique}
				if result.IsUseful(grid) {
					results = append(results, result)
				}
				//Keep going
			}
		}

	}
	return results
}

func subsetCellsWithNPossibilities(k int, inputCells CellList) []CellList {
	//Given a list of cells (often a row, col, or block) and a target group size K,
	//returns a list of groups of cells of size K where the union of each group's possibility list
	//is size K.

	//Note: this function has performance O(n!/k!(n - k)!)

	//First, cull any cells with no possibilities to help minimize n
	cells := inputCells.FilterByHasPossibilities()

	var results []CellList

	for _, indexes := range subsetIndexes(len(cells), k) {
		//Build up set of all possibilties in this subset.
		subset := cells.Subset(indexes)
		union := subset.PossibilitiesUnion()
		//Okay, we built up the set. Is it the target size?
		if len(union) == k {
			results = append(results, subset)
		}
	}

	return results

}
