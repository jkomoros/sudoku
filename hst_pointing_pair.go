package sudoku

import (
	"fmt"
	"math/rand"
)

type pointingPairRow struct {
	basicSolveTechnique
}

type pointingPairCol struct {
	basicSolveTechnique
}

func (self pointingPairRow) Description(step *SolveStep) string {
	if len(step.TargetNums) == 0 {
		return ""
	}
	return fmt.Sprintf("%d is only possible in row %d of block %d, which means it can't be in any other cell in that row not in that block", step.TargetNums[0], step.TargetCells.Row(), step.PointerCells.Block())
}

func (self pointingPairRow) Find(grid *Grid) []*SolveStep {
	//Within each block, for each number, see if all items that allow it are aligned in a row or column.
	//TODO: randomize order of blocks.
	//TODO: this is substantially duplicated in pointingPaircol
	//TODO: test this returns multiple if they exist.

	var results []*SolveStep

	for _, i := range rand.Perm(DIM) {
		block := grid.Block(i)

		for _, num := range rand.Perm(DIM) {
			cells := block.FilterByPossible(num + 1)
			//cellList is now a list of all cells that have that number.
			if len(cells) <= 1 || len(cells) > BLOCK_DIM {
				//Meh, not a match.
				continue
			}
			//Okay, it's possible it's a match. Are all rows the same?
			if cells.SameRow() {
				//Yup!
				result := &SolveStep{grid.Row(cells.Row()).RemoveCells(block), cells, []int{num + 1}, nil, self}
				if result.IsUseful(grid) {
					results = append(results, result)
				}
				//Keep looking for more!
			}
		}
	}
	return results
}

func (self pointingPairCol) Description(step *SolveStep) string {
	if len(step.TargetNums) == 0 {
		return ""
	}
	return fmt.Sprintf("%d is only possible in column %d of block %d, which means it can't be in any other cell in that column not in that block", step.TargetNums[0], step.TargetCells.Col(), step.PointerCells.Block())
}

func (self pointingPairCol) Find(grid *Grid) []*SolveStep {
	//Within each block, for each number, see if all items that allow it are aligned in a row or column.
	//TODO: randomize order of blocks.
	//TODO: this is substantially duplicated in pointingPairRow
	//TODO: test this will find multiple if they exist.

	var results []*SolveStep

	for _, i := range rand.Perm(DIM) {
		block := grid.Block(i)
		//TODO: randomize order of numbers to test for.

		for _, num := range rand.Perm(DIM) {
			cells := block.FilterByPossible(num + 1)
			//cellList is now a list of all cells that have that number.
			if len(cells) <= 1 || len(cells) > BLOCK_DIM {
				//Meh, not a match.
				continue
			}
			//Okay, are all cols?
			if cells.SameCol() {
				//Yup!
				result := &SolveStep{grid.Col(cells.Col()).RemoveCells(block), cells, []int{num + 1}, nil, self}
				if result.IsUseful(grid) {
					results = append(results, result)
				}
				//Keep looking!
			}
		}
	}
	return results
}
