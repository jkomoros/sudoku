package sudoku

import (
	"fmt"
	"log"
	"math"
	"strings"
)

//The actual techniques are intialized in hs_techniques.go, and actually defined in hst_*.go files.
//Techniques is ALL technies. CheapTechniques is techniques that are reasonably cheap to compute.
//ExpensiveTechniques is techniques that should only be used if all else has failed.
var Techniques []SolveTechnique
var CheapTechniques []SolveTechnique
var ExpensiveTechniques []SolveTechnique

//Worst case scenario, how many times we'd call HumanSolve to get a difficulty.
const MAX_DIFFICULTY_ITERATIONS = 50

//This number is the 'Constant' term from the multiple linear regression to learn the weights.
var difficultyConstant float64

//How close we have to get to the average to feel comfortable our difficulty is converging.
const DIFFICULTY_CONVERGENCE = 0.0005

type SolveDirections []*SolveStep

type SolveStep struct {
	//The cells that will be affected by the techinque
	TargetCells CellList
	//The cells that together lead the techinque to being valid
	PointerCells CellList
	//The numbers we will remove (or, in the case of Fill, add)
	//TODO: shouldn't this be renamed TargetNums?
	TargetNums IntSlice
	//The numbers in pointerCells that lead us to remove TargetNums from TargetCells.
	//This is only very rarely needed (at this time only for hiddenSubset techniques)
	PointerNums IntSlice
	//The general technique that underlies this step.
	Technique SolveTechnique
}

func (self *SolveStep) IsUseful(grid *Grid) bool {
	//Returns true IFF calling Apply with this step and the given grid would result in some useful work. Does not modify the gri.d

	//All of this logic is substantially recreated in Apply.

	if self.Technique == nil {
		return false
	}

	//TODO: test this.
	if self.Technique.IsFill() {
		if len(self.TargetCells) == 0 || len(self.TargetNums) == 0 {
			return false
		}
		cell := self.TargetCells[0].InGrid(grid)
		return self.TargetNums[0] != cell.Number()
	} else {
		useful := false
		for _, cell := range self.TargetCells {
			gridCell := cell.InGrid(grid)
			for _, exclude := range self.TargetNums {
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
		if len(self.TargetCells) == 0 || len(self.TargetNums) == 0 {
			return
		}
		cell := self.TargetCells[0].InGrid(grid)
		cell.SetNumber(self.TargetNums[0])
	} else {
		for _, cell := range self.TargetCells {
			gridCell := cell.InGrid(grid)
			for _, exclude := range self.TargetNums {
				gridCell.setExcluded(exclude, true)
			}
		}
	}
}

func (self *SolveStep) Description() string {
	result := ""
	if self.Technique.IsFill() {
		result += fmt.Sprintf("We put %s in cell %s ", self.TargetNums.Description(), self.TargetCells.Description())
	} else {
		//TODO: pluralize based on length of lists.
		result += fmt.Sprintf("We remove the possibilities %s from cells %s ", self.TargetNums.Description(), self.TargetCells.Description())
	}
	result += "because " + self.Technique.Description(self) + "."
	return result
}

func (self *SolveStep) normalize() {
	//Puts the solve step in its normal status. In practice this means that the various slices are sorted, so that the Description of them is stable.
	self.PointerCells.Sort()
	self.TargetCells.Sort()
	self.TargetNums.Sort()
	self.PointerNums.Sort()
}

func (self SolveDirections) Stats() []string {
	//TODO: test this.
	techniqueCount := make(map[string]int)
	for _, step := range self {
		techniqueCount[step.Technique.Name()] += 1
	}
	var result []string

	//TODO: use a standard divider across the codebase
	divider := "-------------------------"

	result = append(result, divider)
	result = append(result, fmt.Sprintf("Difficulty : %f", self.Difficulty()))
	result = append(result, divider)

	//We want a stable ordering for technique counts.
	for _, technique := range Techniques {
		result = append(result, fmt.Sprintf("%s : %d", technique.Name(), techniqueCount[technique.Name()]))
	}

	result = append(result, divider)

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

	//This method assumes the weights have been calibrated empirically to give scores between 0.0 and 1.0
	//without normalization here.

	if len(self) == 0 {
		//The puzzle was not able to be solved, apparently.
		return 1.0
	}

	accum := difficultyConstant
	for _, step := range self {
		accum += step.Technique.Difficulty()
	}

	if accum < 0.0 {
		log.Println("Accumuldated difficulty snapped to 0.0:", accum, strings.Join(self.Stats(), "\n"))
		accum = 0.0
	}

	if accum > 1.0 {
		log.Println("Accumulated difficulty snapped to 1.0:", accum, strings.Join(self.Stats(), "\n"))
	}

	return accum
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

	//Note: trying these all in parallel is much slower (~15x) than doing them in sequence.
	//The reason is that in sequence we bailed early as soon as we found one step; now we try them all.

	for !self.Solved() {
		//TODO: provide hints to the techniques of where to look based on the last filled cell

		possibilities := runTechniques(CheapTechniques, self)

		//Now pick one to apply.
		if len(possibilities) == 0 {
			//Okay, let's try the ExpensiveTechniques, as a hail mary.
			possibilities = runTechniques(ExpensiveTechniques, self)
			if len(possibilities) == 0 {
				//Hmm, didn't find any possivbilities. We failed. :-(
				break
			}
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

func runTechniques(techniques []SolveTechnique, grid *Grid) []*SolveStep {
	numTechniques := len(techniques)
	possibilitiesChan := make(chan []*SolveStep)

	var possibilities []*SolveStep

	for _, technique := range techniques {
		go func(theTechnique SolveTechnique) {
			possibilitiesChan <- theTechnique.Find(grid)
		}(technique)
	}

	//Collect all of the results
	for i := 0; i < numTechniques; i++ {
		for _, possibility := range <-possibilitiesChan {
			possibilities = append(possibilities, possibility)
		}
	}

	return possibilities
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
