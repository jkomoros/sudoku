package sudoku

import (
	"fmt"
	"math"
)

//The actual techniques are intialized in hs_techniques.go, and actually defined in hst_*.go files.
//Techniques is ALL technies. CheapTechniques is techniques that are reasonably cheap to compute.
//ExpensiveTechniques is techniques that should only be used if all else has failed.
var Techniques []SolveTechnique
var CheapTechniques []SolveTechnique
var ExpensiveTechniques []SolveTechnique

var GuessTechnique SolveTechnique

//EVERY technique, even the weird one like Guess
var AllTechniques []SolveTechnique

//Worst case scenario, how many times we'd call HumanSolve to get a difficulty.
const MAX_DIFFICULTY_ITERATIONS = 50

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

func (self *Grid) HumanWalkthrough() string {
	steps := self.HumanSolution()
	return steps.Walkthrough(self)
}

func (self *Grid) HumanSolution() SolveDirections {
	clone := self.Copy()
	defer clone.Done()
	return clone.HumanSolve()
}

/*
 * The HumanSolve method is very complex due to guessing logic.
 *
 * Without guessing, the approach is very straightforward. Every move either fills a cell
 * or removes possibilities. But nothing does anything contradictory, so if they diverge
 * in path, it doesn't matter--they're still working towards the same end state (denoted by @)
 *
 *
 *
 *                   |
 *                  /|\
 *                 / | \
 *                |  |  |
 *                \  |  /
 *                 \ | /
 *                  \|/
 *                   |
 *                   V
 *                   @
 *
 *
 * In human solve, we first try the cheap techniques, and if we can't find enough options, we then additionally try
 * the expensive set of techniques. But both cheap and expensive techniques are similar in that they move us
 * towards the end state.
 *
 * For simplicity, we'll just show paths like this as a single line, even though realistically they could diverge arbitrarily.
 *
 * This all changes when you introduce branching, because at a branch point you could have chosen the wrong path
 * and at some point down that path you will discover an invalidity, which tells you you chose wrong, and
 * you'll have to unwind.
 *
 * Let's explore a puzzle that needs one branch point.
 *
 * We explore with normal techniques until we run into a point where none of hte normal techinques work.
 * This is a DIRE point, and in some cases we might just give up. But we have one last thing to try:
 * branching.
 * We then run the guess technique, which proposes multiple guess steps (big O's) that we could take.
 *
 * The technique will choose cells with only a small number of possibilities, to reduce the branching factor.
 *
 *                  |
 *                  |
 *                  V
 *                  O O O O O ...
 *
 * We will randomly pick one, and then explore all of its possibilities.
 * CRUCIALLY, at a branch point, we never have to pick another cell to explore its possibilities; for each cell,
 * if you plug in each of the possibilites and solve forward, it must result in either an invalidity (at which
 * point you try another possibility, or if they're all gone you unwind if there's a branch point above), or
 * you picked correctly and the solution lies that way. But it's never the case that picking THIS cell won't uncover
 * either the invalidity or the solution.
 * So in reality, when we come to a branch point, we can choose one cell to focus on and throw out all of the others.
 *
 *                  |
 *                  |
 *                  V
 *                  O
 *
 * But within that cell, there are multiple possibilty branches to consider.
 *
 *
 *                  |
 *                  |
 *                  V
 *                  O
 *                 / \
 *                1   3
 *               /     \
 *              |       |
 *
 * We go through each in turn and play forward until we find either an invalidity or a solution.
 * Within each branch, we use the normal techniques as normal--remember it's actually branching but
 * converging, like in the first diagram.
 *
 *                  |
 *                  |
 *                  V
 *                  O
 *                 / \
 *                1   3
 *               /     \
 *              |       |
 *              X       @
 *
 * When we uncover an invalidity, we unwind back to the branch point and then try the next possibility.
 * We should never have to unwind above the top branch, because down one of the branches (possibly somewhere deep)
 * There MUST be a solution (assuming the puzzle is valid)
 * Obviously if we find the solution on our branch, we're good.
 *
 * But what happens if we run out of normal techinques down one of our branches and have to branch again?
 *
 * Nothing much changes, except that you DO unravel if you uncover that all of the possibilities down this
 * side lead to invalidities. You just never unravel past the first branch point.
 *
 *                  |
 *                  |
 *                  V
 *                  O
 *                 / \
 *                1   3
 *               /     \
 *              |       |
 *              O       O
 *             / \     / \
 *            4   5   6   7
 *           /    |   |    \
 *          |     |   |     |
 *          X     X   X     @
 *
 * Down one of the paths MUST lie a solution.
 *
 * The search will fail if we have a max depth limit of branching to try, because then we might not discover a
 * solution down one of the branches. A good sanity point is DIM*DIM branch points is the absolute highest; an
 * assert at that level makes sense.
 *
 * In this implementation, humanSolveHelper does the work of exploring any branch up to a point where a guess must happen.
 * If we run out of ideas on a branch, we call into guess helper, which will pick a guess and then try all of the versions of it
 * until finding one that works. This keeps humanSolveHelper pretty straighforward and keeps most of the complex guess logic out.
 */

