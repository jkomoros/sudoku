package sudoku

import (
	"fmt"
	"math/rand"
)

//Different instantiations of this technique represent naked{pair,triple}{row,block,col}
type nakedSubsetTechnique struct {
	*basicSolveTechnique
}

func (self *nakedSubsetTechnique) Difficulty() float64 {
	return self.difficultyHelper(6.0)
}

func (self *nakedSubsetTechnique) Description(step *SolveStep) string {
	if len(step.TargetNums) < self.k || len(step.PointerCells) < self.k {
		return ""
	}
	//TODO: it feels like this logic should be factored out.
	var groupName string
	var groupNum int
	switch self.groupType {
	case GROUP_BLOCK:
		groupName = "block"
		groupNum = step.TargetCells.Block()
	case GROUP_ROW:
		groupName = "row"
		groupNum = step.TargetCells.Row()
	case GROUP_COL:
		groupName = "column"
		groupNum = step.TargetCells.Col()
	default:
		groupName = "<NONE>"
		groupNum = -1
	}

	return fmt.Sprintf("%s are only possible in %s, which means that they can't be in any other cell in %s %d", step.TargetNums.Description(), step.PointerCells.Description(), groupName, groupNum)
}

func (self *nakedSubsetTechnique) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will find multiple if they exist.
	return nakedSubset(grid, self, self.k, self.getter(grid))
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

				result := &SolveStep{collectionGetter(i).RemoveCells(group), group, group.PossibilitiesUnion(), nil, technique}
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
