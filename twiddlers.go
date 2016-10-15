package sudoku

import (
	"math"
	"sort"
)

//1.0 is a no op. 0.0 to 1.0 is increase goodness; 1.0 and above is decraase good
type probabilityTweak float64

//probabilityTwiddler is the interface that twiddlers should adhere to. They
//should return a value between 0.0 (no change; good) and 1.0 (maximal change;
//bad). If a twiddler's weight is a negative number (e.g. -1.0) it means do
//not modify the twiddler.Note that previousGrid is the grid state BEFORE the
//proposedStep is applied.
type probabilityTwiddler func(proposedStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, previousGrid Grid) probabilityTweak

type probabilityTwiddlerItem struct {
	f      probabilityTwiddler
	name   string
	weight probabilityTweak
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
			f:      twiddleHumanLikelihood,
			name:   "Human Likelihood",
			weight: -1.0,
		},
		{
			f:      twiddleChainedSteps,
			name:   "Chained Steps",
			weight: 50.0,
		},
		{
			f:      twiddleCommonNumbers,
			name:   "Common Numbers",
			weight: 4.0,
		},
		{
			f:      twiddlePointingTargetOverlap,
			name:   "Pointing Target Overlap",
			weight: 1.0,
		},
		{
			f:      twiddlePreferFilledGroups,
			name:   "Prefer Filled Groups",
			weight: 10.0,
		},
	}
}

func (p *probabilityTwiddlerItem) Twiddle(proposedStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, previousGrid Grid) probabilityTweak {

	result := p.f(proposedStep, inProgressCompoundStep, pastSteps, previousGrid)

	if p.weight >= 0.0 {
		//result is a number between 0.0 and 1.0
		result = result * p.weight
	}

	return result

}

//twiddlePointingTargetOverlap twiddles based on how much the targetcells
//overlap with the pointingcells of the proposed step. This tries to capture
//the fact that for cull steps in particular, we want to heavily incentivize
//steps that directly reduce possibilities in the next round of steps. This is
//conceptually similar to ChainSimilarity, but more targeted.
func twiddlePointingTargetOverlap(proposedStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, previousGrid Grid) probabilityTweak {
	if len(inProgressCompoundStep) == 0 {
		return 0.0
	}
	lastStep := inProgressCompoundStep[len(inProgressCompoundStep)-1]

	if proposedStep == nil {
		return 0.0
	}

	//We're going to look for two kinds of overlap: targetCell to targetCell,
	//and targetCell to PointerCell, because some techniques want one or the
	//other (do any want both?). We'll use the higher overlap.

	//Compute Target --> Pointer overlap

	currentStepPointerSet := proposedStep.PointerCells.toCellSet()
	lastStepTargetSet := lastStep.TargetCells.toCellSet()

	targetPointerUnion := currentStepPointerSet.union(lastStepTargetSet)
	targetPointerIntersection := currentStepPointerSet.intersection(lastStepTargetSet)

	targetPointerOverlap := float64(len(targetPointerIntersection)) / float64(len(targetPointerUnion))

	if math.IsNaN(targetPointerOverlap) {
		targetPointerOverlap = 0.0
	}

	//Compute Target --> Target overlap

	currentStepTargetSet := proposedStep.TargetCells.toCellSet()

	targetTargetUnion := currentStepTargetSet.union(lastStepTargetSet)
	targetTargetIntersection := currentStepTargetSet.intersection(lastStepTargetSet)

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

	//Squaring the flipped overlap will accelerate small ones. Adding to 1.0
	//makes sure that all non-zero ones push things UP.
	return probabilityTweak(flippedOverlap * flippedOverlap)

}

//twiddleTechniqueWeight is a fundamental twiddler based on the
//HumanLikeliehood of the current technique. In fact, it's so fundamental that
//it's arguably not even a twiddler at all.
func twiddleHumanLikelihood(proposedStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, previousGrid Grid) probabilityTweak {
	if proposedStep == nil {
		return 1.0
	}
	return probabilityTweak(proposedStep.HumanLikelihood())
}

