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
	techniques = append(techniques, nakedSingleTechnique{})
	techniques = append(techniques, hiddenSingleInRow{})
	techniques = append(techniques, hiddenSingleInCol{})
	techniques = append(techniques, necessaryInBlockTechnique{})
}

type nakedSingleTechnique struct {
}

type hiddenSingleInRow struct {
}

type hiddenSingleInCol struct {
}

type necessaryInBlockTechnique struct {
}

func (self *SolveStep) Apply(grid *Grid) {
	cell := grid.Cell(self.Row, self.Col)
	cell.SetNumber(self.Num)
}

func (self nakedSingleTechnique) Name() string {
	return "Only Legal Number"
}

func (self nakedSingleTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("%d is the only remaining valid number for that cell", step.Num)
}

func (self nakedSingleTechnique) Find(grid *Grid) *SolveStep {
	//This will be a random item
	obj := grid.queue.NewGetter().GetSmallerThan(2)
	if obj == nil {
		//There weren't any cells with one option.
		return nil
	}
	cell := obj.(*Cell)
	return &SolveStep{cell.Row, cell.Col, cell.Block, cell.implicitNumber(), self}
}

func (self hiddenSingleInRow) Name() string {
	return "Necessary In Row"
}

func (self hiddenSingleInRow) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	return fmt.Sprintf("%d is required in the %d row, and %d is the only column it fits", step.Num, step.Row+1, step.Col+1)
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
	return fmt.Sprintf("%d is required in the %d column, and %d is the only row it fits", step.Num, step.Row+1, step.Col+1)
}

func (self hiddenSingleInCol) Find(grid *Grid) *SolveStep {
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

func (self *Grid) HumanSolve() SolveDirections {
	var results []*SolveStep
	for !self.Solved() {
		//TODO: try the techniques in parallel
		//TODO: pick the technique based on a weighting of how common a human is to pick each one.
		//TODO: provide hints to the techniques of where to look based on the last filled cell
		techniqueOrder := rand.Perm(len(techniques))
		for _, index := range techniqueOrder {
			technique := techniques[index]
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
