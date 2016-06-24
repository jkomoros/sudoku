package sudoku

import (
	"container/heap"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
)

//The number of solves we should average the signals together for before asking them for their difficulty
//Note: this should be set to the num-solves parameter used to train the currently configured weights.
const _NUM_SOLVES_FOR_DIFFICULTY = 10

//The list of techniques that HumanSolve will use to try to solve the puzzle, with the oddball Guess split out.
var (
	//All of the 'normal' Techniques that will be used to solve the puzzle
	Techniques []SolveTechnique
	//The special GuessTechnique that is used only if no other techniques find options.
	GuessTechnique SolveTechnique
	//Every technique that HumanSolve could ever use, including the oddball Guess technique.
	AllTechniques []SolveTechnique
	//Every variant name for every TechniqueVariant that HumanSolve could ever use.
	AllTechniqueVariants []string
)

//The actual techniques are intialized in hs_techniques.go, and actually defined in hst_*.go files.

//Worst case scenario, how many times we'd call HumanSolve to get a difficulty.
const _MAX_DIFFICULTY_ITERATIONS = 50

//TODO: consider relaxing this even more.
//How close we have to get to the average to feel comfortable our difficulty is converging.
const _DIFFICULTY_CONVERGENCE = 0.005

//SolveDirections is a list of SolveSteps that, when applied in order to its
//Grid, would cause it to be solved (except if IsHint is true).
type SolveDirections struct {
	//A copy of the Grid when the SolveDirections was generated. Grab a
	//reference from SolveDirections.Grid().
	gridSnapshot *Grid
	//The list of steps that, when applied in order, would cause the
	//SolveDirection's Grid() to be solved.
	Steps []*SolveStep
	//IsHint is whether the SolveDirections tells how to solve the given grid
	//or just what the next set of steps leading to a fill step is. If true,
	//the last step in Steps will be IsFill().
	IsHint bool
}

//SolveStep is a step to fill in a number in a cell or narrow down the possibilities in a cell to
//get it closer to being solved. SolveSteps model techniques that humans would use to solve a
//puzzle.
type SolveStep struct {
	//The technique that was used to identify that this step is logically valid at this point in the solution.
	Technique SolveTechnique
	//The cells that will be affected by the techinque (either the number to fill in or possibilities to exclude).
	TargetCells CellSlice
	//The numbers we will remove (or, in the case of Fill, add) to the TargetCells.
	TargetNums IntSlice
	//The cells that together lead the techinque to logically apply in this case; the cells behind the reasoning
	//why the TargetCells will be mutated in the way specified by this SolveStep.
	PointerCells CellSlice
	//The specific numbers in PointerCells that lead us to remove TargetNums from TargetCells.
	//This is only very rarely needed (at this time only for hiddenSubset techniques)
	PointerNums IntSlice
	//extra is a private place that information relevant to only specific techniques
	//can be stashed.
	extra interface{}
}

//TODO: consider passing a non-pointer humanSolveOptions so that mutations
//deeper  in the solve stack don' tmatter.

//HumanSolveOptions configures how precisely the human solver should operate.
//Passing nil where a HumanSolveOptions is expected will use reasonable
//defaults. Note that the various human solve methods may mutate your options
//object.
type HumanSolveOptions struct {
	//At each step in solving the puzzle, how many candidate SolveSteps should
	//we generate before stopping the search for more? Higher values will give
	//more 'realistic' solves, but at the cost of *much* higher performance
	//costs. Also note that the difficulty may be wrong if the difficulty
	//model in use was trained on a different NumOptionsToCalculate.
	NumOptionsToCalculate int
	//Which techniques to try at each step of the puzzle, sorted in the order
	//to try them out (generally from cheapest to most expensive). A value of
	//nil will use Techniques (the default). Any GuessTechniques will be
	//ignored.
	TechniquesToUse []SolveTechnique
	//NoGuess specifies that even if no other techniques work, the HumanSolve
	//should not fall back on guessing, and instead just return failure.
	NoGuess bool

	//TODO: figure out how to test that we do indeed use different values of
	//numOptionsToCalculate.
	//TODO: add a TwiddleChainDissimilarity bool.

	//The following are flags only used for testing.

	//When we reenter back into humanSolveHelper after making a guess, should
	//we keep the provided TechniquesToUse, or revert back to this set of
	//techniques? (If nil, don't change them) Mainly useful for the case where
	//we want to test that Hint works well when it returns a guess.
	techniquesToUseAfterGuess []SolveTechnique
}

