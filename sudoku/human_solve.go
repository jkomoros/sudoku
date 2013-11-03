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
	Nums         IntSlice
	Technique    SolveTechnique
}

type SolveTechnique interface {
	Name() string
	Description(*SolveStep) string
	Find(*Grid) *SolveStep
	IsFill() bool
}

type fillSolveTechnique struct {
	basicSolveTechnique
}

type cullSolveTechnique struct {
	basicSolveTechnique
}

type basicSolveTechnique struct {
}

var techniques []SolveTechnique

func init() {

	techniques = []SolveTechnique{
		nakedSingleTechnique{},
		hiddenSingleInRow{},
		hiddenSingleInCol{},
		hiddenSingleInBlock{},
		pointingPairRow{},
		pointingPairCol{},
		nakedPairCol{},
		nakedPairRow{},
		nakedPairBlock{},
		nakedTripleCol{},
		nakedTripleRow{},
		nakedTripleBlock{},
	}
}

type nakedSingleTechnique struct {
	fillSolveTechnique
}

type hiddenSingleInRow struct {
	fillSolveTechnique
}

type hiddenSingleInCol struct {
	fillSolveTechnique
}

type hiddenSingleInBlock struct {
	fillSolveTechnique
}

type pointingPairRow struct {
	cullSolveTechnique
}

type pointingPairCol struct {
	cullSolveTechnique
}

type nakedPairCol struct {
	cullSolveTechnique
}

type nakedPairRow struct {
	cullSolveTechnique
}

type nakedPairBlock struct {
	cullSolveTechnique
}

type nakedTripleCol struct {
	cullSolveTechnique
}

type nakedTripleRow struct {
	cullSolveTechnique
}

type nakedTripleBlock struct {
	cullSolveTechnique
}

func (self fillSolveTechnique) IsFill() bool {
	return true
}

func (self cullSolveTechnique) IsFill() bool {
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
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that they can't be in any other cell in column %d", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Col())
}

func (self nakedPairCol) Find(grid *Grid) *SolveStep {
	colGetter := func(i int) CellList {
		return grid.Col(i)
	}
	return nakedSubset(self, 2, colGetter)
}

func (self nakedPairRow) Name() string {
	return "Naked pair row"
}

func (self nakedPairRow) Description(step *SolveStep) string {
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that they can't be in any other cell in row %d", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Row())
}

func (self nakedPairRow) Find(grid *Grid) *SolveStep {
	rowGetter := func(i int) CellList {
		return grid.Row(i)
	}
	return nakedSubset(self, 2, rowGetter)
}

func (self nakedPairBlock) Name() string {
	return "Naked pair block"
}

func (self nakedPairBlock) Description(step *SolveStep) string {
	if len(step.Nums) < 2 || len(step.PointerCells) < 2 {
		return ""
	}
	return fmt.Sprintf("%d and %d are only possible in (%d,%d) and (%d,%d), which means that they can't be in any other cell in block %d", step.Nums[0], step.Nums[1], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Block())
}

func (self nakedPairBlock) Find(grid *Grid) *SolveStep {
	blockGetter := func(i int) CellList {
		return grid.Block(i)
	}
	return nakedSubset(self, 2, blockGetter)
}

func (self nakedTripleCol) Name() string {
	return "Naked triple col"
}

func (self nakedTripleCol) Description(step *SolveStep) string {
	if len(step.Nums) < 3 || len(step.PointerCells) < 3 {
		return ""
	}
	return fmt.Sprintf("%d, %d, and %d are only possible in (%d,%d), (%d,%d) and (%d,%d), which means that they can't be in any other cell in column %d", step.Nums[0], step.Nums[1], step.Nums[2], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.PointerCells[2].Row+1, step.PointerCells[1].Col+1, step.TargetCells.Col())
}

func (self nakedTripleCol) Find(grid *Grid) *SolveStep {
	colGetter := func(i int) CellList {
		return grid.Col(i)
	}
	return nakedSubset(self, 3, colGetter)
}

func (self nakedTripleRow) Name() string {
	return "Naked triple row"
}

func (self nakedTripleRow) Description(step *SolveStep) string {
	if len(step.Nums) < 3 || len(step.PointerCells) < 3 {
		return ""
	}
	return fmt.Sprintf("%d, %d, and %d are only possible in (%d,%d), (%d, %d) and (%d,%d), which means that they can't be in any other cell in row %d", step.Nums[0], step.Nums[1], step.Nums[2], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.PointerCells[2].Row+1, step.PointerCells[2].Col+1, step.TargetCells.Row())
}

func (self nakedTripleRow) Find(grid *Grid) *SolveStep {
	rowGetter := func(i int) CellList {
		return grid.Row(i)
	}
	return nakedSubset(self, 3, rowGetter)
}

func (self nakedTripleBlock) Name() string {
	return "Naked triple block"
}

