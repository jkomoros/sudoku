package sudoku

import (
	"math"
)

//1.0 is a no op. 0.0 to 1.0 is increase goodness; 1.0 and above is decraase good
type probabilityTweak float64

type probabilityTwiddler func(*SolveStep, []*SolveStep, []*CompoundSolveStep, *Grid) probabilityTweak

type probabilityTwiddlerItem struct {
	f    probabilityTwiddler
	name string
	//TODO: add a strength field here. ... This is actually not obvious how to
	//do. If it was just a multiplication, then that would effectively be a
	//constant for any CompoundSolveStep with the same number of Steps(). But
	//if we switch twiddlers to be addition then adjusting down requires a
	//negative number, and there it would be possible to get very negative.
	//Alternatively, could have it be where all twiddlers have to be added
	//together, and just set the expectation that they must be between 0 and
	//positive infininty. Hmmm...
}

//twiddlers is the list of all of the twiddlers we should apply to change the
//probability distribution of possibilities at each step. They capture biases
//that humans have about which cells to focus on (which is separate from
//Technique.humanLikelihood, since that is about how common a technique in
//general, not in a specific context.)
var twiddlers []probabilityTwiddlerItem

func init() {
	//twiddlers is not a map because a) we need to attach more info anyway,
	//and b) we want a stable ordering.
	twiddlers = []probabilityTwiddlerItem{
		{
			f:    twiddleHumanLikelihood,
			name: "Human Likelihood",
		},
		{
			f:    twiddleChainedSteps,
			name: "Chained Steps",
		},
		{
			f:    twiddleCommonNumbers,
			name: "Common Numbers",
		},
		{
			f:    twiddlePointingTargetOverlap,
			name: "Pointing Target Overlap",
		},
	}
}

//twiddlePointingTargetOverlap twiddles based on how much the targetcells
//overlap with the pointingcells of the proposed step. This tries to capture
//the fact that for cull steps in particular, we want to heavily incentivize
//steps that directly reduce possibilities in the next round of steps. This is
//conceptually similar to ChainSimilarity, but more targeted.
func twiddlePointingTargetOverlap(currentStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, grid *Grid) probabilityTweak {
	if len(inProgressCompoundStep) == 0 {
		return 1.0
	}
	lastStep := inProgressCompoundStep[len(inProgressCompoundStep)-1]

	if currentStep == nil {
		return 1.0
	}

	//We're going to look for two kinds of overlap: targetCell to targetCell,
	//and targetCell to PointerCell, because some techniques want one or the
	//other (do any want both?). We'll use the higher overlap.

	//Compute Target --> Pointer overlap

	targetPointerUnion := currentStep.PointerCells.Union(lastStep.TargetCells)
	targetPointerIntersection := currentStep.PointerCells.Intersection(lastStep.TargetCells)

	targetPointerOverlap := float64(len(targetPointerIntersection)) / float64(len(targetPointerUnion))

	if math.IsNaN(targetPointerOverlap) {
		targetPointerOverlap = 0.0
	}

	//Compute Target --> Target overlap

	targetTargetUnion := currentStep.TargetCells.Union(lastStep.TargetCells)
	targetTargetIntersection := currentStep.TargetCells.Intersection(lastStep.TargetCells)

	targetTargetOverlap := float64(len(targetTargetIntersection)) / float64(len(targetTargetUnion))

	if math.IsNaN(targetTargetOverlap) {
		targetTargetOverlap = 0.0
	}

	//Pick the larger overlap to go with.

	overlap := targetPointerOverlap

	if targetTargetOverlap > targetPointerOverlap {
		overlap = targetTargetOverlap
		//TargetTargetOverlap is slightly better than targetPointer overlap.
		//This number will be flipped in the next step, so bigger is better.
		overlap *= 1.1
	}

	//The more overlap, the better, at an increasing rate. And the smaller the
	//output, the better the twiddle is.

	flippedOverlap := 1.0 - overlap

	//A value of 0--if there's perfect overlap--is nonsense. It's too strong!
	if flippedOverlap == 0.0 {
		flippedOverlap = 0.001
	}

	//Squaring the flipped overlap will accelerate small ones.
	return probabilityTweak(flippedOverlap * flippedOverlap)

}

//twiddleTechniqueWeight is a fundamental twiddler based on the
//HumanLikeliehood of the current technique. In fact, it's so fundamental that
//it's arguably not even a twiddler at all.
func twiddleHumanLikelihood(currentStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, grid *Grid) probabilityTweak {
	if currentStep == nil {
		return 1.0
	}
	return probabilityTweak(currentStep.HumanLikelihood())
}

//twiddleCommonNumbers will twiddle up steps whose TargetNumbers are over-
//represented in the grid (but not DIM). This captures that humans, in
//practice, will often choose to look for cells to fill for a number that is
//represented more in the grid, since they're more likely to be constrained by
//neighorbors with the same number.
func twiddleCommonNumbers(currentStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, grid *Grid) probabilityTweak {

	//Skip steps that aren't fill or fill multiple
	if !currentStep.Technique.IsFill() || len(currentStep.TargetNums) > 1 {
		return 1.0
	}

	keyNum := currentStep.TargetNums[0]

	count := 0

	for _, cell := range grid.Cells() {
		if cell.Number() == keyNum {
			count++
		}
	}

	if count == 0 || count == DIM {
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

	var lastModifiedCells CellSlice

	if len(inProgressCompoundStep) > 0 {
		lastModifiedCells = inProgressCompoundStep[len(inProgressCompoundStep)-1].TargetCells
	} else if len(pastSteps) > 0 {
		lastCompoundStep := pastSteps[len(pastSteps)-1]
		if lastCompoundStep.FillStep != nil {
			lastModifiedCells = lastCompoundStep.FillStep.TargetCells
		}
	}

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
