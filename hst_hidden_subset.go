package sudoku

import (
	"fmt"
	"math/rand"
)

type hiddenPairCol struct {
	basicSolveTechnique
}

type hiddenPairRow struct {
	basicSolveTechnique
}

type hiddenPairBlock struct {
	basicSolveTechnique
}

func (self hiddenPairCol) Description(step *SolveStep) string {
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that only those numbers could be in those cells", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Col())
}

func (self hiddenPairCol) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will find multiple if they exist.
	//TODO: factor out colGetter, we use it so often...
	colGetter := func(i int) CellList {
		return grid.Col(i)
	}
	return hiddenSubset(grid, self, 2, colGetter)
}

func (self hiddenPairRow) Description(step *SolveStep) string {
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that only those numbers could be in those cells", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Row())
}

func (self hiddenPairRow) Find(grid *Grid) []*SolveStep {
	//TODO: test we find multiple if they exist.
	rowGetter := func(i int) CellList {
		return grid.Row(i)
	}
	return hiddenSubset(grid, self, 2, rowGetter)
}

func (self hiddenPairBlock) Description(step *SolveStep) string {
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that only those numbers could be in those cells", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Block())
}

func (self hiddenPairBlock) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will return multiple if they exist.
	blockGetter := func(i int) CellList {
		return grid.Block(i)
	}
	return hiddenSubset(grid, self, 2, blockGetter)
}

func hiddenSubset(grid *Grid, technique SolveTechnique, k int, collectionGetter func(int) CellList) []*SolveStep {
	//NOTE: very similar implemenation in nakedSubset.
	var results []*SolveStep
	for _, i := range rand.Perm(DIM) {

		groups, nums := subsetCellsWithNUniquePossibilities(k, collectionGetter(i))

		if len(groups) > 0 {
			for _, groupIndex := range rand.Perm(len(groups)) {

				numList := nums[groupIndex]
				group := groups[groupIndex]

				//numList is the numbers we want to KEEP. We need to invertit.

				numsToRemove := group.PossibilitiesUnion().Difference(numList)

				result := &SolveStep{group, group, numsToRemove, technique}
				if result.IsUseful(grid) {
					results = append(results, result)
				}
			}
		}
	}
	return results
}

//TODO: come up with a better name for this HiddenSubset technique helper method
func subsetCellsWithNUniquePossibilities(k int, inputCells CellList) ([]CellList, []IntSlice) {
	//Given a list of cells (often a row, col, or block) and a target group size K,
	//returns a list of groups of cells of size K where all of the cells have K
	//candidates that don't appear anywhere else in the group.

	//TODO: note the runtime complexity (if it's large)

	//First, cull any cells with no possibilities to help minimize n
	cells := inputCells.FilterByHasPossibilities()

	var cellResults []CellList
	var intResults []IntSlice

	for _, indexes := range subsetIndexes(len(cells), k) {
		subset := cells.Subset(indexes)
		inverseSubset := cells.InverseSubset(indexes)

		//All of the OTHER numbers
		possibilitiesUnion := subset.PossibilitiesUnion()
		inversePossibilitiesUnion := inverseSubset.PossibilitiesUnion()

		//Now try every K-sized subset of possibilitiesUnion
		//subsetIndexes will detect the case where the set is already too small and return nil
		for _, possibilitiesIndexes := range subsetIndexes(len(possibilitiesUnion), k) {
			set := possibilitiesUnion.Subset(possibilitiesIndexes)
			if len(set.Intersection(inversePossibilitiesUnion)) == 0 {
				cellResults = append(cellResults, subset)
				intResults = append(intResults, set)

			}
		}
	}
	return cellResults, intResults
}
