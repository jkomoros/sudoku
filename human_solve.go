package sudoku

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
)

//Worst case scenario, how many times we'd call HumanSolve to get a difficulty.
const MAX_DIFFICULTY_ITERATIONS = 50

//We will use this as our max to return a normalized difficulty.
//TODO: set this more accurately so we rarely hit it (it's very important to get this right!)
//This is just set emperically.
const MAX_RAW_DIFFICULTY = 18000.0

//How close we have to get to the average to feel comfortable our difficulty is converging.
const DIFFICULTY_CONVERGENCE = 0.0005

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
	//IMPORTANT: a step should return a step IFF that step is valid AND the step would cause useful work to be done if applied.

	//NOTE: this is a critical weakness, because it allows each technique to find only one step, even though multiple might apply.
	//However, HumanSolve assumes that we will pick a step randomly at any point based on its difficulty proportion. Because we will only
	//have at max 1 of easy (and more likely) techniques, this will systematically over-prefer more complex techniques.
	//TODO: fix this.
	Find(*Grid) *SolveStep
	IsFill() bool
	//How difficult a real human would say this technique is. Generally inversely related to how often a real person would pick it. 0.0 to 1.0.
	Difficulty() float64
}

type basicSolveTechnique struct {
	name       string
	isFill     bool
	difficulty float64
}

var Techniques []SolveTechnique

func init() {

	//TODO: calculate more realistic weights.

	Techniques = []SolveTechnique{
		hiddenSingleInRow{
			basicSolveTechnique{
				"Necessary In Row",
				true,
				0.0,
			},
		},
		hiddenSingleInCol{
			basicSolveTechnique{
				"Necessary In Col",
				true,
				0.0,
			},
		},
		hiddenSingleInBlock{
			basicSolveTechnique{
				"Necessary In Block",
				true,
				0.0,
			},
		},
		nakedSingleTechnique{
			basicSolveTechnique{
				"Only Legal Number",
				true,
				5.0,
			},
		},
		pointingPairRow{
			basicSolveTechnique{
				"Pointing Pair Row",
				false,
				25.0,
			},
		},
		pointingPairCol{
			basicSolveTechnique{
				"Pointing Pair Col",
				false,
				25.0,
			},
		},
		nakedPairCol{
			basicSolveTechnique{
				"Naked Pair Col",
				false,
				75.0,
			},
		},
		nakedPairRow{
			basicSolveTechnique{
				"Naked Pair Row",
				false,
				75.0,
			},
		},
		nakedPairBlock{
			basicSolveTechnique{
				"Naked Pair Block",
				false,
				85.0,
			},
		},
		nakedTripleCol{
			basicSolveTechnique{
				"Naked Triple Col",
				false,
				125.0,
			},
		},
		nakedTripleRow{
			basicSolveTechnique{
				"Naked Triple Row",
				false,
				125.0,
			},
		},
		nakedTripleBlock{
			basicSolveTechnique{
				"Naked Triple Block",
				false,
				140.0,
			},
		},
	}
}

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

type pointingPairRow struct {
	basicSolveTechnique
}

type pointingPairCol struct {
	basicSolveTechnique
}

type nakedPairCol struct {
	basicSolveTechnique
}

type nakedPairRow struct {
	basicSolveTechnique
}

type nakedPairBlock struct {
	basicSolveTechnique
}

type nakedTripleCol struct {
	basicSolveTechnique
}

type nakedTripleRow struct {
	basicSolveTechnique
}

type nakedTripleBlock struct {
	basicSolveTechnique
}

func (self basicSolveTechnique) Name() string {
	return self.name
}

func (self basicSolveTechnique) IsFill() bool {
	return self.isFill
}

func (self basicSolveTechnique) Difficulty() float64 {
	return self.difficulty
}

func newFillSolveStep(cell *Cell, num int, technique SolveTechnique) *SolveStep {
	cellArr := []*Cell{cell}
	numArr := []int{num}
	return &SolveStep{cellArr, nil, numArr, technique}
}

func (self *SolveStep) IsUseful(grid *Grid) bool {
	//Returns true IFF calling Apply with this step and the given grid would result in some useful work. Does not modify the gri.d

	//All of this logic is substantially recreated in Apply.

	if self.Technique == nil {
		return false
	}

	//TODO: test this.
	if self.Technique.IsFill() {
		if len(self.TargetCells) == 0 || len(self.Nums) == 0 {
			return false
		}
		cell := self.TargetCells[0].InGrid(grid)
		return self.Nums[0] != cell.Number()
	} else {
		useful := false
		for _, cell := range self.TargetCells {
			gridCell := cell.InGrid(grid)
			for _, exclude := range self.Nums {
				//It's right to use Possible because it includes the logic of "it's not possible if there's a number in there already"
				//TODO: ensure the comment above is correct logically.
				if gridCell.Possible(exclude) {
					useful = true
				}
			}
		}
		return useful
	}
}

