package sudoku

import (
	"fmt"
	"math/rand"
	"strconv"
)

type nakedSingleTechnique struct {
	basicSolveTechnique
}

type hiddenSingleTechnique struct {
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

func (self hiddenSingleTechnique) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	if len(step.TargetCells) == 0 || len(step.TargetNums) == 0 {
		return ""
	}
	cell := step.TargetCells[0]
	num := step.TargetNums[0]

	var groupName string
	var otherGroupName string
	var groupNum int
	var otherGroupNum string
	switch self.groupType {
	case GROUP_BLOCK:
		groupName = "block"
		otherGroupName = "cell"
		groupNum = step.TargetCells.Block()
		otherGroupNum = step.TargetCells.Description()
	case GROUP_ROW:
		groupName = "row"
		otherGroupName = "column"
		groupNum = step.TargetCells.Row()
		otherGroupNum = strconv.Itoa(cell.Col)
	case GROUP_COL:
		groupName = "column"
		otherGroupName = "row"
		groupNum = step.TargetCells.Col()
		otherGroupNum = strconv.Itoa(cell.Row)
	default:
		groupName = "<NONE>"
		otherGroupName = "<NONE>"
		groupNum = -1
		otherGroupNum = "<NONE>"
	}

	return fmt.Sprintf("%d is required in the %d %s, and %s is the only %s it fits", num, groupNum, groupName, otherGroupNum, otherGroupName)
}

func (self hiddenSingleTechnique) Find(grid *Grid) []*SolveStep {
	//TODO: test that if there are multiple we find them both.
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
