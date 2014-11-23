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
		if len(majorGroups) != 2 {
			//Nope, continue on to next number.
			continue
		}

		var targetCells CellList

		//Are the possibilities in each row in the same column as the one above?
		//We need to do this differently depending on if we're row or col.
		if self.groupType == GROUP_ROW {
			//TODO: figure out a way to factor group row and col better so we don't duplicate code like this.
			if majorGroups[0][0].Col != majorGroups[1][0].Col || majorGroups[0][1].Col != majorGroups[1][1].Col {
				//Nope, the cells didn't line up.
				continue
			}
			//All of the cells in those two columns
			targetCells = append(grid.Col(majorGroups[0][0].Col), grid.Col(majorGroups[1][0].Col)...)

		} else if self.groupType == GROUP_COL {
			if majorGroups[0][0].Row != majorGroups[1][0].Row || majorGroups[0][1].Row != majorGroups[1][1].Row {
				//Nope, the cells didn't line up.
				continue
			}
			//All of the cells in those two columns
			targetCells = append(grid.Row(majorGroups[0][0].Row), grid.Row(majorGroups[1][0].Row)...)

		}

		//Then remove the cells that are the pointerCells
		targetCells = targetCells.RemoveCells(majorGroups[0])
		targetCells = targetCells.RemoveCells(majorGroups[1])

		//Okay, we found a pair that works. Create a step for it (if it's useful)
		step := &SolveStep{targetCells, append(majorGroups[0], majorGroups[1]...), IntSlice{i}, nil, self}
		if step.IsUseful(grid) {
			results = append(results, step)
		}
	}
	return results
}