func (self nakedTripleBlock) Description(step *SolveStep) string {
	if len(step.Nums) < 3 || len(step.PointerCells) < 3 {
		return ""
	}
	return fmt.Sprintf("%d, %d and %d are only possible in (%d,%d), (%d,%d) and (%d,%d), which means that they can't be in any other cell in block %d", step.Nums[0], step.Nums[1], step.Nums[2], step.PointerCells[0].Row+1, step.PointerCells[0].Col+1, step.PointerCells[1].Row+1, step.PointerCells[1].Col+1, step.PointerCells[2].Row+1, step.PointerCells[2].Col+1, step.TargetCells.Block())
}

func (self nakedTripleBlock) Find(grid *Grid) *SolveStep {
	blockGetter := func(i int) CellList {
		return grid.Block(i)
	}
	return nakedSubset(self, 3, blockGetter)
}

func nakedSubset(technique SolveTechnique, k int, collectionGetter func(int) CellList) *SolveStep {
	//TODO: randomize order we visit things.
	for i := 0; i < DIM; i++ {

		groups := subsetCellsWithNPossibilities(k, collectionGetter(i))

		if len(groups) > 0 {
			//TODO: pick a random one
			group := groups[0]
			return &SolveStep{collectionGetter(i).RemoveCells(group), group, group.PossibilitiesUnion(), technique}
		}

	}
	return nil
}

func subsetCellsWithNPossibilities(k int, inputCells CellList) []CellList {
	//Given a list of cells (often a row, col, or block) and a target group size K,
	//returns a list of groups of cells of size K where the union of each group's possibility list
	//is size K.

	//Note: this function has performance O(n!/k!(n - k)!)

	//First, cull any cells with no possibilities to help minimize n
	cells := inputCells.FilterByHasPossibilities()

	var results []CellList

	for _, indexes := range subsetIndexes(len(cells), k) {
		//Build up set of all possibilties in this subset.
		subset := cells.Subset(indexes)
		union := subset.PossibilitiesUnion()
		//Okay, we built up the set. Is it the target size?
		if len(union) == k {
			results = append(results, subset)
		}
	}

	return results

}

func subsetIndexes(len int, size int) [][]int {
	//Sanity check
	if size > len {
		return nil
	}

	//returns an array of slices of size size that give you all of the subsets of a list of length len
	result := make([][]int, 0)
	counters := make([]int, size)
	for i, _ := range counters {
		counters[i] = i
	}
	for {
		innerResult := make([]int, size)
		for i, counter := range counters {
			innerResult[i] = counter
		}
		result = append(result, innerResult)
		//Now, increment.
		//Start at the end and try to increment each counter one.
		incremented := false
		for i := size - 1; i >= 0; i-- {

			counter := counters[i]
			if counter < len-(size-i) {
				//Found one!
				counters[i]++
				incremented = true
				if i < size-1 {
					//It was an inner counter; need to set all of the higher counters to one above the one to the left.
					base := counters[i] + 1
					for j := i + 1; j < size; j++ {
						counters[j] = base
						base++
					}
				}
				break
			}
		}
		//If we couldn't increment any, there's nothing to do.
		if !incremented {
			break
		}
	}
	return result
}

func weightsNormalized(weights []float64) bool {
	var sum float64
	for _, weight := range weights {
		sum += weight
	}
	return sum == 1.0
}

func normalizedWeights(weights []float64) []float64 {
	if weightsNormalized(weights) {
		return weights
	}
	var sum float64
	for _, weight := range weights {
		sum += weight
	}
	result := make([]float64, len(weights))
	for i, weight := range weights {
		result[i] = weight / sum
	}
	return result
}

func randomIndexWithWeights(weights []float64) int {
	//TODO: shouldn't this be called weightedRandomIndex?
	return randomIndexWithNormalizedWeights(normalizedWeights(weights))
}

func randomIndexWithNormalizedWeights(weights []float64) int {
	//assumes that weights is normalized--that is, weights all sum to 1.
	sample := rand.Float64()
	var counter float64
	for i, weight := range weights {
		counter += weight
		if sample <= weight {
			return i
		}
	}
	//This shouldn't happen if the weights are properly normalized.
	return len(weights) - 1
}

func (self *Grid) HumanSolve() SolveDirections {
	var results []*SolveStep
	numTechniques := len(techniques)

	//Note: trying these all in parallel is much slower (~15x) than doing them in sequence.
	//The reason is that in sequence we bailed early as soon as we found one step; now we try them all.

	for !self.Solved() {
		//TODO: provide hints to the techniques of where to look based on the last filled cell

		possibilitiesChan := make(chan *SolveStep)

		var possibilities []*SolveStep

		for _, technique := range techniques {
			go func(theTechnique SolveTechnique) {
				possibilitiesChan <- theTechnique.Find(self)
			}(technique)
		}

		//Collect all of the results

		for i := 0; i < numTechniques; i++ {
			possibility := <-possibilitiesChan
			if possibility != nil {
				possibilities = append(possibilities, possibility)
			}
		}

		//Now pick one to apply.
		if len(possibilities) == 0 {
			//Hmm, didn't find any possivbilities. We failed. :-(
			break
		}

		//TODO: pick the technique based on a weighting of how common a human is to pick each one.

		step := possibilities[rand.Intn(len(possibilities))]

		results = append(results, step)
		step.Apply(self)

	}
	if !self.Solved() {
		//We couldn't solve the puzzle.
		return nil
	}
	return results
}