//twiddleCommonNumbers will twiddle up steps whose TargetNumbers are over-
//represented in the grid (but not DIM). This captures that humans, in
//practice, will often choose to look for cells to fill for a number that is
//represented more in the grid, since they're more likely to be constrained by
//neighorbors with the same number.
func twiddleCommonNumbers(proposedStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, previousGrid Grid) probabilityTweak {

	//Skip steps that aren't fill or fill multiple
	if !proposedStep.Technique.IsFill() || len(proposedStep.TargetNums) > 1 {
		//Wait, isn't this privileging steps that aren't filled unnecessarily?
		//No, basically there are some twiddlers that only apply to FillSteps
		//and are worth nothing after that.
		return 0.0
	}

	keyNum := proposedStep.TargetNums[0]

	count := 0

	for _, cell := range previousGrid.Cells() {
		if cell.Number() == keyNum {
			count++
		}
	}

	if count == DIM {
		//This shouldn't happen; this only occurs when something is fully
		//filled.
		return probabilityTweak(0.0)
	}

	//More filled is good, so flip this count to get # of unfilled for that
	//number in the grid.
	count = DIM - count

	//CommonNumbers is way too strong. Do a percentage of DIM, multiplied by
	//say 2, plus 1.

	percentage := float64(count) / float64(DIM)

	return probabilityTweak(percentage)

}

//This function will tweak weights quite a bit to make it more likely that we will pick a subsequent step that
// is 'related' to the cells modified in the last step. For example, if the
// last step had targetCells that shared a row, then a step with
//target cells in that same row will be more likely this step. This captures the fact that humans, in practice,
//will have 'chains' of steps that are all related.
func twiddleChainedSteps(proposedStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, previousGrid Grid) probabilityTweak {

	var lastModifiedCells CellRefSlice

	if len(inProgressCompoundStep) > 0 {
		lastModifiedCells = inProgressCompoundStep[len(inProgressCompoundStep)-1].TargetCells
	} else if len(pastSteps) > 0 {
		lastCompoundStep := pastSteps[len(pastSteps)-1]
		if lastCompoundStep.FillStep != nil {
			lastModifiedCells = lastCompoundStep.FillStep.TargetCells
		}
	}

	if lastModifiedCells == nil {
		return 0.0
	}

	//Tweak every weight by how related they are.
	//Remember: these are INVERTED weights, so tweaking them down is BETTER.

	//Logically we should be attenuating Dissimilarity here, but for some reason the math.Pow(dissimilairty, 10) doesn't actually
	//appear to work here, which is maddening.

	similarity := proposedStep.TargetCells.chainSimilarity(lastModifiedCells)

	//We want it to be dissimilar is larger; flip it.
	dissimilarity := 1.0 - similarity

	//Get the exponetntial shape
	dissimilarity = math.Pow(10, dissimilarity)

	//Convert to a percentage between 0 and 1
	dissimilarity /= 10.0

	return probabilityTweak(dissimilarity)

}

//twiddlePreferFilledGroups benefits steps that fill cells in groups that are
//more filled than others. This reflects that humans tend to focus on groups
//that are more constrained when picking a cell to focus on.
func twiddlePreferFilledGroups(proposedStep *SolveStep, inProgressCompoundStep []*SolveStep, pastSteps []*CompoundSolveStep, previousGrid Grid) probabilityTweak {

	//Sanity check
	if !proposedStep.Technique.IsFill() || len(proposedStep.TargetCells) > 1 {
		return 0.0
	}

	cell := proposedStep.TargetCells[0]

	//We want unfilled, because lower is better.

	blockUnfilledCount := len(previousGrid.Block(cell.Block()).FilterByUnfilled())
	rowUnfilledCount := len(previousGrid.Row(cell.Row).FilterByUnfilled())
	colUnfilledCount := len(previousGrid.Col(cell.Col).FilterByUnfilled())

	//Use DIM-1 because of course one item in that collection is unfilled--the
	//Target Cell for this step!
	blockUnfilledPercentage := float64(blockUnfilledCount) / float64(DIM-1)
	rowUnfilledPercentage := float64(rowUnfilledCount) / float64(DIM-1)
	colUnfilledPercentage := float64(colUnfilledCount) / float64(DIM-1)

	//Normalize the values for the fact that blocks are way easier for humans
	//to see than rows, which are slightly better than cols. (Remember, lower
	//is better)
	blockValue := 0.5 * blockUnfilledPercentage
	rowValue := 0.9 * rowUnfilledPercentage
	colValue := 1.0 * colUnfilledPercentage

	//Accumulate the whole value together, with the strongest percentage
	//accounting for most of the cumulative effect. This reflects the
	//perception that it's great to have high filled values for all, but even
	//just having one group is great.

	//Similar code is in CellSlice.chainSimilarity

	values := []float64{
		blockValue,
		rowValue,
		colValue,
	}

	sort.Float64s(values)

	weights := []int{
		16,
		6,
		1,
	}

	accum := 0.0
	divisor := 0.0

	for index, value := range values {
		for i := 0; i < weights[index]; i++ {
			accum += value
			divisor += 1.0
		}
	}

	//Divide by the number of samples we fed in to get the weighted average
	accum /= divisor

	return probabilityTweak(accum)

}