//DefaultHumanSolveOptions returns a HumanSolveOptions object configured to
//have reasonable defaults.
func DefaultHumanSolveOptions() *HumanSolveOptions {
	result := &HumanSolveOptions{}

	result.NumOptionsToCalculate = 15
	result.TechniquesToUse = Techniques
	result.NoGuess = false

	//Have to set even zero valued properties, because the Options isn't
	//necessarily default initalized.
	result.techniquesToUseAfterGuess = nil

	return result

}

//Grid returns a snapshot of the grid at the time this SolveDirections was
//generated. Returns a fresh copy every time.
func (self SolveDirections) Grid() *Grid {
	//TODO: this is the only pointer receiver method on SolveDirections.
	return self.gridSnapshot.Copy()
}

//Modifies the options object to make sure all of the options are set
//in a legal way. Returns itself for convenience.
func (self *HumanSolveOptions) validate() *HumanSolveOptions {

	if self.TechniquesToUse == nil {
		self.TechniquesToUse = Techniques
	}

	if self.NumOptionsToCalculate < 1 {
		self.NumOptionsToCalculate = 1
	}

	//Remove any GuessTechniques that might be in there because
	//the are invalid.
	var techniques []SolveTechnique

	for _, technique := range self.TechniquesToUse {
		if technique == GuessTechnique {
			continue
		}
		techniques = append(techniques, technique)
	}

	self.TechniquesToUse = techniques

	return self

}

//IsUseful returns true if this SolveStep, when applied to the given grid, would do useful work--that is, it would
//either fill a previously unfilled number, or cull previously un-culled possibilities. This is useful to ensure
//HumanSolve doesn't get in a loop of applying the same useless steps.
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

//Apply does the solve operation to the Grid that is defined by the configuration of the SolveStep, mutating the
//grid and bringing it one step closer to being solved.
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
				gridCell.SetExcluded(exclude, true)
			}
		}
	}
}

//Description returns a human-readable sentence describing what the SolveStep instructs the user to do, and what reasoning
//it used to decide that this step was logically valid to apply.
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

//HumanLikelihood is how likely a user would be to pick this step when compared with other possible steps.
//Generally inversely related to difficulty (but not perfectly).
//This value will be used to pick which technique to apply when compared with other candidates.
//Based on the technique's HumanLikelihood, possibly attenuated by this particular step's variant
//or specifics.
func (self *SolveStep) HumanLikelihood() float64 {
	//TODO: attenuate by variant
	return self.Technique.humanLikelihood(self)
}

//TechniqueVariant returns the name of the precise variant of the Technique
//that this step represents. This information is useful for figuring out
//which weight to apply when calculating overall difficulty. A Technique would have
//variants (as opposed to simply other Techniques) when the work to calculate all
//variants is the same, but the difficulty of produced steps may vary due to some
//property of the technique. Forcing Chains is the canonical example.
func (self *SolveStep) TechniqueVariant() string {
	//Defer to the Technique.variant implementation entirely.
	//This allows us to most easily share code for the simple case.
	return self.Technique.variant(self)
}

//normalize puts the step in a known, deterministic state, which eases testing.
func (self *SolveStep) normalize() {
	//Different techniques will want to normalize steps in different ways.
	self.Technique.normalizeStep(self)
}

//HumanSolution returns the SolveDirections that represent how a human would
//solve this puzzle. It does not mutate the grid. If options is nil, will use
//reasonable defaults.
func (self *Grid) HumanSolution(options *HumanSolveOptions) *SolveDirections {
	clone := self.Copy()
	defer clone.Done()
	return clone.HumanSolve(options)
}

