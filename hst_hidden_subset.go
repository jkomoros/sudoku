package sudoku

import (
	"fmt"
	"math/rand"
)

//Different instantiations of this technique represent naked{pair,triple}{row,block,col}
type hiddenSubsetTechnique struct {
	*basicSolveTechnique
}

func (self *hiddenSubsetTechnique) humanLikelihood(step *SolveStep) float64 {
	return self.difficultyHelper(120.0)
}

func (self hiddenSubsetTechnique) Description(step *SolveStep) string {
	if len(step.TargetNums) < self.k || len(step.PointerCells) < self.k {
		return ""
	}
	//TODO: this is substantially duplicated logic in nakedSubsetTechnique
	var groupName string
	var groupNum int
	switch self.groupType {
	case _GROUP_BLOCK:
		groupName = "block"
		groupNum = step.TargetCells.Block()
	case _GROUP_ROW:
		groupName = "row"
		groupNum = step.TargetCells.Row()
	case _GROUP_COL:
		groupName = "column"
		groupNum = step.TargetCells.Col()
	default:
		groupName = "<NONE>"
		groupNum = -1
	}

	return fmt.Sprintf("%s are only possible in %s within %s %d, which means that only those numbers could be in those cells", step.PointerNums.Description(), step.PointerCells.Description(), groupName, groupNum)
}

func (self *hiddenSubsetTechnique) Candidates(grid Grid, maxResults int) []*SolveStep {
	return self.candidatesHelper(self, grid, maxResults)
}

func (self *hiddenSubsetTechnique) find(grid Grid, coordinator findCoordinator) {
	//TODO: test that this will find multiple if they exist.
	hiddenSubset(grid, self, self.k, self.getter(grid), coordinator)
}

func hiddenSubset(grid Grid, technique SolveTechnique, k int, collectionGetter func(int) CellSlice, coordinator findCoordinator) {
	//NOTE: very similar implemenation in nakedSubset.
	for _, i := range rand.Perm(DIM) {

		if coordinator.shouldExitEarly() {
			return
		}

		groups, nums := subsetCellsWithNUniquePossibilities(k, collectionGetter(i))

		if len(groups) > 0 {
			for _, groupIndex := range rand.Perm(len(groups)) {

				numList := nums[groupIndex]
				group := groups[groupIndex]

				//numList is the numbers we want to KEEP. We need to invertit.

				numsToRemove := group.CellSlice(grid).PossibilitiesUnion().Difference(numList)

				step := &SolveStep{technique, group, numsToRemove, group, numList, nil}
				if step.IsUseful(grid) {
					if coordinator.foundResult(step) {
						return
					}
				}
			}
		}
	}
}

//TODO: come up with a better name for this HiddenSubset technique helper method
func subsetCellsWithNUniquePossibilities(k int, inputCells CellSlice) ([]CellRefSlice, []IntSlice) {
	//Given a list of cells (often a row, col, or block) and a target group size K,
	//returns a list of groups of cells of size K where all of the cells have K
	//candidates that don't appear anywhere else in the group.

	//TODO: note the runtime complexity (if it's large)

	//First, cull any cells with no possibilities to help minimize n
	cells := inputCells.FilterByHasPossibilities()

	var cellResults []CellRefSlice
	var intResults []IntSlice

	for _, indexes := range subsetIndexes(len(cells), k) {
		subset := cells.Subset(indexes)
		inverseSubset := cells.InverseSubset(indexes)

		//All of the OTHER numbers
		possibilitiesUnion := subset.PossibilitiesUnion()
		inversePossibilitiesUnionSet := inverseSubset.PossibilitiesUnion().toIntSet()

		//Now try every K-sized subset of possibilitiesUnion
		//subsetIndexes will detect the case where the set is already too small and return nil
		for _, possibilitiesIndexes := range subsetIndexes(len(possibilitiesUnion), k) {
			set := possibilitiesUnion.Subset(possibilitiesIndexes).toIntSet()
			if !set.overlaps(inversePossibilitiesUnionSet) {
				//This happens very, very rarely (< 0.1% of the time)
				cellResults = append(cellResults, subset.CellReferenceSlice())
				intResults = append(intResults, set.toSlice())

			}
		}
	}
	return cellResults, intResults
}
