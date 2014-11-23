package sudoku

import (
	"math/rand"
)

type xwingTechnique struct {
	*basicSolveTechnique
}

func (self *xwingTechnique) Difficulty() float64 {
	return self.difficultyHelper(5.0)
}

func (self *xwingTechnique) Description(step *SolveStep) string {
	//TODO: implement this
	return ""
}

func (self *xwingTechnique) Find(grid *Grid) []*SolveStep {

	var results []*SolveStep

	getter := self.getter(grid)

	//For each number
	for _, i := range rand.Perm(DIM) {
		//In comments we'll say "Row" for the major group type, and "col" for minor group type, just for easier comprehension.
		//Look for each row that has that number possible in only two cells.

		//i is zero indexed right now
		i++

		var majorGroups []CellList

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
			var targetCells CellList

			currentGroups := []CellList{majorGroups[subsets[0]], majorGroups[subsets[1]]}

			//Are the possibilities in each row in the same column as the one above?
			//We need to do this differently depending on if we're row or col.
			if self.groupType == GROUP_ROW {
				//TODO: figure out a way to factor group row and col better so we don't duplicate code like this.
				if currentGroups[0][0].Col != currentGroups[1][0].Col || currentGroups[0][1].Col != currentGroups[1][1].Col {
					//Nope, the cells didn't line up.
					continue
				}
				//All of the cells in those two columns
				targetCells = append(grid.Col(currentGroups[0][0].Col), grid.Col(currentGroups[0][1].Col)...)

			} else if self.groupType == GROUP_COL {
				if currentGroups[0][0].Row != currentGroups[1][0].Row || currentGroups[0][1].Row != currentGroups[1][1].Row {
					//Nope, the cells didn't line up.
					continue
				}
				//All of the cells in those two columns
				targetCells = append(grid.Row(currentGroups[0][0].Row), grid.Row(currentGroups[0][1].Row)...)

			}

			//Then remove the cells that are the pointerCells
			targetCells = targetCells.RemoveCells(currentGroups[0])
			targetCells = targetCells.RemoveCells(currentGroups[1])

			//And remove the cells that don't have the target number to remove (to keep the set tight; technically it's OK to include them it would just be a no-op for those cells)
			targetCells = targetCells.FilterByPossible(i)

			//Okay, we found a pair that works. Create a step for it (if it's useful)
			step := &SolveStep{targetCells, append(currentGroups[0], currentGroups[1]...), IntSlice{i}, nil, self}
			if step.IsUseful(grid) {
				results = append(results, step)
			}
		}

	}
	return results
}
