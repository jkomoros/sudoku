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
	TargetCells  CellList
	PointerCells CellList
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

type cullSolveTechnique struct {
}

var fillTechniques []SolveTechnique
var cullTechniques []SolveTechnique

func init() {
	//TODO: init techniques with enough space
	fillTechniques = append(fillTechniques, nakedSingleTechnique{})
	fillTechniques = append(fillTechniques, hiddenSingleInRow{})
	fillTechniques = append(fillTechniques, hiddenSingleInCol{})
	fillTechniques = append(fillTechniques, hiddenSingleInBlock{})
	cullTechniques = append(cullTechniques, pointingPairRow{})
	cullTechniques = append(cullTechniques, pointingPairCol{})
	cullTechniques = append(cullTechniques, nakedPairCol{})
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

type pointingPairRow struct {
	*cullSolveTechnique
}

type pointingPairCol struct {
	*cullSolveTechnique
}

type nakedPairCol struct {
	*cullSolveTechnique
}

func (self *fillSolveTechnique) IsFill() bool {
	return true
}

func (self *cullSolveTechnique) IsFill() bool {
	return false
}

func newFillSolveStep(cell *Cell, num int, technique SolveTechnique) *SolveStep {
	//TODO: why do these need to be pulled out separately?
	cellArr := [...]*Cell{cell}
	numArr := [...]int{num}
	return &SolveStep{cellArr[:], nil, numArr[:], technique}
}

func (self *SolveStep) Apply(grid *Grid) {
	if self.Technique.IsFill() {
		if len(self.TargetCells) == 0 || len(self.Nums) == 0 {
			return
		}
		cell := self.TargetCells[0].InGrid(grid)
		cell.SetNumber(self.Nums[0])
	} else {
		for _, cell := range self.TargetCells {
			gridCell := cell.InGrid(grid)
			for _, exclude := range self.Nums {
				gridCell.setExcluded(exclude, true)
			}
		}
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

func (self pointingPairRow) Name() string {
	return "Pointing pair row"
}

func (self pointingPairRow) Description(step *SolveStep) string {
	if len(step.Nums) == 0 {
		return ""
	}
	return fmt.Sprintf("%d is only possible in row %d of block %d, which means it can't be in any other cell in that row not in that block", step.Nums[0], step.TargetCells.Row(), step.PointerCells.Block())
}

func (self pointingPairRow) Find(grid *Grid) *SolveStep {
	//Within each block, for each number, see if all items that allow it are aligned in a row or column.
	//TODO: randomize order of blocks.
	//TODO: this is substantially duplicated in pointingPaircol
	for i := 0; i < DIM; i++ {
		block := grid.Block(i)
		//TODO: randomize order of numbers to test for.
		for num := 0; num < DIM; num++ {
			cells := block.FilterByPossible(num + 1)
			//cellList is now a list of all cells that have that number.
			if len(cells) <= 1 || len(cells) > BLOCK_DIM {
				//Meh, not a match.
				continue
			}
			//Okay, it's possible it's a match. Are all rows the same?
			if cells.SameRow() {
				//Yup!
				return &SolveStep{grid.Row(cells.Row()).RemoveCells(block), cells, []int{num + 1}, self}
			}
		}
	}
	return nil
}

func (self pointingPairCol) Name() string {
	return "Pointing pair col"
}

func (self pointingPairCol) Description(step *SolveStep) string {
	if len(step.Nums) == 0 {
		return ""
	}
	return fmt.Sprintf("%d is only possible in column %d of block %d, which means it can't be in any other cell in that column not in that block", step.Nums[0], step.TargetCells.Col(), step.PointerCells.Block())
}

func (self pointingPairCol) Find(grid *Grid) *SolveStep {
	//Within each block, for each number, see if all items that allow it are aligned in a row or column.
	//TODO: randomize order of blocks.
	//TODO: this is substantially duplicated in pointingPairRow
	for i := 0; i < DIM; i++ {
		block := grid.Block(i)
		//TODO: randomize order of numbers to test for.
		for num := 0; num < DIM; num++ {
			cells := block.FilterByPossible(num + 1)
			//cellList is now a list of all cells that have that number.
			if len(cells) <= 1 || len(cells) > BLOCK_DIM {
				//Meh, not a match.
				continue
			}
			//Okay, are all cols?
			if cells.SameCol() {
				//Yup!
				return &SolveStep{grid.Col(cells.Col()).RemoveCells(block), cells, []int{num + 1}, self}
			}
		}
	}
	return nil
}

func (self nakedPairCol) Name() string {
	return "Naked pair col"
}

func (self nakedPairCol) Description(step *SolveStep) string {
	//TODO: implement
	return ""
}

func (self nakedPairCol) Find(grid *Grid) *SolveStep {
	//TODO: randomize order we visit things.
	for i := 0; i < DIM; i++ {
		//Grab all of the cells in this row that have exactly two possibilities
		//Note: we can assume that there aren't any cells with a single possibility in cells right now
		//since those would have already been filled in before we tried this more advanced technique.
		cells := grid.Col(i).FilterByNumPossibilities(2)

		//Now we compare each cell to every other to see if they are the same list of possibilties.
		for j, cell := range cells {
			for k := j + 1; k < len(cells); k++ {
				otherCell := cells[k]
				if intList(cell.Possibilities()).SameAs(intList(otherCell.Possibilities())) {
					twoCells := []*Cell{cell, otherCell}
					return &SolveStep{grid.Col(cell.Col).RemoveCells(twoCells), twoCells, cell.Possibilities(), self}
				}
			}
		}

	}
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
		techniqueOrder = rand.Perm(len(cullTechniques))
		for _, index := range techniqueOrder {
			technique := cullTechniques[index]
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
