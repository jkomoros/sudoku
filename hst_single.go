package sudoku

import (
	"fmt"
	"math/rand"
)

type nakedSingleTechnique struct {
	basicSolveTechnique
}

type hiddenSingleInRow struct {
	basicSolveTechnique
}

type hiddenSingleInCol struct {
	basicSolveTechnique
}

type hiddenSingleInBlock struct {
	basicSolveTechnique
}

func newFillSolveStep(cell *Cell, num int, technique SolveTechnique) *SolveStep {
	cellArr := []*Cell{cell}
	numArr := []int{num}
	return &SolveStep{cellArr, nil, numArr, nil, technique}
}

func (self nakedSingleTechnique) Description(step *SolveStep) string {
	if len(step.TargetNums) == 0 {
		return ""
	}
	num := step.TargetNums[0]
	return fmt.Sprintf("%d is the only remaining valid number for that cell", num)
}

func (self nakedSingleTechnique) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will find multiple if they exist.
	var results []*SolveStep
	getter := grid.queue.NewGetter()
	for {
		obj := getter.GetSmallerThan(2)
		if obj == nil {
			//There weren't any cells with one option left.
			//If there weren't any, period, then results is still nil already.
			return results
		}
		cell := obj.(*Cell)
		result := newFillSolveStep(cell, cell.implicitNumber(), self)
		if result.IsUseful(grid) {
			results = append(results, result)
		}
	}
}

func (self hiddenSingleInRow) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	if len(step.TargetCells) == 0 || len(step.TargetNums) == 0 {
		return ""
	}
	cell := step.TargetCells[0]
	num := step.TargetNums[0]
	return fmt.Sprintf("%d is required in the %d row, and %d is the only column it fits", num, cell.Row+1, cell.Col+1)
}

func (self hiddenSingleInRow) Find(grid *Grid) []*SolveStep {
	//TODO: test that if there are multiple we find them both.
	return necessaryInCollection(grid, self, self.getter(grid))
}

func (self hiddenSingleInCol) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	if len(step.TargetCells) == 0 || len(step.TargetNums) == 0 {
		return ""
	}
	cell := step.TargetCells[0]
	num := step.TargetNums[0]
	return fmt.Sprintf("%d is required in the %d column, and %d is the only row it fits", num, cell.Row+1, cell.Col+1)
}

func (self hiddenSingleInCol) Find(grid *Grid) []*SolveStep {
	//TODO: test this will find multiple if they exist.
	return necessaryInCollection(grid, self, self.getter(grid))
}

func (self hiddenSingleInBlock) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	if len(step.TargetCells) == 0 || len(step.TargetNums) == 0 {
		return ""
	}
	cell := step.TargetCells[0]
	num := step.TargetNums[0]
	return fmt.Sprintf("%d is required in the %d block, and %d, %d is the only cell it fits", num, cell.Block+1, cell.Row+1, cell.Col+1)
}

func (self hiddenSingleInBlock) Find(grid *Grid) []*SolveStep {
	//TODO: Verify we find multiples if they exist.
	return necessaryInCollection(grid, self, self.getter(grid))
}

func necessaryInCollection(grid *Grid, technique SolveTechnique, collectionGetter func(index int) CellList) []*SolveStep {
	//This will be a random item
	indexes := rand.Perm(DIM)

	var results []*SolveStep

	for _, i := range indexes {
		seenInCollection := make([]int, DIM)
		collection := collectionGetter(i)
		for _, cell := range collection {
			for _, possibility := range cell.Possibilities() {
				seenInCollection[possibility-1]++
			}
		}
		seenIndexes := rand.Perm(DIM)
		for _, index := range seenIndexes {
			seen := seenInCollection[index]
			if seen == 1 {
				//Okay, we know our target number. Which cell was it?
				for _, cell := range collection {
					if cell.Possible(index + 1) {
						//Found it... just make sure it's useful (it would be rare for it to not be).
						result := newFillSolveStep(cell, index+1, technique)
						if result.IsUseful(grid) {
							results = append(results, result)
							break
						}
						//Hmm, wasn't useful. Keep trying...
					}
				}
			}
		}
	}
	return results
}