/*
 *
 * NOTE: THIS ENTIRE BLOCK NEEDS TO BE UPDATED NOW THAT WE JUST GUESS BY "CHEATING"
 * TODO: update this comment block
 *
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
 * For simplicity, we'll just show paths like this as a single line, even though realistically they could diverge arbitrarily,
 * before converging on the end state.
 *
 * This all changes when you introduce branching, because at a branch point you could have chosen the wrong path
 * and at some point down that path you will discover an invalidity, which tells you you chose wrong, and
 * you'll have to unwind.
 *
 * Let's explore a puzzle that needs one branch point.
 *
 * We explore with normal techniques until we run into a point where none of the normal techinques work.
 * This is a DIRE point, and in some cases we might just give up. But we have one last thing to try:
 * branching.
 * We then run the guess technique, which proposes multiple guess steps (big O's, in this diagram) that we could take.
 *
 * The technique will choose cells with only a small number of possibilities, to reduce the branching factor.
 *
 *                  |
 *                  |
 *                  V
 *                  O O O O O ...
 *
 * We will randomly pick one cell, and then explore all of its possibilities.
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

//HumanSolve is the workhorse of the package. It solves the puzzle much like a
//human would, applying complex logic techniques iteratively to find a
//sequence of steps that a reasonable human might apply to solve the puzzle.
//HumanSolve is an expensive operation because at each step it identifies all
//of the valid logic rules it could apply and then selects between them based
//on various weightings. HumanSolve endeavors to find the most realistic human
//solution it can by using a large number of possible techniques with
//realistic weights, as well as by doing things like being more likely to pick
//a cell that is in the same row/cell/block as the last filled cell. Returns
//nil if the puzzle does not have a single valid solution. If options is nil,
//will use reasonable defaults. Mutates the grid.
func (self *Grid) HumanSolve(options *HumanSolveOptions) *SolveDirections {
	return humanSolveHelper(self, options, true)
}

//SolveDirections returns a chain of SolveDirections, containing exactly one
//IsFill step at the end, that is a reasonable next step to move the puzzle
//towards being completed. It is effectively a hint to the user about what
//Fill step to do next, and why it's logically implied; the truncated return
//value of HumanSolve. Returns nil if the puzzle has multiple solutions or is
//otherwise invalid. If options is nil, will use reasonable defaults. Does not
//mutate the grid.
func (self *Grid) Hint(options *HumanSolveOptions) *SolveDirections {

	//TODO: return HintDirections instead of SolveDirections

	//TODO: test that non-fill steps before the last one are necessary to unlock
	//the fill step at the end (cull them if not), and test that.

	clone := self.Copy()
	defer clone.Done()

	result := humanSolveHelper(clone, options, false)
	result.IsHint = true

	return result

}

//humanSolveHelper does most of the set up for both HumanSolve and Hint.
func humanSolveHelper(grid *Grid, options *HumanSolveOptions, endConditionSolved bool) *SolveDirections {
	//Short circuit solving if it has multiple solutions.
	if grid.HasMultipleSolutions() {
		return nil
	}

	if options == nil {
		options = DefaultHumanSolveOptions()
	}

	options.validate()

	snapshot := grid.Copy()

	steps := humanSolveNonGuessSearcher(grid, options, endConditionSolved)

	return &SolveDirections{snapshot, steps, false}
}

/*

	Next steps planning


	potentialNextStep grows a Grid, which is a snapshot of the grid will all of the steps applied

	potentialNextStep grows a parent, a pointer to the potentialNextStep from which it was cloned

	Goodness becomes a derived field, which is the parent's goodness * all of
	the twiddlers at this level. Base case if no parent: 1.0.

	The only way to create a new potentialNextStep in a frontier is to find
	one and ForkWithStep. This creates a new potentialNextStep with its parent
	as the current step. It keeps a list of its twiddles that it applies to
	its parent, as well as a new Grid snapshot based on its parent. (Potential
	optimziation: grid is derived and can be thrown away)

	At each search step, we Pop the lowest item off the heap and explore it.
	Exploring searches for all techniques rooted here (stopping early if the
	pool is ever big enough, of course). That item that was popped is never
	added back into the frontier; it exists only in the parent chain.

	When you create a frontier, we add a special potentialNextStep that has no
	steps in it. This is the base case. We then explore from there and that's
	it.


*/

