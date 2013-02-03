package sudoku

import (
	"fmt"
	"math/rand"
)

type SolveDirections []*SolveStep

const (
	ONLY_LEGAL_NUMBER = iota
	NECESSARY_IN_ROW
	NECESSARY_IN_COL
	NECESSARY_IN_BLOCK
)

type SolveStep struct {
	Row       int
	Col       int
	Block     int
	Num       int
	Technique SolveTechnique
}

type SolveTechnique interface {
	Name() string
	Description(*SolveStep) string
	Find(*Grid) *SolveStep
}

var techniques []SolveTechnique

func init() {
	//TODO: init techniques with enough space
	techniques = append(techniques, onlyLegalNumberTechnique{})
	techniques = append(techniques, necessaryInRowTechnique{})
	techniques = append(techniques, necessaryInColTechnique{})
	techniques = append(techniques, necessaryInBlockTechnique{})
}

type onlyLegalNumberTechnique struct {
}

type necessaryInRowTechnique struct {
}

type necessaryInColTechnique struct {
}

type necessaryInBlockTechnique struct {
}

func (self *SolveStep) Apply(grid *Grid) {
	cell := grid.Cell(self.Row, self.Col)
	cell.SetNumber(self.Num)
}

func (self onlyLegalNumberTechnique) Name() string {
	return "Only Legal Number"
}

func (self onlyLegalNumberTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("%d is the only remaining valid number for that cell", step.Num)
}

func (self onlyLegalNumberTechnique) Find(grid *Grid) *SolveStep {
	//This will be a random item
	obj := grid.queue.NewGetter().GetSmallerThan(2)
	if obj == nil {
		//There weren't any cells with one option.
		return nil
	}
	cell := obj.(*Cell)
	return &SolveStep{cell.Row, cell.Col, cell.Block, cell.implicitNumber(), self}
}

func (self necessaryInRowTechnique) Name() string {
	return "Necessary In Row"
}

func (self necessaryInRowTechnique) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	return fmt.Sprintf("%d is required in the %d row, and %d is the only column it fits", step.Num, step.Row+1, step.Col+1)
}

func (self necessaryInRowTechnique) Find(grid *Grid) *SolveStep {
	getter := func(index int) []*Cell {
		return grid.Row(index)
	}
	return necessaryInCollection(grid, self, getter)
}

func (self necessaryInColTechnique) Name() string {
	return "Necessary In Col"
}

func (self necessaryInColTechnique) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	return fmt.Sprintf("%d is required in the %d column, and %d is the only row it fits", step.Num, step.Row+1, step.Col+1)
}

func (self necessaryInColTechnique) Find(grid *Grid) *SolveStep {
	getter := func(index int) []*Cell {
		return grid.Col(index)
	}
	return necessaryInCollection(grid, self, getter)
}

func (self necessaryInBlockTechnique) Name() string {
	return "Necessary In Block"
}

func (self necessaryInBlockTechnique) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	return fmt.Sprintf("%d is required in the %d block, and %d, %d is the only cell it fits", step.Num, step.Block+1, step.Row+1, step.Col+1)
}

func (self necessaryInBlockTechnique) Find(grid *Grid) *SolveStep {
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
						return &SolveStep{cell.Row, cell.Col, cell.Block, index + 1, technique}
					}
				}
			}
		}
	}
	//Nope.
	return nil
}

func (self *Grid) HumanSolve() *SolveDirections {
	return nil
}
