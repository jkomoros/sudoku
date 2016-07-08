package sudoku

import (
	"fmt"
	"math/rand"
)

type xwingTechnique struct {
	*basicSolveTechnique
}

func (self *xwingTechnique) humanLikelihood(step *SolveStep) float64 {
	return self.difficultyHelper(50.0)
}

func (self *xwingTechnique) Description(step *SolveStep) string {
	majorAxis := "NONE"
	minorAxis := "NONE"
	var majorGroups IntSlice
	var minorGroups IntSlice
	switch self.groupType {
	case _GROUP_ROW:
		majorAxis = "rows"
		minorAxis = "columns"
		majorGroups = step.PointerCells.CollectNums(getRow).Unique()
		minorGroups = step.PointerCells.CollectNums(getCol).Unique()
	case _GROUP_COL:
		majorAxis = "columns"
		minorAxis = "rows"
		majorGroups = step.PointerCells.CollectNums(getCol).Unique()
		minorGroups = step.PointerCells.CollectNums(getRow).Unique()
	}

	//Ensure a stable description; Unique() doesn't have a guranteed order.
	majorGroups.Sort()
	minorGroups.Sort()

	return fmt.Sprintf("in %s %s, %d is only possible in %s %s, and %d must be in one of those cells per %s, so it can't be in any other cells in those %s",
		majorAxis,
		majorGroups.Description(),
		step.TargetNums[0],
		minorAxis,
		minorGroups.Description(),
		step.TargetNums[0],
		majorAxis,
		minorAxis,
	)
}

func (self *xwingTechnique) Candidates(grid Grid, maxResults int) []*SolveStep {
	return self.candidatesHelper(self, grid, maxResults)
}

func (self *xwingTechnique) find(grid Grid, coordinator findCoordinator) {

	getter := self.getter(grid)

	//For each number
	for _, i := range rand.Perm(DIM) {
		//In comments we'll say "Row" for the major group type, and "col" for minor group type, just for easier comprehension.
		//Look for each row that has that number possible in only two cells.

		if coordinator.shouldExitEarly() {
			return
		}

		//i is zero indexed right now
		i++

		var majorGroups []CellSlice

		for groupIndex := 0; groupIndex < DIM; groupIndex++ {
			group := getter(groupIndex)
			cells := group.FilterByPossible(i)
			if len(cells) == 2 {
				//Found a row that might fit the bill.
				majorGroups = append(majorGroups, cells)
			}
		}

		//Okay, did we have two rows?
		if len(majorGroups) < 2 {
			//Not enough rows
			continue
		}

		//Now look at each pair of rows and see if their numbers line up.
		for _, subsets := range subsetIndexes(len(majorGroups), 2) {

			if coordinator.shouldExitEarly() {
				return
			}

			var targetCells CellSlice

			currentGroups := []CellSlice{majorGroups[subsets[0]], majorGroups[subsets[1]]}

			//Are the possibilities in each row in the same column as the one above?
			//We need to do this differently depending on if we're row or col.
			if self.groupType == _GROUP_ROW {
				//TODO: figure out a way to factor group row and col better so we don't duplicate code like this.
				if currentGroups[0][0].Col() != currentGroups[1][0].Col() || currentGroups[0][1].Col() != currentGroups[1][1].Col() {
					//Nope, the cells didn't line up.
					continue
				}
				//All of the cells in those two columns
				targetCells = append(grid.Col(currentGroups[0][0].Col()), grid.Col(currentGroups[0][1].Col())...)

			} else if self.groupType == _GROUP_COL {
				if currentGroups[0][0].Row() != currentGroups[1][0].Row() || currentGroups[0][1].Row() != currentGroups[1][1].Row() {
					//Nope, the cells didn't line up.
					continue
				}
				//All of the cells in those two columns
				targetCells = append(grid.Row(currentGroups[0][0].Row()), grid.Row(currentGroups[0][1].Row())...)

			}

			//Then remove the cells that are the pointerCells
			targetCells = targetCells.RemoveCells(currentGroups[0])
			targetCells = targetCells.RemoveCells(currentGroups[1])

			//And remove the cells that don't have the target number to remove (to keep the set tight; technically it's OK to include them it would just be a no-op for those cells)
			targetCells = targetCells.FilterByPossible(i)

			//Okay, we found a pair that works. Create a step for it (if it's useful)
			step := &SolveStep{self, targetCells, IntSlice{i}, append(currentGroups[0], currentGroups[1]...), nil, nil}
			if step.IsUseful(grid) {
				if coordinator.foundResult(step) {
					return
				}
			}
		}

	}
}