func (self *SolveStep) Apply(grid *Grid) {
	//All of this logic is substantially recreated in IsUseful.
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

func (self *SolveStep) Description() string {
	result := ""
	if self.Technique.IsFill() {
		result += fmt.Sprintf("We put %s in cell %s ", self.Nums.Description(), self.TargetCells.Description())
	} else {
		//TODO: pluralize based on length of lists.
		result += fmt.Sprintf("We remove the possibilities %s from cells %s ", self.Nums.Description(), self.TargetCells.Description())
	}
	result += "because " + self.Technique.Description(self) + "."
	return result
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
	result := newFillSolveStep(cell, cell.implicitNumber(), self)
	if result.IsUseful(grid) {
		return result
	}
	return nil
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

	var result *SolveStep

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
						result = newFillSolveStep(cell, index+1, technique)
						if result.IsUseful(grid) {
							return result
						}
						//Hmm, wasn't useful. Keep trying...
					}
				}
			}
		}
	}
	//Nope.
	return nil
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

	var result *SolveStep

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
				result = &SolveStep{grid.Row(cells.Row()).RemoveCells(block), cells, []int{num + 1}, self}
				if result.IsUseful(grid) {
					return result
				}
				//Hmm, guess it found some not-actually useful thing. Keep looking.
			}
		}
	}
	return nil
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

	var result *SolveStep

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
				result = &SolveStep{grid.Col(cells.Col()).RemoveCells(block), cells, []int{num + 1}, self}
				if result.IsUseful(grid) {
					return result
				}
				//Hmm, guess it found some not-actually useful thing. Keep looking.
			}
		}
	}
	return nil
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
	return nakedSubset(grid, self, 2, colGetter)
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
	return nakedSubset(grid, self, 2, rowGetter)
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
	return nakedSubset(grid, self, 2, blockGetter)
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
	return nakedSubset(grid, self, 3, colGetter)
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
	return nakedSubset(grid, self, 3, rowGetter)
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
	return nakedSubset(grid, self, 3, blockGetter)
}

