package sudoku

import (
	"fmt"
	"math/rand"
)

//Different instantiations of this technique represent naked{pair,triple}{row,block,col}
type hiddenSubsetTechnique struct {
	basicSolveTechnique
}

func (self hiddenSubsetTechnique) Description(step *SolveStep) string {
	if len(step.Nums) < self.k || len(step.PointerCells) < self.k {
		return ""
	}
	//TODO: this message should say something about the group and number right after the second %s.
	//TODO: also: this message is wrong... the numbers we report are backwards.
	return fmt.Sprintf("%s are only possible in %s, which means that only those numbers could be in those cells", step.Nums.Description(), step.PointerCells.Description())
}

func (self hiddenSubsetTechnique) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will find multiple if they exist.
	return hiddenSubset(grid, self, self.k, self.getter(grid))
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

				result := &SolveStep{group, group, numsToRemove, numList, technique}
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