func (self *Grid) HumanSolve() SolveDirections {

	//Short circuit solving if it has multiple solutions.
	if self.HasMultipleSolutions() {
		return nil
	}

	return humanSolveHelper(self)
}

//Do we even need a helper here? Can't we just make HumanSolve actually humanSolveHelper?
//The core worker of human solve, it does all of the solving between branch points.
func humanSolveHelper(grid *Grid) []*SolveStep {

	var results []*SolveStep

	//Note: trying these all in parallel is much slower (~15x) than doing them in sequence.
	//The reason is that in sequence we bailed early as soon as we found one step; now we try them all.

	for !grid.Solved() {
		//TODO: provide hints to the techniques of where to look based on the last filled cell

		if grid.Invalid() {
			//We much have been in a branch and found an invalidity.
			//Bail immediately.
			return nil
		}

		possibilities := runTechniques(CheapTechniques, grid)

		//Now pick one to apply.
		if len(possibilities) == 0 {
			//Okay, let's try the ExpensiveTechniques, as a hail mary.
			possibilities = runTechniques(ExpensiveTechniques, grid)
			if len(possibilities) == 0 {
				//Hmm, didn't find any possivbilities. We failed. :-(
				break
			}
		}

		//TODO: consider if we should stop picking techniques based on their weight here.
		//Now that Find returns a slice instead of a single, we're already much more likely to select an "easy" technique. ... Right?

		possibilitiesWeights := make([]float64, len(possibilities))
		for i, possibility := range possibilities {
			possibilitiesWeights[i] = possibility.Technique.HumanLikelihood()
		}
		step := possibilities[randomIndexWithInvertedWeights(possibilitiesWeights)]

		results = append(results, step)
		step.Apply(grid)

	}
	if !grid.Solved() {
		//We couldn't solve the puzzle.
		//But let's do one last ditch effort and try guessing.
		guessSteps := humanSolveGuess(grid)
		if len(guessSteps) == 0 {
			//Okay, we just totally failed.
			return nil
		}
		return append(results, guessSteps...)
	}
	return results
}

//Called when we have run out of options at a given state and need to guess.
func humanSolveGuess(grid *Grid) []*SolveStep {

	//TODO: consider doing a normal solve forward from here to figure out what the right branch is and just do that.
	guesses := GuessTechnique.Find(grid)

	if len(guesses) == 0 {
		//Coludn't find a guess step, oddly enough.
		return nil
	}

	//Take just the first guess step and forget about the other ones.
	guess := guesses[0]

	//The guess technique passes back the other nums as PointerNums, which is a hack.
	//Unpack them and then nil it out to prevent confusing other people in the future with them.
	otherNums := guess.PointerNums
	guess.PointerNums = nil

	var gridCopy *Grid

	for {
		gridCopy = grid.Copy()

		guess.Apply(gridCopy)

		solveSteps := humanSolveHelper(gridCopy)

		if len(solveSteps) != 0 {
			//Success!
			//Make ourselves look like that grid (to pass back the state of what the solution was) and return.
			grid.replace(gridCopy)
			gridCopy.Done()
			return append([]*SolveStep{guess}, solveSteps...)
		}
		//We need to try the next solution.

		if len(otherNums) == 0 {
			//No more numbers to try. We failed!
			break
		}

		nextNum := otherNums[0]
		otherNums = otherNums[1:]

		//Stuff it into the TargetNums for the branch step.
		guess.TargetNums = IntSlice{nextNum}

		gridCopy.Done()

	}

	gridCopy.Done()

	//We failed to find anything (which should never happen...)
	return nil

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
	//This is so expensive and during testing we don't care if converges.
	//So we split out the meat of the method separately.
	return self.calcluateDifficulty(true)
}

func (self *Grid) calcluateDifficulty(accurate bool) float64 {
	//This can be an extremely expensive method. Do not call repeatedly!
	//returns the difficulty of the grid, which is a number between 0.0 and 1.0.
	//This is a probabilistic measure; repeated calls may return different numbers, although generally we wait for the results to converge.

	//We solve the same puzzle N times, then ask each set of steps for their difficulty, and combine those to come up with the overall difficulty.

	accum := 0.0
	average := 0.0
	lastAverage := 0.0

	//Since this is so expensive, in testing situations we want to run it in less accurate mode (so it goes fast!)
	maxIterations := MAX_DIFFICULTY_ITERATIONS
	if !accurate {
		maxIterations = 5.0
	}

	for i := 0; i < maxIterations; i++ {
		grid := self.Copy()
		steps := grid.HumanSolve()
		difficulty := steps.Difficulty()

		accum += difficulty
		average = accum / (float64(i) + 1.0)

		if math.Abs(average-lastAverage) < DIFFICULTY_CONVERGENCE {
			//Okay, we've already converged. Just return early!
			grid.Done()
			return average
		}

		lastAverage = average
		grid.Done()
	}

	//We weren't converging... oh well!
	return average
}
