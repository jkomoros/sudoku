package sudoku

import (
	"math"
)

//1.0 is a no op. 0.0 to 1.0 is increase goodness; 1.0 and above is decraase good
type probabilityTweak float64

type probabilityTwiddler func(*SolveStep, []*SolveStep, []*CompoundSolveStep, *Grid) probabilityTweak

//twiddlers is the list of all of the twiddlers we should apply to change the
//probability distribution of possibilities at each step. They capture biases
//that humans have about which cells to focus on (which is separate from
//Technique.humanLikelihood, since that is about how common a technique in
//general, not in a specific context.)
var twiddlers []probabilityTwiddler

func init() {
	twiddlers = []probabilityTwiddler{
		twiddleChainedSteps,
		twiddleCommonNumbers,
	}
}

//twiddleCommonNumbers will twiddle up steps whose TargetNumbers are over-
//represented in the grid (but not DIM). This captures that humans, in
//practice, will often choose to look for cells to fill for a number that is
//represented more in the grid, since they're more likely to be constrained by
//neighorbors with the same number.
func twiddleCommonNumbers(currentStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, grid *Grid) probabilityTweak {
	//Calculate which numbers are most represented in the grid.
	numCounts := make(map[int]int)

	for _, cell := range grid.Cells() {
		if cell.Number() == 0 {
			continue
		}
		numCounts[cell.Number()]++
	}

	//Remove counts for targetNums that are already filled in every block.
	//This shouldn't matter, since no steps should suggest filling it, but
	//just as a sanity check.
	for key, val := range numCounts {
		if val == DIM {
			delete(numCounts, key)
		}
	}

	//TODO: we can optimize this by moving the early bail up to the front, and
	//by only calculating count for the targetnum of us.

	//Skip steps that aren't fill or fill multiple
	if !currentStep.Technique.IsFill() || len(currentStep.TargetNums) > 1 {
		return 1.0
	}

	count := numCounts[currentStep.TargetNums[0]]
	if count == 0 {
		count = 1
	}

	//TODO: figure out if this should be on a curve.
	return probabilityTweak(count)

}

//This function will tweak weights quite a bit to make it more likely that we will pick a subsequent step that
// is 'related' to the cells modified in the last step. For example, if the
// last step had targetCells that shared a row, then a step with
//target cells in that same row will be more likely this step. This captures the fact that humans, in practice,
//will have 'chains' of steps that are all related.
func twiddleChainedSteps(currentStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, grid *Grid) probabilityTweak {

	if len(inProgressCompoundStep) == 0 {
		return 1.0
	}

	lastModifiedCells := inProgressCompoundStep[len(inProgressCompoundStep)-1].TargetCells

	if lastModifiedCells == nil {
		return 1.0
	}

	//Tweak every weight by how related they are.
	//Remember: these are INVERTED weights, so tweaking them down is BETTER.

	//TODO: consider attentuating the effect of this; chaining is nice but shouldn't totally change the calculation for hard techniques.
	//It turns out that we probably want to STRENGTHEN the effect.
	//Logically we should be attenuating Dissimilarity here, but for some reason the math.Pow(dissimilairty, 10) doesn't actually
	//appear to work here, which is maddening.

	similarity := currentStep.TargetCells.chainSimilarity(lastModifiedCells)
	//Make sure that similarity is higher than 1 so raising 2 to this power will make it go up.
	similarity *= 10

	return probabilityTweak(math.Pow(10, similarity))

}
