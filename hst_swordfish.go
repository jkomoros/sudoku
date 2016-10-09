package sudoku

import (
	"fmt"
	"math/rand"
)

type swordfishTechnique struct {
	*basicSolveTechnique
}

func (self *swordfishTechnique) humanLikelihood(step *SolveStep) float64 {
	//TODO: reason more carefully about how hard this technique is.
	return self.difficultyHelper(70.0)
}

func (self *swordfishTechnique) Description(step *SolveStep) string {

	if len(step.TargetNums) != 1 {
		return ""
	}

	groupName := "columns"
	otherGroupName := "rows"

	if self.groupType == _GROUP_ROW {
		groupName = "rows"
		otherGroupName = "columns"
	}

	return fmt.Sprintf("%d is only possible in two cells each in three different %s, all of which align onto three %s, which means that %d can't be in any of the other cells in those %s (%s)",
		step.TargetNums[0],
		groupName,
		otherGroupName,
		step.TargetNums[0],
		otherGroupName,
		step.TargetCells.Description(),
	)

}

func (self *swordfishTechnique) Candidates(grid Grid, maxResults int) []*SolveStep {
	return self.candidatesHelper(self, grid, maxResults)
}

func (self *swordfishTechnique) find(grid Grid, coordinator findCoordinator) {

	getter := self.getter(grid)

	//For this technique, the primary access is the first type of group we look at to find
	//cells with only 2 spots for a given number.

	//TODO: Implement the "relaxed" version of this technique, too.

	for _, i := range rand.Perm(DIM + 1) {

		if i == 0 {
			continue
		}
		//The candidate we're considering

		//Check if it's time to stop.
		if coordinator.shouldExitEarly() {
			return
		}

		//Consider each of the major-axis groups to see if more than three have
		//only two candidates for

		var majorAxisGroupsWithTwoOptionsForCandidate []CellSlice

		for c := 0; c < DIM; c++ {
			majorAxisGroup := getter(c)

			cellsWithCandidatePossibility := majorAxisGroup.FilterByPossible(i)

			if len(cellsWithCandidatePossibility) == 2 {
				//TODO: shouldn't we keep track of the rows where the possibilities were, too?
				majorAxisGroupsWithTwoOptionsForCandidate = append(majorAxisGroupsWithTwoOptionsForCandidate, cellsWithCandidatePossibility)
			}

		}

		//Do we have more than three major axis groups identified?
		if len(majorAxisGroupsWithTwoOptionsForCandidate) < 3 {
			continue
		}

		//Consider every subset of size three
		for _, indexes := range subsetIndexes(len(majorAxisGroupsWithTwoOptionsForCandidate), 3) {
			majorAxisGroups := make([]CellSlice, 3)
			for i, index := range indexes {
				majorAxisGroups[i] = majorAxisGroupsWithTwoOptionsForCandidate[index]
			}

			//OK, now majorAxisGroups has the set of three we're operating on.
			//Do their minorAxis groups line up to a set of three as well?

			var minorGroupIndexSet intSet

			for _, group := range majorAxisGroups {
				if self.groupType == _GROUP_COL {
					minorGroupIndexSet = minorGroupIndexSet.union(group.AllRows().toIntSet())
				} else {
					minorGroupIndexSet = minorGroupIndexSet.union(group.AllCols().toIntSet())
				}
			}

			if len(minorGroupIndexSet) != 3 {
				//Nah, didn't have three rows.
				continue
			}

			//Woot, looks like we found a valid set.

			//Generate the list of cells that will be affected
			var affectedCells CellSlice

			for minorGroupIndex, _ := range minorGroupIndexSet {
				if self.groupType == _GROUP_COL {
					affectedCells = append(affectedCells, grid.Row(minorGroupIndex)...)
				} else {
					affectedCells = append(affectedCells, grid.Col(minorGroupIndex)...)
				}
			}

			//Get rid of cells that are already filled, or where removing 'i'
			//would be a no-op anyway since it's not already a possible there.
			affectedCells = affectedCells.FilterByPossible(i)

			//Gather the list of all cells that are "pointing" to this

			var pointerCells CellSlice

			for _, group := range majorAxisGroups {
				pointerCells = append(pointerCells, group...)
			}

			//Get rid of all pointer cells from the rows of cells that will be
			//affected.
			affectedCells = affectedCells.RemoveCells(pointerCells)

			//Okay, now the set is solid.

			//Okay, we have a candidate step (unchunked). Is it useful?
			step := &SolveStep{self,
				affectedCells,
				IntSlice{i},
				pointerCells,
				nil,
				nil,
			}

			if step.IsUseful(grid) {
				if coordinator.foundResult(step) {
					return
				}
			}

		}

	}

}
