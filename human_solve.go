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

	//Find returns as many steps as it can find in the grid for that technique. This helps ensure that when we pick a step,
	//it's more likely to be an "easy" step because there will be more of them at any time.
	Find(*Grid) []*SolveStep
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
		hiddenPairRow{
			basicSolveTechnique{
				"Hidden Pair Row",
				false,
				300.0,
			},
		},
		hiddenPairCol{
			basicSolveTechnique{
				"Hidden Pair Col",
				false,
				300.0,
			},
		},
		hiddenPairBlock{
			basicSolveTechnique{
				"Hidden Pair Block",
				false,
				250.0,
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
	if len(step.TargetCells) == 0 || len(step.Nums) == 0 {
		return ""
	}
	cell := step.TargetCells[0]
	num := step.Nums[0]
	return fmt.Sprintf("%d is required in the %d row, and %d is the only column it fits", num, cell.Row+1, cell.Col+1)
}

func (self hiddenSingleInRow) Find(grid *Grid) []*SolveStep {
	//TODO: test that if there are multiple we find them both.
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

func (self hiddenSingleInCol) Find(grid *Grid) []*SolveStep {
	//TODO: test this will find multiple if they exist.
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

func (self hiddenSingleInBlock) Find(grid *Grid) []*SolveStep {
	//TODO: Verify we find multiples if they exist.
	getter := func(index int) []*Cell {
		return grid.Block(index)
	}
	return necessaryInCollection(grid, self, getter)
}

func necessaryInCollection(grid *Grid, technique SolveTechnique, collectionGetter func(index int) []*Cell) []*SolveStep {
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

func (self pointingPairRow) Description(step *SolveStep) string {
	if len(step.Nums) == 0 {
		return ""
	}
	return fmt.Sprintf("%d is only possible in row %d of block %d, which means it can't be in any other cell in that row not in that block", step.Nums[0], step.TargetCells.Row(), step.PointerCells.Block())
}

func (self pointingPairRow) Find(grid *Grid) []*SolveStep {
	//Within each block, for each number, see if all items that allow it are aligned in a row or column.
	//TODO: randomize order of blocks.
	//TODO: this is substantially duplicated in pointingPaircol
	//TODO: test this returns multiple if they exist.

	var results []*SolveStep

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
				result := &SolveStep{grid.Row(cells.Row()).RemoveCells(block), cells, []int{num + 1}, self}
				if result.IsUseful(grid) {
					results = append(results, result)
				}
				//Keep looking for more!
			}
		}
	}
	return results
}

func (self pointingPairCol) Description(step *SolveStep) string {
	if len(step.Nums) == 0 {
		return ""
	}
	return fmt.Sprintf("%d is only possible in column %d of block %d, which means it can't be in any other cell in that column not in that block", step.Nums[0], step.TargetCells.Col(), step.PointerCells.Block())
}

func (self pointingPairCol) Find(grid *Grid) []*SolveStep {
	//Within each block, for each number, see if all items that allow it are aligned in a row or column.
	//TODO: randomize order of blocks.
	//TODO: this is substantially duplicated in pointingPairRow
	//TODO: test this will find multiple if they exist.

	var results []*SolveStep

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
				result := &SolveStep{grid.Col(cells.Col()).RemoveCells(block), cells, []int{num + 1}, self}
				if result.IsUseful(grid) {
					results = append(results, result)
				}
				//Keep looking!
			}
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

		possibilitiesChan := make(chan []*SolveStep)

		var possibilities []*SolveStep

		for _, technique := range Techniques {
			go func(theTechnique SolveTechnique) {
				possibilitiesChan <- theTechnique.Find(self)
			}(technique)
		}

		//Collect all of the results

		for i := 0; i < numTechniques; i++ {

			for _, possibility := range <-possibilitiesChan {
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

		//TODO: consider if we should stop picking techniques based on their weight here.
		//Now that Find returns a slice instead of a single, we're already much more likely to select an "easy" technique. ... Right?

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