//potentialNextStep keeps track of the next step we may want to return for
//HumanSolve.
type potentialNextStep struct {
	Steps []*SolveStep
	//A LOWER Goodness is better. There's not enough precision between 0.0 and
	//1.0 if we try to cram all values in there and they get very small.
	Goodness  float64
	HeapIndex int
	Frontier  *nextStepFrontier
	//TODO: keep track of individual twiddles for long-term sake.
}

//AddStep adds a new step to the end of Steps and twiddles the goodness by the
//humanlikelihood of the new step.
func (p *potentialNextStep) AddStep(step *SolveStep) {
	p.Steps = append(p.Steps, step)
	//TODO: humanLikelihood is actually a bigger number the more unlikely.
	//Should small Goodness be better, or big goodness?
	p.Twiddle(step.Technique.humanLikelihood(step), "Human Likelihood for "+step.TechniqueVariant())
}

//Twiddle modifies goodness by the given amount and keeps track of the reason
//for debugging purposes.
func (p *potentialNextStep) Twiddle(amount float64, description string) {
	//TODO: store string somewhere.
	p.Goodness *= amount
	heap.Fix(p.Frontier, p.HeapIndex)
}

func (p *potentialNextStep) String() string {
	return fmt.Sprintf("%v %f %d", p.Steps, p.Goodness, p.HeapIndex)
}

func (p *potentialNextStep) IsComplete() bool {
	if len(p.Steps) == 0 {
		return false
	}
	return p.Steps[len(p.Steps)-1].Technique.IsFill()
}

type nextStepFrontier []*potentialNextStep

func newNextStepFrontier() *nextStepFrontier {
	frontier := &nextStepFrontier{}
	heap.Init(frontier)
	return frontier
}

func (n *nextStepFrontier) String() string {
	result := "[\n"
	for _, item := range *n {
		result += item.String() + "\n"
	}
	result += "]\n"
	return result
}

func (n nextStepFrontier) Len() int {
	return len(n)
}

func (n nextStepFrontier) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return n[i].Goodness > n[j].Goodness
}

func (n nextStepFrontier) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
	n[i].HeapIndex = i
	n[j].HeapIndex = j
}

func (n *nextStepFrontier) Push(x interface{}) {
	length := len(*n)
	item := x.(*potentialNextStep)
	item.HeapIndex = length
	*n = append(*n, item)
}

func (n *nextStepFrontier) Pop() interface{} {
	old := *n
	length := len(old)
	item := old[length-1]
	item.HeapIndex = -1 // for safety
	*n = old[0 : length-1]
	return item
}

func (n *nextStepFrontier) AddItem(steps []*SolveStep) *potentialNextStep {
	result := &potentialNextStep{
		Steps:     nil,
		Goodness:  1.0,
		HeapIndex: -1,
		Frontier:  n,
	}
	n.Push(result)
	//Add each step one at a time to get all of the twiddles.
	for _, step := range steps {
		result.AddStep(step)
	}

	return result
}

//newHumanSolveSearcher is a new implementation of the core implementation of
//HumanSolve.
func newHumanSolveSearcher(grid *Grid, options *HumanSolveOptions) []*SolveStep {
	//TODO: drop the 'new' from the name.
	var result []*SolveStep

	gridCopy := grid.Copy()

	for !gridCopy.Solved() {
		newStep := newHumanSolveSearcherSingleStep(gridCopy, options, result)
		result = append(result, newStep...)
	}

	//TODO: if we broke out and we didn't manage to solve the puzzle handle
	//that return value correctly

	return result
}

