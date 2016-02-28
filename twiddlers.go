package sudoku

import (
	"math"
)

type probabilityTwiddler func([]*SolveStep, CellSlice) probabilityDistributionTweak

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

//This function will tweak weights quite a bit to make it more likely that we will pick a subsequent step that
// is 'related' to the cells modified in the last step. For example, if the
// last step had targetCells that shared a row, then a step with
//target cells in that same row will be more likely this step. This captures the fact that humans, in practice,
//will have 'chains' of steps that are all related.
func twiddleChainedSteps(possibilities []*SolveStep, lastModififedCells CellSlice) (tweaks probabilityDistributionTweak) {

	result := make(probabilityDistributionTweak, len(possibilities))

	//Initialize result to a no op tweaks
	for i := 0; i < len(possibilities); i++ {
		result[i] = 1.0
	}

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
