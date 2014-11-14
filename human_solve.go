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

var GuessTechnique SolveTechnique

//EVERY technique, even the weird one like Guess
var AllTechniques []SolveTechnique

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
	result = append(result, fmt.Sprintf("Step count: %d", len(self)))
	result = append(result, divider)

	//We want a stable ordering for technique counts.
	for _, technique := range AllTechniques {
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
		log.Println("Accumuldated difficulty snapped to 0.0:", accum)
		accum = 0.0
	}

	if accum > 1.0 {
		log.Println("Accumulated difficulty snapped to 1.0:", accum)
		accum = 1.0
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

		result += clone.Diagram()

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

type branchPoint struct {
	//The point at which we branched for a guess. non-nil if we are in a branch.
	grid *Grid
	//The step we'll apply to get us into the branch point.
	step *SolveStep
	//Ther other numbers to try from the branch point.
	otherNums IntSlice
	//The steps we've taken since the branch point.
	branchSteps []*SolveStep
	//The earlier branch point
	previousBranchPoint *branchPoint
	nextBranchPoint     *branchPoint
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
 */

//TODO: guessing logic will be much cleaner if we just a recursive helper.

func (self *Grid) HumanSolve() SolveDirections {

	var results []*SolveStep

	var branch *branchPoint

	//Note: trying these all in parallel is much slower (~15x) than doing them in sequence.
	//The reason is that in sequence we bailed early as soon as we found one step; now we try them all.

	for !self.Solved() {

		var possibilities []*SolveStep

		if branch != nil && self.Invalid() {
			//We're in a branch, and got to a point where we found an invalidity.
			//We chose the wrong branch. We should unwind to the branch point and go down the other branch.

			//Unwind ourselves
			self.Load(branch.grid.DataString())

			//Throwout the steps down the wrong branch we took.
			branch.branchSteps = nil

			if len(branch.otherNums) > 0 {

				//Pop off the nextNum to do
				nextNum := branch.otherNums[0]
				branch.otherNums = branch.otherNums[1:]

				//Stuff it into the TargetNums for the branch step.
				branch.step.TargetNums = IntSlice{nextNum}

				//Stuff the possibility list with the mangled branch.step.
				possibilities = []*SolveStep{branch.step}
			} else {
				//Well, crap. We're out of luck, nothing more for us to do.
				if branch.previousBranchPoint == nil {
					//We're at the first branchpoint. This really shouldn't have happened... oh well.
					break
				} else {
					//Roll back up one level of branching, apparently everything below us leads to invalidity.
					branch = branch.previousBranchPoint
					branch.nextBranchPoint = nil

					//Leave the puzzle in an invalid state; this will cause the next run through to load us
					//back up the next number and continue.
					continue

				}
			}

		} else {

			//Normal operation; get potential solve steps by running them all.
			possibilities = runTechniques(CheapTechniques, self)
		}

		//TODO: provide hints to the techniques of where to look based on the last filled cell

		//Now pick one to apply.
		if len(possibilities) == 0 {
			//Okay, let's try the ExpensiveTechniques, as a hail mary.
			possibilities = runTechniques(ExpensiveTechniques, self)
			if len(possibilities) == 0 {
				//Hmm, didn't find any possivbilities. We're getting to be out of options...

				//Try to guess as a hail mary
				possibilities = runTechniques([]SolveTechnique{GuessTechnique}, self)

				if len(possibilities) == 0 {
					//Okay, we're well and truly done--not even any guesses came up with something. Nothing we can do.
					break
				} else {
					//Yay, found something! remember the branch point, so we can jump back to it.

					//Push new branch point onto the doubly-linked list of branch points
					newBranch := &branchPoint{
						previousBranchPoint: branch,
					}
					if branch != nil {
						branch.nextBranchPoint = newBranch
					}
					branch = newBranch

					//We're just going to choose the first one.
					possibilities = possibilities[0:1]

					//TODO: this doesn't hold our special excludes, which we might have worked quite a bit to set up.
					//Ideally we'd have a way to keep those overrides.
					//The worst case is that we have a few unnecessary Cull steps just before the branch point.
					branch.grid = self.Copy()
					branch.step = possibilities[0]
					branch.otherNums = branch.step.PointerNums

					//Null out the branchPointStep's pointerNums; their only point was to communicate out the other possibilities.
					//And from now on they'll just be confusing.
					branch.step.PointerNums = nil

				}
			}
		}

		//TODO: consider if we should stop picking techniques based on their weight here.
		//Now that Find returns a slice instead of a single, we're already much more likely to select an "easy" technique. ... Right?

		possibilitiesWeights := make([]float64, len(possibilities))
		for i, possibility := range possibilities {
			possibilitiesWeights[i] = possibility.Technique.Difficulty()
		}
		step := possibilities[randomIndexWithInvertedWeights(possibilitiesWeights)]

		if branch == nil {
			results = append(results, step)
		} else {
			//We're in a branch point; we don't know if it's the RIGHT branch
			//So keep the steps somewhere else so we can throw them out if we unwind.
			branch.branchSteps = append(branch.branchSteps, step)
		}
		step.Apply(self)

	}

	if !self.Solved() {
		//We couldn't solve the puzzle.
		return nil
	}

	if branch != nil {
		//Apparently we're in the branch where the solution acutally lay. commit those steps and return them.

		//Walk up to the first branch.
		currentBranch := branch
		for {
			if currentBranch.previousBranchPoint == nil {
				break
			}
			currentBranch = currentBranch.previousBranchPoint
		}

		//now currentBranch is the earliest branch point
		//Walk down the list and copy in all of those steps
		for currentBranch != nil {
			results = append(results, currentBranch.branchSteps...)
			currentBranch = currentBranch.nextBranchPoint
		}
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
