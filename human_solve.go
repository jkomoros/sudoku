package sudoku

import (
	"fmt"
	"log"
	"math"
	"strings"
)

//The actual techniques are intialized in hs_techniques.go, and actually defined in hst_*.go files.
var Techniques []SolveTechnique

//Worst case scenario, how many times we'd call HumanSolve to get a difficulty.
const MAX_DIFFICULTY_ITERATIONS = 50

//We will use this as our max to return a normalized difficulty.
//TODO: set this more accurately so we rarely hit it (it's very important to get this right!)
//This is just set emperically.
const MAX_RAW_DIFFICULTY = 18000.0

//How close we have to get to the average to feel comfortable our difficulty is converging.
const DIFFICULTY_CONVERGENCE = 0.0005

type SolveDirections []*SolveStep

type SolveStep struct {
	TargetCells  CellList
	PointerCells CellList
	Nums         IntSlice
	Technique    SolveTechnique
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