func nakedSubset(grid *Grid, technique SolveTechnique, k int, collectionGetter func(int) CellList) *SolveStep {
	//TODO: randomize order we visit things.
	var result *SolveStep
	for i := 0; i < DIM; i++ {

		groups := subsetCellsWithNPossibilities(k, collectionGetter(i))

		if len(groups) > 0 {
			//TODO: pick a random one instead of the first useful one.

			for _, group := range groups {

				result = &SolveStep{collectionGetter(i).RemoveCells(group), group, group.PossibilitiesUnion(), technique}
				if result.IsUseful(grid) {
					return result
				}
				//Hmm, it's not actually useful. Keep going.
			}
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
		if weight < 0 {
			return false
		}
		sum += weight
	}
	return sum == 1.0
}

func normalizedWeights(weights []float64) []float64 {
	if weightsNormalized(weights) {
		return weights
	}
	var sum float64
	var lowestNegative float64

	//Check for a negative.
	for _, weight := range weights {
		if weight < 0 {
			if weight < lowestNegative {
				lowestNegative = weight
			}
		}
	}

	fixedWeights := make([]float64, len(weights))

	copy(fixedWeights, weights)

	if lowestNegative != 0 {
		//Found a negative. Need to fix up all weights.
		for i := 0; i < len(weights); i++ {
			fixedWeights[i] = weights[i] - lowestNegative
		}

	}

	for _, weight := range fixedWeights {
		sum += weight
	}

	result := make([]float64, len(weights))
	for i, weight := range fixedWeights {
		result[i] = weight / sum
	}
	return result
}

func randomIndexWithInvertedWeights(invertedWeights []float64) int {
	normalizedInvertedWeights := normalizedWeights(invertedWeights)
	normalizedWeights := make([]float64, len(invertedWeights))
	for i, weight := range normalizedInvertedWeights {
		normalizedWeights[i] = 1.0 - weight
	}
	return randomIndexWithNormalizedWeights(normalizedWeights)
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

func (self SolveDirections) Description() []string {

	if len(self) == 0 {
		return []string{""}
	}

	descriptions := make([]string, len(self))

	for i, step := range self {
		intro := ""
		switch i {
		case 0:
			intro = "First, "
		case len(self) - 1:
			intro = "Finally, "
		default:
			//TODO: switch between "then" and "next" randomly.
			intro = "Next, "
		}
		descriptions[i] = intro + strings.ToLower(step.Description())

	}
	return descriptions
}

func (self SolveDirections) Difficulty() float64 {
	//How difficult the solve directions described are. The measure of difficulty we use is
	//just summing up weights we see; this captures:
	//* Number of steps
	//* Average difficulty of steps
	//* Number of hard steps
	//* (kind of) the hardest step: because the difficulties go up expontentionally.

	//TODO: what's a good max bound for difficulty? This should be normalized to 0<->1 based on that.

	accum := 0.0
	for _, step := range self {
		accum += step.Technique.Difficulty()
	}

	if accum > MAX_RAW_DIFFICULTY {
		log.Println("Accumulated difficulty exceeded max difficulty: ", accum)
		accum = MAX_RAW_DIFFICULTY
	}

	return accum / MAX_RAW_DIFFICULTY

}

func (self SolveDirections) Walkthrough(grid *Grid) string {

	//TODO: test this.

	clone := grid.Copy()
	defer clone.Done()

	DIVIDER := "\n\n--------------------------------------------\n\n"

	intro := fmt.Sprintf("This will take %d steps to solve.", len(self))

	intro += "\nWhen you start, your grid looks like this:\n"

	intro += clone.Diagram()

	intro += "\n"

	intro += DIVIDER

	descriptions := self.Description()

	results := make([]string, len(self))

	for i, description := range descriptions {

		result := description + "\n"
		result += "After doing that, your grid will look like: \n\n"

		self[i].Apply(clone)

		result += grid.Diagram()

		results[i] = result
	}

	return intro + strings.Join(results, DIVIDER) + DIVIDER + "Now the puzzle is solved."
}

func (self *Grid) HumanWalkthrough() string {
	steps := self.HumanSolution()
	return steps.Walkthrough(self)
}

func (self *Grid) HumanSolution() SolveDirections {
	clone := self.Copy()
	defer clone.Done()
	return clone.HumanSolve()
}

func (self *Grid) HumanSolve() SolveDirections {
	var results []*SolveStep
	numTechniques := len(Techniques)

	//Note: trying these all in parallel is much slower (~15x) than doing them in sequence.
	//The reason is that in sequence we bailed early as soon as we found one step; now we try them all.

	for !self.Solved() {
		//TODO: provide hints to the techniques of where to look based on the last filled cell

		possibilitiesChan := make(chan *SolveStep)

		var possibilities []*SolveStep

		for _, technique := range Techniques {
			go func(theTechnique SolveTechnique) {
				possibilitiesChan <- theTechnique.Find(self)
			}(technique)
		}

		//Collect all of the results

		for i := 0; i < numTechniques; i++ {
			possibility := <-possibilitiesChan

			if possibility != nil {
				if possibility.IsUseful(self) {
					possibilities = append(possibilities, possibility)
				} else {
					log.Println("Rejecting a not useful suggestion: ", possibility)
				}
			}
		}

		//Now pick one to apply.
		if len(possibilities) == 0 {
			//Hmm, didn't find any possivbilities. We failed. :-(
			break
		}

		possibilitiesWeights := make([]float64, len(possibilities))
		for i, possibility := range possibilities {
			possibilitiesWeights[i] = possibility.Technique.Difficulty()
		}
		step := possibilities[randomIndexWithInvertedWeights(possibilitiesWeights)]

		results = append(results, step)
		step.Apply(self)

	}
	if !self.Solved() {
		//We couldn't solve the puzzle.
		return nil
	}
	return results
}

func (self *Grid) Difficulty() float64 {
	//This can be an extremely expensive method. Do not call repeatedly!
	//returns the difficulty of the grid, which is a number between 0.0 and 1.0.
	//This is a probabilistic measure; repeated calls may return different numbers, although generally we wait for the results to converge.

	//We solve the same puzzle N times, then ask each set of steps for their difficulty, and combine those to come up with the overall difficulty.

	accum := 0.0
	average := 0.0
	lastAverage := 0.0

	for i := 0; i < MAX_DIFFICULTY_ITERATIONS; i++ {
		grid := self.Copy()
		steps := grid.HumanSolve()
		difficulty := steps.Difficulty()

		accum += difficulty
		average = accum / (float64(i) + 1.0)

		if math.Abs(average-lastAverage) < DIFFICULTY_CONVERGENCE {
			//Okay, we've already converged. Just return early!
			return average
		}

		lastAverage = average
	}

	//We weren't converging... oh well!
	return average

}