//newHumanSolveSearcherSingleStep is the workhorse of the new HumanSolve. It
//searches for the next FillStepChain on the puzzle: a series of steps that
//contains exactly one fill step at its end.
func newHumanSolveSearcherSingleStep(grid *Grid, options *HumanSolveOptions, previousSteps []*SolveStep) []*SolveStep {

	//TODO: drop the 'new' from the name

	//TODO: consider making a special FillStepChain type to use for all of
	//this that asserts in the type system that the chain of solve steps has
	//precisely one fill step and it's at the end of the chain.

	//TODO: implement this.
	return nil

}

//HumanSolvePossibleSteps returns a list of SolveSteps that could apply at
//this state, along with the probability distribution that a human would pick
//each one. The optional lastModifiedCells argument is the list of cells that
//were touched in the last action that was performed on the grid, and is used
//primarily to tweak the probability distribution and make, for example, it
//more likely to pick cells in the same block as the cell that was just
//filled. This method is the workhorse at the core of HumanSolve() and is
//exposed here primarily so users of this library can get a peek at which
//possibilites exist at each step. cmd/i-sudoku is one user of this method.
func (self *Grid) HumanSolvePossibleSteps(options *HumanSolveOptions, lastModifiedCells CellSlice) (steps []*SolveStep, distribution ProbabilityDistribution) {

	//TODO: hoist this special guess logic out if we decide to commit this.

	stepsToActuallyUse := options.TechniquesToUse

	if !options.NoGuess {
		stepsToActuallyUse = append(stepsToActuallyUse, GuessTechnique)
	}

	steps = runTechniques(stepsToActuallyUse, self, options.NumOptionsToCalculate)

	//Now pick one to apply.
	if len(steps) == 0 {
		return nil, nil
	}

	//TODO: consider if we should stop picking techniques based on their weight here.
	//Now that Find returns a slice instead of a single, we're already much more likely to select an "easy" technique. ... Right?

	invertedProbabilities := make(ProbabilityDistribution, len(steps))
	for i, possibility := range steps {
		invertedProbabilities[i] = possibility.HumanLikelihood()
	}

	d := invertedProbabilities.invert()

	for _, twiddler := range twiddlers {
		d = d.tweak(twiddler(steps, self, lastModifiedCells))
	}

	return steps, d
}

//Do we even need a helper here? Can't we just make HumanSolve actually humanSolveHelper?
//The core worker of human solve, it does all of the solving between branch points.
func humanSolveNonGuessSearcher(grid *Grid, options *HumanSolveOptions, endConditionSolved bool) []*SolveStep {

	var results []*SolveStep

	//Note: trying these all in parallel is much slower (~15x) than doing them in sequence.
	//The reason is that in sequence we bailed early as soon as we found one step; now we try them all.

	var lastStep *SolveStep

	//Is this the first time through the loop?
	firstRun := true

	for firstRun || (endConditionSolved && !grid.Solved()) || (!endConditionSolved && lastStep != nil && !lastStep.Technique.IsFill()) {

		firstRun = false

		var cells CellSlice

		if lastStep != nil {
			cells = lastStep.TargetCells
		}

		possibilities, probabilityDistribution := grid.HumanSolvePossibleSteps(options, cells)

		if len(possibilities) == 0 {
			//Hmm, we failed to find anything :-/
			break
		}

		step := possibilities[probabilityDistribution.RandomIndex()]

		results = append(results, step)
		lastStep = step
		step.Apply(grid)

	}
	return results
}

