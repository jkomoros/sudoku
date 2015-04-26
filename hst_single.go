package sudoku

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
)

type nakedSingleTechnique struct {
	*basicSolveTechnique
}

type hiddenSingleTechnique struct {
	*basicSolveTechnique
}

type obviousInCollectionTechnique struct {
	*basicSolveTechnique
}

func newFillSolveStep(cell *Cell, num int, technique SolveTechnique) *SolveStep {
	cellArr := []*Cell{cell}
	numArr := []int{num}
	return &SolveStep{technique, cellArr, numArr, nil, nil}
}

func (self *obviousInCollectionTechnique) HumanLikelihood() float64 {
	return self.difficultyHelper(1.0)
}

func (self *obviousInCollectionTechnique) Description(step *SolveStep) string {
	if len(step.TargetNums) == 0 {
		return ""
	}
	num := step.TargetNums[0]
	groupName := "<NONE>"
	groupNumber := 0
	switch self.groupType {
	case _GROUP_BLOCK:
		groupName = "block"
		groupNumber = step.TargetCells.Block()
	case _GROUP_COL:
		groupName = "column"
		groupNumber = step.TargetCells.Col()
	case _GROUP_ROW:
		groupName = "row"
		groupNumber = step.TargetCells.Row()
	}

	return fmt.Sprintf("%s is the only cell in %s %d that is unfilled, and it must be %d", step.TargetCells.Description(), groupName, groupNumber, num)
}

func (self *obviousInCollectionTechnique) Find(grid *Grid) []*SolveStep {
	return obviousInCollection(grid, self, self.getter(grid))
}

func obviousInCollection(grid *Grid, technique SolveTechnique, collectionGetter func(index int) CellSlice) []*SolveStep {
	indexes := rand.Perm(DIM)
	var results []*SolveStep
	for _, index := range indexes {
		collection := collectionGetter(index)
		openCells := collection.FilterByHasPossibilities()
		if len(openCells) == 1 {
			//Okay, only one cell in this collection has an opening, which must mean it has one possibilty.
			cell := openCells[0]
			possibilities := cell.Possibilities()
			if len(possibilities) != 1 {
				log.Fatalln("Expected the cell to only have one possibility")
			} else {
				possibility := possibilities[0]
				step := newFillSolveStep(cell, possibility, technique)
				if step.IsUseful(grid) {
					results = append(results, step)
				}
			}

		}
	}
	return results
}

func (self *nakedSingleTechnique) HumanLikelihood() float64 {
	return self.difficultyHelper(20.0)
}

func (self *nakedSingleTechnique) Description(step *SolveStep) string {
	if len(step.TargetNums) == 0 {
		return ""
	}
	num := step.TargetNums[0]
	return fmt.Sprintf("%d is the only remaining valid number for that cell", num)
}

func (self *nakedSingleTechnique) Find(grid *Grid) []*SolveStep {
	//TODO: test that this will find multiple if they exist.
	var results []*SolveStep
	getter := grid.queue().NewGetter()
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

func (self *hiddenSingleTechnique) HumanLikelihood() float64 {
	return self.difficultyHelper(18.0)
}

func (self *hiddenSingleTechnique) Description(step *SolveStep) string {
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
	case _GROUP_BLOCK:
		groupName = "block"
		otherGroupName = "cell"
		groupNum = step.TargetCells.Block()
		otherGroupNum = step.TargetCells.Description()
	case _GROUP_ROW:
		groupName = "row"
		otherGroupName = "column"
		groupNum = step.TargetCells.Row()
		otherGroupNum = strconv.Itoa(cell.Col())
	case _GROUP_COL:
		groupName = "column"
		otherGroupName = "row"
		groupNum = step.TargetCells.Col()
		otherGroupNum = strconv.Itoa(cell.Row())
	default:
		groupName = "<NONE>"
		otherGroupName = "<NONE>"
		groupNum = -1
		otherGroupNum = "<NONE>"
	}

	return fmt.Sprintf("%d is required in the %d %s, and %s is the only %s it fits", num, groupNum, groupName, otherGroupNum, otherGroupName)
}

func (self *hiddenSingleTechnique) Find(grid *Grid) []*SolveStep {
	//TODO: test that if there are multiple we find them both.
	return necessaryInCollection(grid, self, self.getter(grid))
}

func necessaryInCollection(grid *Grid, technique SolveTechnique, collectionGetter func(index int) CellSlice) []*SolveStep {
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
