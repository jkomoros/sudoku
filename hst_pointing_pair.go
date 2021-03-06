package sudoku

import (
	"fmt"
	"math/rand"
)

type pointingPairTechnique struct {
	*basicSolveTechnique
}

func (self *pointingPairTechnique) humanLikelihood(step *SolveStep) float64 {
	return self.difficultyHelper(50.0)
}

func (self *pointingPairTechnique) Description(step *SolveStep) string {
	if len(step.TargetNums) == 0 {
		return ""
	}

	groupName := "<NONE>"
	groupNum := -1
	if self.groupType == _GROUP_ROW {
		groupName = "row"
		groupNum = step.TargetCells.Row()
	} else if self.groupType == _GROUP_COL {
		groupName = "column"
		groupNum = step.TargetCells.Col()
	}

	return fmt.Sprintf("%d is only possible in %s %d of block %d, which means it can't be in any other cell in that %s not in that block", step.TargetNums[0], groupName, groupNum, step.PointerCells.Block(), groupName)
}

func (self *pointingPairTechnique) Candidates(grid Grid, maxResults int) []*SolveStep {
	return self.candidatesHelper(self, grid, maxResults)
}

func (self *pointingPairTechnique) find(grid Grid, coordinator findCoordinator) {
	//Within each block, for each number, see if all items that allow it are aligned in a row or column.
	//TODO: test this returns multiple if they exist.

	for _, i := range rand.Perm(DIM) {
		block := block(i)
		blockCells := block.CellSlice(grid)

		for _, num := range rand.Perm(DIM) {

			if coordinator.shouldExitEarly() {
				return
			}

			cells := blockCells.FilterByPossible(num + 1).CellReferenceSlice()
			//cellList is now a list of all cells that have that number.
			if len(cells) <= 1 || len(cells) > BLOCK_DIM {
				//Meh, not a match.
				continue
			}
			//Okay, it's possible it's a match. Are all relevant groups (rows or cols, depending on groupType) the same?
			if (self.groupType == _GROUP_ROW && cells.SameRow()) || (self.groupType == _GROUP_COL && cells.SameCol()) {
				//Yup!
				var step *SolveStep
				if self.groupType == _GROUP_ROW {
					step = &SolveStep{self, row(cells.Row()).RemoveCells(block), []int{num + 1}, cells, nil, nil}
				} else {
					step = &SolveStep{self, col(cells.Col()).RemoveCells(block), []int{num + 1}, cells, nil, nil}
				}
				if step.IsUseful(grid) {
					if coordinator.foundResult(step) {
						return
					}
				}
				//Keep looking for more!
			}
		}
	}
}
