package sudoku

import (
	"math"
)

type probabilityTwiddler func([]*SolveStep, *Grid, CellSlice) probabilityDistributionTweak

//twiddlers is the list of all of the twiddlers we should apply to change the
//probability distribution of possibilities at each step. They capture biases
//that humans have about which cells to focus on (which is separate from
//Technique.humanLikelihood, since that is about how common a technique in
//general, not in a specific context.)
var twiddlers []probabilityTwiddler

func init() {
	twiddlers = []probabilityTwiddler{
		twiddleChainedSteps,
	}
}

//Returns a no-op (all 1.0) probabilityDistributionTweak of the given length.
func defaultProbabilityDistributionTweak(length int) probabilityDistributionTweak {
	result := make(probabilityDistributionTweak, length)

	//Initialize result to a no op tweaks
	for i := 0; i < length; i++ {
		result[i] = 1.0
	}

	return result
}

//twiddleCommonNumbers will twiddle up steps whose TargetNumbers are over-
//represented in the grid (but not DIM). This captures that humans, in
//practice, will often choose to look for cells to fill for a number that is
//represented more in the grid, since they're more likely to be constrained by
//neighorbors with the same number.
func twiddleCommonNumbers(possibilities []*SolveStep, grid *Grid, lastModififedCells CellSlice) (tweaks probabilityDistributionTweak) {
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

	result := defaultProbabilityDistributionTweak(len(possibilities))

	for i, possibility := range possibilities {
		//Skip steps that aren't fill or fill multiple
		if !possibility.Technique.IsFill() || len(possibility.TargetNums) > 1 {
			continue
		}

		count := numCounts[possibility.TargetNums[0]]
		if count == 0 {
			count = 1
		}
		//TODO: figure out the curve/amount to tweak by.
		result[i] *= float64(count)
	}

	return result

}

//This function will tweak weights quite a bit to make it more likely that we will pick a subsequent step that
// is 'related' to the cells modified in the last step. For example, if the
// last step had targetCells that shared a row, then a step with
//target cells in that same row will be more likely this step. This captures the fact that humans, in practice,
//will have 'chains' of steps that are all related.
func twiddleChainedSteps(possibilities []*SolveStep, grid *Grid, lastModififedCells CellSlice) (tweaks probabilityDistributionTweak) {

	result := defaultProbabilityDistributionTweak(len(possibilities))

	if lastModififedCells == nil || len(possibilities) == 0 {
		return result
	}

	for i := 0; i < len(possibilities); i++ {
		possibility := possibilities[i]
		//Tweak every weight by how related they are.
		//Remember: these are INVERTED weights, so tweaking them down is BETTER.

		//TODO: consider attentuating the effect of this; chaining is nice but shouldn't totally change the calculation for hard techniques.
		//It turns out that we probably want to STRENGTHEN the effect.
		//Logically we should be attenuating Dissimilarity here, but for some reason the math.Pow(dissimilairty, 10) doesn't actually
		//appear to work here, which is maddening.

		similarity := possibility.TargetCells.chainSimilarity(lastModififedCells)
		//Make sure that similarity is higher than 1 so raising 2 to this power will make it go up.
		similarity *= 10

		result[i] = math.Pow(10, similarity)
	}

	return result
}