func runTechniques(techniques []SolveTechnique, grid *Grid, numRequestedSteps int) []*SolveStep {

	/*
		This function went from being a mere convenience function to
		being a complex piece of multi-threaded code.

		The basic idea is to parellelize all of the technique's.Find
		work.

		Each technique is designed so it will bail early if we tell it
		(via closing the done channel) we've already got enough steps
		found.

		We only want to spin up numTechniquesToStartByDefault # of
		techniques at a time, because likely we'll find enough steps
		before getting to the harder (and generally more expensive to
		calculate) techniques if earlier ones fail.

		There is one thread for each currently running technique's
		Find. The main thread collects results and figures out when it
		has enough that all of the other threads can stop searching
		(or, when it hears that no more results will be coming in and
		it should just stop). There are two other threads. One waits
		until the waitgroup is all done and then signals that back to
		the main thread by closing resultsChan. The other thread is
		notified every time a technique thread is done, and decides
		whether or not it should start a new technique thread now. The
		interplay of those last two threads is very timing sensitive;
		if wg.Done were called before we'd started up the new
		technique, we could return from the whole endeavor before
		getting enough steps collected.

	*/

	if numRequestedSteps < 1 {
		numRequestedSteps = 1
	}

	//We make a copy of the grid to search on to avoid race conditions where
	// main thread has already returned up to humanSolveHelper, but not all of the techinques have gotten
	//the message and freak out a bit because the grid starts changing under them.
	gridCopy := grid.Copy()

	//TODO: make this configurable, and figure out what the optimal values are
	numTechniquesToStartByDefault := 10

	//Handle the case where we were given a short list of techniques.
	if len(techniques) < numTechniquesToStartByDefault {
		numTechniquesToStartByDefault = len(techniques)
	}

	//Leave some room in resultsChan so all of the techniques don't have to block as often
	//waiting for the mainthread to clear resultsChan. Leads to a 20% reduction in time compared
	//to unbuffered.
	//We'll close this channel to signal the collector that no more results are coming.
	resultsChan := make(chan *SolveStep, len(techniques))
	done := make(chan bool)

	//Deliberately unbuffered; we want it to run sync inside of startTechnique
	//the thread that's waiting on it will pass its own chan that it should send to when it's done
	techniqueFinished := make(chan chan bool)

	var wg sync.WaitGroup

	//The next technique to spin up
	nextTechniqueIndex := 0

	//We'll be kicking off this routine from multiple places so just define it once
	startTechnique := func(theTechnique SolveTechnique) {
		theTechnique.Find(gridCopy, resultsChan, done)
		//This is where a new technique should be kicked off, if one's going to be, before we tell the waitgroup that we're done.
		//We need to communicate synchronously with that thread
		comms := make(chan bool)
		techniqueFinished <- comms
		//Wait to hear back that a new technique is started, if one is going to be.
		<-comms

		//Okay, now the other thread has either started a new technique going, or hasn't.
		wg.Done()
	}

	var results []*SolveStep

	//Get the first batch of techniques going
	wg.Add(numTechniquesToStartByDefault)

	//Since Techniques is in sorted order, we're starting off with the easiest techniques.
	for nextTechniqueIndex = 0; nextTechniqueIndex < numTechniquesToStartByDefault; nextTechniqueIndex++ {
		go startTechnique(techniques[nextTechniqueIndex])
	}

	//Listen for when all items are done and signal the collector to stop collecting
	go func() {
		wg.Wait()
		//All of the techniques must be done here; no one can send on resultsChan at this point.
		//Signal to the collector that it should break out.
		close(resultsChan)
		close(techniqueFinished)
	}()

	//The thread that will kick off new techinques
	go func() {
		for {
			returnChan, ok := <-techniqueFinished
			if !ok {
				//If channel is closed, that's our cue to die.
				return
			}
			//Start a technique here, if we're going to.
			//First, check if the collector has signaled that we're all done
			select {
			case <-done:
				//Don't start a new one
			default:
				//Potentially start a new technique going as things aren't shutting down yet.
				//Is there another technique?
				if nextTechniqueIndex < len(techniques) {
					wg.Add(1)
					go startTechnique(techniques[nextTechniqueIndex])
					//Next time we're considering starting a new technique, start the next one
					nextTechniqueIndex++
				}
			}

			//Tell our caller that we're done
			returnChan <- true
		}
	}()

	//Collect the results as long as more are coming
OuterLoop:
	for {
		result, ok := <-resultsChan
		if !ok {
			//resultsChan was closed, which is our signal that no more results are coming and we should break
			break OuterLoop
		}
		results = append(results, result)
		//Do we have enough steps accumulate?
		if len(results) > numRequestedSteps {
			//Communicate to all still-running routines that they can stop
			close(done)
			break OuterLoop
		}
	}

	return results
}

