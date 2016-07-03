package sudoku

import (
	"fmt"
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

func (self *obviousInCollectionTechnique) humanLikelihood(step *SolveStep) float64 {
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

func (self *obviousInCollectionTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {
	obviousInCollection(grid, self, self.getter(grid), results, done)
	close(results)
}

func obviousInCollection(grid *Grid, technique SolveTechnique, collectionGetter func(index int) CellSlice, results chan *SolveStep, done chan bool) {
	indexes := rand.Perm(DIM)
	for _, index := range indexes {

		select {
		case <-done:
			return
		default:
		}

		collection := collectionGetter(index)
		openCells := collection.FilterByHasPossibilities()
		if len(openCells) == 1 {
			//Okay, only one cell in this collection has an opening, which must mean it has one possibilty.
			cell := openCells[0]
			possibilities := cell.Possibilities()
			//len(possibiltiies) SHOULD be 1, but check just in case.
			if len(possibilities) == 1 {
				possibility := possibilities[0]
				step := &SolveStep{
					Technique:    technique,
					TargetCells:  CellSlice{cell},
					TargetNums:   IntSlice{possibility},
					PointerCells: collection.RemoveCells(CellSlice{cell}),
				}
				if step.IsUseful(grid) {
					select {
					case results <- step:
					case <-done:
						return
					}
				}
			}

		}
	}
}

func (self *nakedSingleTechnique) humanLikelihood(step *SolveStep) float64 {
	return self.difficultyHelper(20.0)
}

func (self *nakedSingleTechnique) Description(step *SolveStep) string {
	if len(step.TargetNums) == 0 {
		return ""
	}
	num := step.TargetNums[0]
	return fmt.Sprintf("%d is the only remaining valid number for that cell", num)
}

func (self *nakedSingleTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {

	defer close(results)
	//TODO: test that this will find multiple if they exist.
	getter := grid.queue().NewGetter()
	for {

		select {
		case <-done:
			return
		default:
		}
		obj := getter.GetSmallerThan(2)
		if obj == nil {
			//There weren't any cells with one option left.
			//If there weren't any, period, then results is still nil already.
			return
		}
		cell := obj.(*Cell)
		step := &SolveStep{
			Technique:    self,
			TargetCells:  CellSlice{cell},
			TargetNums:   IntSlice{cell.implicitNumber()},
			PointerCells: cell.Neighbors().FilterByFilled(),
		}
		if step.IsUseful(grid) {
			select {
			case results <- step:
			case <-done:
				return
			}
		}
	}
}

func (self *hiddenSingleTechnique) humanLikelihood(step *SolveStep) float64 {
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

func (self *hiddenSingleTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {
	//TODO: test that if there are multiple we find them both.
	necessaryInCollection(grid, self, self.getter(grid), results, done)
	close(results)
}

func necessaryInCollection(grid *Grid, technique SolveTechnique, collectionGetter func(index int) CellSlice, results chan *SolveStep, done chan bool) {
	//This will be a random item
	indexes := rand.Perm(DIM)

	for _, i := range indexes {

		select {
		case <-done:
			return
		default:
		}

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
						step := &SolveStep{
							Technique:    technique,
							TargetCells:  CellSlice{cell},
							TargetNums:   IntSlice{index + 1},
							PointerCells: collection.FilterByUnfilled().RemoveCells(CellSlice{cell}),
						}
						if step.IsUseful(grid) {
							select {
							case results <- step:
							case <-done:
								return
							}
							break
						}
						//Hmm, wasn't useful. Keep trying...
					}
				}
			}
		}
	}
}
