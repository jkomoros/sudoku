package sudoku

import (
	"fmt"
	"math/rand"
)

type SolveDirections []*SolveStep

const (
	NAKED_SINGLE = iota
	HIDDEN_SINGLE_IN_ROW
	HIDDEN_SINGLE_IN_COL
	HIDDEN_SINGLE_IN_BLOCK
)

type SolveStep struct {
	TargetCells  []*Cell
	PointerCells []*Cell
	Nums         []int
	Technique    SolveTechnique
}

type SolveTechnique interface {
	Name() string
	Description(*SolveStep) string
	Find(*Grid) *SolveStep
	IsFill() bool
}

type fillSolveTechnique struct {
}

var fillTechniques []SolveTechnique

func init() {
	//TODO: init techniques with enough space
	fillTechniques = append(fillTechniques, nakedSingleTechnique{})
	fillTechniques = append(fillTechniques, hiddenSingleInRow{})
	fillTechniques = append(fillTechniques, hiddenSingleInCol{})
	fillTechniques = append(fillTechniques, hiddenSingleInBlock{})
}

type nakedSingleTechnique struct {
	*fillSolveTechnique
}

type hiddenSingleInRow struct {
	*fillSolveTechnique
}

type hiddenSingleInCol struct {
	*fillSolveTechnique
}

type hiddenSingleInBlock struct {
	*fillSolveTechnique
}

func (self *fillSolveTechnique) IsFill() bool {
	return true
}

func newFillSolveStep(cell *Cell, num int, technique SolveTechnique) *SolveStep {
	//TODO: why do these need to be pulled out separately?
	cellArr := [...]*Cell{cell}
	numArr := [...]int{num}
	return &SolveStep{cellArr[:], nil, numArr[:], technique}
}

func (self *SolveStep) Apply(grid *Grid) {
	//TODO: also handle non isFill items.
	if self.Technique.IsFill() {
		if len(self.TargetCells) == 0 || len(self.Nums) == 0 {
			return
		}
		cell := self.TargetCells[0].InGrid(grid)
		cell.SetNumber(self.Nums[0])
	}
}

func (self nakedSingleTechnique) Name() string {
	return "Only Legal Number"
}

func (self nakedSingleTechnique) Description(step *SolveStep) string {
	if len(step.Nums) == 0 {
		return ""
	}
	num := step.Nums[0]
	return fmt.Sprintf("%d is the only remaining valid number for that cell", num)
}

func (self nakedSingleTechnique) Find(grid *Grid) *SolveStep {
	//This will be a random item
	obj := grid.queue.NewGetter().GetSmallerThan(2)
	if obj == nil {
		//There weren't any cells with one option.
		return nil
	}
	cell := obj.(*Cell)
	return newFillSolveStep(cell, cell.implicitNumber(), self)
}

func (self hiddenSingleInRow) Name() string {
	return "Necessary In Row"
}

func (self hiddenSingleInRow) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	if len(step.TargetCells) == 0 || len(step.Nums) == 0 {
		return ""
	}
	cell := step.TargetCells[0]
	num := step.Nums[0]
	return fmt.Sprintf("%d is required in the %d row, and %d is the only column it fits", num, cell.Row+1, cell.Col+1)
}

func (self hiddenSingleInRow) Find(grid *Grid) *SolveStep {
	getter := func(index int) []*Cell {
		return grid.Row(index)
	}
	return necessaryInCollection(grid, self, getter)
}

func (self hiddenSingleInCol) Name() string {
	return "Necessary In Col"
}

func (self hiddenSingleInCol) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	if len(step.TargetCells) == 0 || len(step.Nums) == 0 {
		return ""
	}
	cell := step.TargetCells[0]
	num := step.Nums[0]
	return fmt.Sprintf("%d is required in the %d column, and %d is the only row it fits", num, cell.Row+1, cell.Col+1)
}

func (self hiddenSingleInCol) Find(grid *Grid) *SolveStep {
	getter := func(index int) []*Cell {
		return grid.Col(index)
	}
	return necessaryInCollection(grid, self, getter)
}

func (self hiddenSingleInBlock) Name() string {
	return "Necessary In Block"
}

func (self hiddenSingleInBlock) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	if len(step.TargetCells) == 0 || len(step.Nums) == 0 {
		return ""
	}
	cell := step.TargetCells[0]
	num := step.Nums[0]
	return fmt.Sprintf("%d is required in the %d block, and %d, %d is the only cell it fits", num, cell.Block+1, cell.Row+1, cell.Col+1)
}

func (self hiddenSingleInBlock) Find(grid *Grid) *SolveStep {
	getter := func(index int) []*Cell {
		return grid.Block(index)
	}
	return necessaryInCollection(grid, self, getter)
}

func necessaryInCollection(grid *Grid, technique SolveTechnique, collectionGetter func(index int) []*Cell) *SolveStep {
	//This will be a random item
	indexes := rand.Perm(DIM)

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
						//Found it!
						return newFillSolveStep(cell, index+1, technique)
					}
				}
			}
		}
	}
	//Nope.
	return nil
}

func (self *Grid) HumanSolve() SolveDirections {
	var results []*SolveStep
	for !self.Solved() {
		//TODO: try the techniques in parallel
		//TODO: pick the technique based on a weighting of how common a human is to pick each one.
		//TODO: provide hints to the techniques of where to look based on the last filled cell
		techniqueOrder := rand.Perm(len(fillTechniques))
		for _, index := range techniqueOrder {
			technique := fillTechniques[index]
			step := technique.Find(self)
			if step != nil {
				results = append(results, step)
				step.Apply(self)
				break
			}
		}
	}
	if !self.Solved() {
		//We couldn't solve the puzzle.
		return nil
	}
	return results
}