//Difficulty returns a value between 0.0 and 1.0, representing how hard the puzzle would be
//for a human to solve. :This is an EXTREMELY expensive method (although repeated calls without
//mutating the grid return a cached value quickly). It human solves the puzzle, extracts signals
//out of the solveDirections, and then passes those signals into a machine-learned model that
//was trained on hundreds of thousands of solves by real users in order to generate a candidate difficulty.
//It then repeats the process multiple times until the difficultly number begins to converge to
//an average.
func (self *Grid) Difficulty() float64 {

	//TODO: test that the memoization works (that is, the cached value is thrown out if the grid is modified)
	//It's hard to test because self.calculateDifficulty(true) is so expensive to run.

	//This is so expensive and during testing we don't care if converges.
	//So we split out the meat of the method separately.

	if self == nil {
		return 0.0
	}

	//Yes, this memoization will fail in the (rare!) cases where a grid's actual difficulty is 0.0, but
	//the worst case scenario is that we just return the same value.
	if self.cachedDifficulty == 0.0 {
		self.cachedDifficulty = self.calcluateDifficulty(true)
	}
	return self.cachedDifficulty
}

func (self *Grid) calcluateDifficulty(accurate bool) float64 {
	//This can be an extremely expensive method. Do not call repeatedly!
	//returns the difficulty of the grid, which is a number between 0.0 and 1.0.
	//This is a probabilistic measure; repeated calls may return different numbers, although generally we wait for the results to converge.

	//We solve the same puzzle N times, then ask each set of steps for their difficulty, and combine those to come up with the overall difficulty.

	accum := 0.0
	average := 0.0
	lastAverage := 0.0

	self.HasMultipleSolutions()

	//Since this is so expensive, in testing situations we want to run it in less accurate mode (so it goes fast!)
	maxIterations := _MAX_DIFFICULTY_ITERATIONS
	if !accurate {
		maxIterations = 1
	}

	for i := 0; i < maxIterations; i++ {
		difficulty := gridDifficultyHelper(self)

		accum += difficulty
		average = accum / (float64(i) + 1.0)

		if math.Abs(average-lastAverage) < _DIFFICULTY_CONVERGENCE {
			//Okay, we've already converged. Just return early!
			return average
		}

		lastAverage = average
	}

	//We weren't converging... oh well!
	return average
}

//This function will HumanSolve _NUM_SOLVES_FOR_DIFFICULTY times, then average the signals together, then
//give the difficulty for THAT. This is more accurate becuase the weights were trained on such averaged signals.
func gridDifficultyHelper(grid *Grid) float64 {

	collector := make(chan DifficultySignals, _NUM_SOLVES_FOR_DIFFICULTY)
	//Might as well run all of the human solutions in parallel
	for i := 0; i < _NUM_SOLVES_FOR_DIFFICULTY; i++ {
		go func(gridToUse *Grid) {
			solution := gridToUse.HumanSolution(nil)
			if solution == nil {
				log.Println("A generated grid turned out to have mutiple solutions (or otherwise return nil), indicating a very serious error:", gridToUse.DataString())
				os.Exit(1)
			}
			collector <- solution.Signals()
		}(grid)
	}

	combinedSignals := DifficultySignals{}

	for i := 0; i < _NUM_SOLVES_FOR_DIFFICULTY; i++ {
		signals := <-collector
		combinedSignals.sum(signals)
	}

	//Now average all of the signal values
	for key := range combinedSignals {
		combinedSignals[key] /= _NUM_SOLVES_FOR_DIFFICULTY
	}

	return combinedSignals.difficulty()

}
