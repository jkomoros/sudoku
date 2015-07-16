package sudoku

import (
	"fmt"
	"log"
	"math"
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

//SolveDirections is a list of SolveSteps that, when applied in order to a given Grid, would
//cause it to be solved.
type SolveDirections []*SolveStep

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
	//costs. Also note that the results may be wrong if the difficulty model
	//in use was trained on a different NumOptionsToCalculate.
	NumOptionsToCalculate int
	//TODO: add a TwiddleChainDissimilarity bool.
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

//HumanWalkthrough returns a human-readable, verbose walkthrough of how a
//human would solve the provided puzzle, without mutating the grid. A
//covenience wrapper around grid.HumanSolution and
//SolveDirections.Walkthrough. If options is nil, will use reasonable
//defaults.z
func (self *Grid) HumanWalkthrough(options *HumanSolveOptions) string {
	steps := self.HumanSolution(options)
	return steps.Walkthrough(self)
}

//HumanSolution returns the SolveDirections that represent how a human would
//solve this puzzle. It does not mutate the grid. If options is nil, will use
//reasonable defaults.
func (self *Grid) HumanSolution(options *HumanSolveOptions) SolveDirections {
	clone := self.Copy()
	defer clone.Done()
	return clone.HumanSolve(options)
}

func defaultHumanSolveOptions() *HumanSolveOptions {
	//The methods may mutate the options object, so create a new one each
	//time.
	return &HumanSolveOptions{
		NumOptionsToCalculate: 15,
	}
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
//will use reasonable defaults.
func (self *Grid) HumanSolve(options *HumanSolveOptions) SolveDirections {

	//TODO: there are lots of options to HumanSolve, like how hard to search, whether to weight based on chaining, etc. Should there be a way to configure those options?

	//Short circuit solving if it has multiple solutions.
	if self.HasMultipleSolutions() {
		return nil
	}

	if options == nil {
		options = defaultHumanSolveOptions()
	}

	return humanSolveHelper(self, options, true)
}

//SolveDirections returns a chain of SolveDirections, containing exactly one
//IsFill step at the end, that is a reasonable next step to move the puzzle
//towards being completed. It is effectively a hint to the user about what
//Fill step to do next, and why it's logically implied; the truncated return
//value of HumanSolve. Returns nil if the puzzle has multiple solutions or is
//otherwise invalid. If options is nil, will use reasonable defaults.
func (self *Grid) Hint(options *HumanSolveOptions) SolveDirections {

	//TODO: return HintDirections instead of SolveDirections

	//TODO: test that non-fill steps before the last one are necessary to unlock
	//the fill step at the end (cull them if not), and test that.

	//Short circuit solving of it if it has multiple solutions.
	if self.HasMultipleSolutions() {
		return nil
	}

	if options == nil {
		options = defaultHumanSolveOptions()
	}

	return humanSolveHelper(self, options, false)

}

//Do we even need a helper here? Can't we just make HumanSolve actually humanSolveHelper?
//The core worker of human solve, it does all of the solving between branch points.
func humanSolveHelper(grid *Grid, options *HumanSolveOptions, endConditionSolved bool) []*SolveStep {

	var results []*SolveStep

	//Note: trying these all in parallel is much slower (~15x) than doing them in sequence.
	//The reason is that in sequence we bailed early as soon as we found one step; now we try them all.

	var lastStep *SolveStep

	//Is this the first time through the loop?
	firstRun := true

	for firstRun || (endConditionSolved && !grid.Solved()) || (!endConditionSolved && lastStep != nil && !lastStep.Technique.IsFill()) {

		firstRun = false

		if grid.Invalid() {
			//We must have been in a branch and found an invalidity.
			//Bail immediately.
			return nil
		}

		possibilities := runTechniques(Techniques, grid)

		//Now pick one to apply.
		if len(possibilities) == 0 {
			//Hmm, didn't find any possivbilities. We failed. :-(
			break
		}

		//TODO: consider if we should stop picking techniques based on their weight here.
		//Now that Find returns a slice instead of a single, we're already much more likely to select an "easy" technique. ... Right?

		possibilitiesWeights := make([]float64, len(possibilities))
		for i, possibility := range possibilities {
			possibilitiesWeights[i] = possibility.HumanLikelihood()
		}

		tweakChainedStepsWeights(lastStep, possibilities, possibilitiesWeights)

		step := possibilities[randomIndexWithInvertedWeights(possibilitiesWeights)]

		results = append(results, step)
		lastStep = step
		step.Apply(grid)

	}
	if (endConditionSolved && !grid.Solved()) || (!endConditionSolved && lastStep != nil && !lastStep.Technique.IsFill()) {
		//We couldn't solve the puzzle.
		//But let's do one last ditch effort and try guessing.
		guessSteps := humanSolveGuess(grid, options, endConditionSolved)
		if len(guessSteps) == 0 {
			//Okay, we just totally failed.
			return nil
		}
		return append(results, guessSteps...)
	}
	return results
}

//Called when we have run out of options at a given state and need to guess.
func humanSolveGuess(grid *Grid, options *HumanSolveOptions, endConditionSolved bool) []*SolveStep {

	//Yes, using DIM*DIM is a gross hack... I really should be calling Find inside a goroutine...
	results := make(chan *SolveStep, DIM*DIM)
	done := make(chan bool)

	//TODO: consider doing a normal solve forward from here to figure out what the right branch is and just do that.

	//Find is meant to be run in a goroutine; it won't complete until it's searched everything.
	GuessTechnique.Find(grid, results, done)
	close(done)

	var guess *SolveStep

	//TODO: test cases where we expectmultipel results...
	select {
	case guess = <-results:
	default:
		//Coludn't find a guess step, oddly enough.
		return nil
	}

	//We'll just take the first guess step and forget about the other ones.

	//The guess technique passes back the other nums as PointerNums, which is a hack.
	//Unpack them and then nil it out to prevent confusing other people in the future with them.
	otherNums := guess.PointerNums
	guess.PointerNums = nil

	var gridCopy *Grid

	for {
		gridCopy = grid.Copy()

		guess.Apply(gridCopy)

		//Even if endConditionSolved is true, this guess we will return will be an IsFill,
		//thus terminating the search. From here on out all we're doing is verifying that
		//we picked the right branch at the guess if endConditionSolved is not true.
		solveSteps := humanSolveHelper(gridCopy, options, true)

		if len(solveSteps) != 0 {
			//Success!
			//Make ourselves look like that grid (to pass back the state of what the solution was) and return.
			grid.replace(gridCopy)
			gridCopy.Done()
			if endConditionSolved {
				return append([]*SolveStep{guess}, solveSteps...)
			} else {
				//Since we're trying to find a hint that terminates in an IsFill step,
				//and this guess IS the IsFill step, and we've verified that this
				//guess we chose is correct, just return the guess step back up.
				return []*SolveStep{guess}
			}

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

//This function will tweak weights quite a bit to make it more likely that we will pick a subsequent step that
// is 'related' to the last step. For example, if the last step had targetCells that shared a row, then a step with
//target cells in that same row will be more likely this step. This captures the fact that humans, in practice,
//will have 'chains' of steps that are all related.
func tweakChainedStepsWeights(lastStep *SolveStep, possibilities []*SolveStep, weights []float64) {

	if len(possibilities) != len(weights) {
		log.Println("Mismatched lenghts of weights and possibilities: ", possibilities, weights)
		return
	}

	if lastStep == nil || len(possibilities) == 0 {
		return
	}

	for i := 0; i < len(possibilities); i++ {
		possibility := possibilities[i]
		//Tweak every weight by how related they are.
		//Remember: these are INVERTED weights, so tweaking them down is BETTER.

		//TODO: consider attentuating the effect of this; chaining is nice but shouldn't totally change the calculation for hard techniques.
		//It turns out that we probably want to STRENGTHEN the effect.
		//Logically we should be attenuating Dissimilarity here, but for some reason the math.Pow(dissimilairty, 10) doesn't actually
		//appear to work here, which is maddening.
		weights[i] *= possibility.TargetCells.chainDissimilarity(lastStep.TargetCells)
	}
}

func runTechniques(techniques []SolveTechnique, grid *Grid) []*SolveStep {

	/*
		This function went from being a mere convenience function to being a complex piece of multi-threaded code.

		The basic idea is to parellelize all of the technique's.Find work.

		Each technique is designed so it will bail early if we tell it (via closing the done channel) we've already
		got enough steps found.

		We only want to spin up numTechniquesToStartByDefault # of techniques at a time, because likely we'll find
		enough steps before getting to the harder (and generally more expensive to calculate) techniques if earlier
		ones fail.

		There is one thread for each currently running technique's Find. The main thread collects results and
		figures out when it has enough that all of the other threads can stop searching (or, when it hears that no
		more results will be coming in and it should just stop). There are two other threads. One waits until the
		waitgroup is all done and then signals that back to the main thread by closing resultsChan. The other thread
		is notified every time a technique thread is done, and decides whether or not it should start a new technique
		thread now. The interplay of those last two threads is very timing sensitive; if wg.Done were called before
		we'd started up the new technique, we could return from the whole endeavor before getting enough steps collected.

	*/

	//We make a copy of the grid to search on to avoid race conditions where
	// main thread has already returned up to humanSolveHelper, but not all of the techinques have gotten
	//the message and freak out a bit because the grid starts changing under them.
	gridCopy := grid.Copy()

	//TODO: make these configurable, and figure out what the optimal values are
	numRequestedSteps := 15
	numTechniquesToStartByDefault := 10

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
			collector <- gridToUse.HumanSolution(nil).Signals()
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
