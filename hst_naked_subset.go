package sudoku

import (
	"fmt"
	"math/rand"
)

//Different instantiations of this technique represent naked{pair,triple}{row,block,col}
type nakedSubsetTechnique struct {
	*basicSolveTechnique
}

func (self *nakedSubsetTechnique) humanLikelihood(step *SolveStep) float64 {
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

	return fmt.Sprintf("%s are only possible in %s, which means that they can't be in any other cell in %s %d", step.TargetNums.Description(), step.PointerCells.Description(), groupName, groupNum)
}

func (self *nakedSubsetTechnique) Candidates(grid *Grid, maxResults int) []*SolveStep {
	return self.candidatesHelper(self, grid, maxResults)
}

func (self *nakedSubsetTechnique) find(grid *Grid, results chan *SolveStep, done chan bool) {
	//TODO: test that this will find multiple if they exist.
	nakedSubset(grid, self, self.k, self.getter(grid), results, done)
}

func nakedSubset(grid *Grid, technique SolveTechnique, k int, collectionGetter func(int) CellSlice, results chan *SolveStep, done chan bool) {
	//NOTE: very similar implemenation in hiddenSubset.
	//TODO: randomize order we visit things.
	for _, i := range rand.Perm(DIM) {

		select {
		case <-done:
			return
		default:
		}

		groups := subsetCellsWithNPossibilities(k, collectionGetter(i))

		if len(groups) > 0 {

			for _, groupIndex := range rand.Perm(len(groups)) {

				group := groups[groupIndex]

				step := &SolveStep{technique, collectionGetter(i).RemoveCells(group).FilterByUnfilled(), group.PossibilitiesUnion(), group, nil, nil}
				if step.IsUseful(grid) {
					select {
					case results <- step:
					case <-done:
						return
					}
				}
				//Keep going
			}
		}

	}
}

func subsetCellsWithNPossibilities(k int, inputCells CellSlice) []CellSlice {
	//Given a list of cells (often a row, col, or block) and a target group size K,
	//returns a list of groups of cells of size K where the union of each group's possibility list
	//is size K.

	//Note: this function has performance O(n!/k!(n - k)!)

	//First, cull any cells with no possibilities to help minimize n
	cells := inputCells.FilterByHasPossibilities()

	var results []CellSlice

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
