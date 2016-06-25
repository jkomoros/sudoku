package sudoku

import (
	"container/heap"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
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
	CompoundSteps []*CompoundSolveStep
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

//CompoundSolveStep is a special type of step that has 0 to n precursor cull
//(non-fill) steps, followed by precisely one fill step. It reflects the
//notion that logically only fill steps actually advance the grid towards
//being solved, and all cull steps are in service of getting the grid to a
//state where a Fill step can be found.
type CompoundSolveStep struct {
	PrecursorSteps []*SolveStep
	FillStep       *SolveStep
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

//Steps returns the list of all CompoundSolveSteps flattened into one stream
//of SolveSteps.
func (s SolveDirections) Steps() []*SolveStep {
	var result []*SolveStep
	for _, compound := range s.CompoundSteps {
		result = append(result, compound.Steps()...)
	}
	return result
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

//effectiveTechniquesToUse returns the effective list of techniques to use.
//Basically just o.TechniquesToUse + Guess if NoGuess is not provided.
func (o *HumanSolveOptions) effectiveTechniquesToUse() []SolveTechnique {
	//TODO: now that we don't treat guess that specially in solving, shouldn't
	//we just get rid of all of the special casing in options?
	if o.NoGuess {
		return o.TechniquesToUse
	}
	return append(o.TechniquesToUse, GuessTechnique)
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

//Description returns a human-readable sentence describing what the SolveStep
//instructs the user to do, and what reasoning it used to decide that this
//step was logically valid to apply.
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

//newCompoundSolveStep will create a CompoundSolveStep from a series of
//SolveSteps, along as that series is a valid CompoundSolveStep.
func newCompoundSolveStep(steps []*SolveStep) *CompoundSolveStep {
	var result *CompoundSolveStep

	if len(steps) < 1 {
		return nil
	} else if len(steps) == 1 {
		result = &CompoundSolveStep{
			FillStep: steps[0],
		}
	} else {
		result = &CompoundSolveStep{
			PrecursorSteps: steps[0 : len(steps)-2],
			FillStep:       steps[len(steps)-1],
		}
	}
	if result.valid() {
		return result
	}
	return nil
}

//valid returns true iff there are 0 or more cull-steps in PrecursorSteps and
//a non-nill Fill step.
func (c *CompoundSolveStep) valid() bool {
	if c.FillStep == nil {
		return false
	}
	if !c.FillStep.Technique.IsFill() {
		return false
	}
	for _, step := range c.PrecursorSteps {
		if step.Technique.IsFill() {
			return false
		}
	}
	return true
}

//Apply applies all of the steps in the compound step to the grid in order:
//first each of the PrecursorSteps in order, then the fill step.
func (c *CompoundSolveStep) Apply(grid *Grid) {
	//TODO: test this
	if !c.valid() {
		return
	}
	for _, step := range c.PrecursorSteps {
		step.Apply(grid)
	}
	c.FillStep.Apply(grid)
}

//Description returns a human-readable sentence describing what the SolveStep
//instructs the user to do, and what reasoning it used to decide that this
//step was logically valid to apply.
func (c *CompoundSolveStep) Description() string {
	//TODO: this terminology is too tuned for the Online Sudoku use case.
	//it practice it should probably name the cell in text.
	var result []string
	result = append(result, "Based on the other numbers you've entered, "+c.FillStep.TargetCells[0].ref().String()+" can only be a "+strconv.Itoa(c.FillStep.TargetNums[0])+".")
	result = append(result, "How do we know that?")
	if len(c.PrecursorSteps) > 0 {
		result = append(result, "We can't fill any cells right away so first we need to cull some possibilities.")
	}
	steps := c.Steps()
	for i, step := range steps {
		intro := ""
		description := step.Description()
		if len(steps) > 1 {
			description = strings.ToLower(description)
			switch i {
			case 0:
				intro = "First, "
			case len(steps) - 1:
				intro = "Finally, "
			default:
				//TODO: switch between "then" and "next" randomly.
				intro = "Next, "
			}
		}
		result = append(result, intro+description)
	}
	return strings.Join(result, " ")
}

//Steps returns the simple list of SolveSteps that this CompoundSolveStep represents.
func (c *CompoundSolveStep) Steps() []*SolveStep {
	return append(c.PrecursorSteps, c.FillStep)
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
 * TODO: write a package-level comment about how the new solver works
 *
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

	var steps []*CompoundSolveStep

	if endConditionSolved {
		steps = newHumanSolveSearcher(grid, options)
	} else {
		steps = []*CompoundSolveStep{newHumanSolveSearcherSingleStep(grid, options, nil)}
	}

	return &SolveDirections{snapshot, steps}
}

//potentialNextStep keeps track of the next step we may want to return for
//HumanSolve.
type potentialNextStep struct {
	//All potentialNextSteps, except the initial in a frontier, must have a parent.
	parent    *potentialNextStep
	step      *SolveStep
	twiddles  map[string]float64
	heapIndex int
	frontier  *nextStepFrontier
}

//Grid returns a grid with all of this item's steps applied
func (p *potentialNextStep) Grid() *Grid {
	//TODO: it's conceivable that it might be best to memoize this... It's
	//unlikely thoguh, since grid will only be accessed once and many items
	//will never have their grids accessed.
	if p.frontier.grid == nil {
		return nil
	}

	result := p.frontier.grid.Copy()

	for _, step := range p.Steps() {
		step.Apply(result)
	}

	return result
}

//Goodness is how good the next step chain is in total. A LOWER Goodness is better. There's not enough precision between 0.0 and
//1.0 if we try to cram all values in there and they get very small.
func (p *potentialNextStep) Goodness() float64 {
	if p.parent == nil {
		return 1.0
	}
	//TODO: as an optimization we could cache this; each step is immutable basically.
	ownMultiplicationFactor := 1.0
	for _, twiddle := range p.twiddles {
		ownMultiplicationFactor *= twiddle
	}
	return p.parent.Goodness() * ownMultiplicationFactor
}

func (p *potentialNextStep) Steps() []*SolveStep {
	//TODO: can memoize this since it will never change
	if p.parent == nil {
		return nil
	}
	return append(p.parent.Steps(), p.step)
}

func (p *potentialNextStep) AddStep(step *SolveStep) *potentialNextStep {
	//TODO: Run the rest of the twiddlers!
	//TODO: add the twiddler for PointingCells/TargetCells overlap.
	result := &potentialNextStep{
		parent: p,
		step:   step,
		twiddles: map[string]float64{
			"Human Likelihood for " + step.TechniqueVariant(): step.Technique.humanLikelihood(step),
		},
		heapIndex: -1,
		frontier:  p.frontier,
	}
	if result.IsComplete() {
		p.frontier.CompletedItems = append(p.frontier.CompletedItems, result)
	} else {
		heap.Push(p.frontier, result)
	}
	return result
}

//Twiddle modifies goodness by the given amount and keeps track of the reason
//for debugging purposes.
func (p *potentialNextStep) Twiddle(amount float64, description string) {
	p.twiddles[description] = amount
	heap.Fix(p.frontier, p.heapIndex)
}

func (p *potentialNextStep) String() string {
	return fmt.Sprintf("%v %f %d", p.Steps(), p.Goodness(), p.heapIndex)
}

func (p *potentialNextStep) IsComplete() bool {
	steps := p.Steps()
	if len(steps) == 0 {
		return false
	}
	return steps[len(steps)-1].Technique.IsFill()
}

//Explore is the workhorse of HumanSolve; it's the thing that identifies all
//of the new steps rooted from here in parallel (and bails early if we've
//found enough results)
func (p *potentialNextStep) Explore() {

	/*

		//TODO: update this documentation

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

	//TODO: make this configurable, and figure out what the optimal values are
	numTechniquesToStartByDefault := 10

	techniques := p.frontier.options.effectiveTechniquesToUse()

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

	gridToUse := p.Grid()

	var wg sync.WaitGroup

	//The next technique to spin up
	nextTechniqueIndex := 0

	//We'll be kicking off this routine from multiple places so just define it once
	startTechnique := func(theTechnique SolveTechnique) {
		theTechnique.Find(gridToUse, resultsChan, done)
		//This is where a new technique should be kicked off, if one's going to be, before we tell the waitgroup that we're done.
		//We need to communicate synchronously with that thread
		comms := make(chan bool)
		techniqueFinished <- comms
		//Wait to hear back that a new technique is started, if one is going to be.
		<-comms

		//Okay, now the other thread has either started a new technique going, or hasn't.
		wg.Done()
	}

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
		p.AddStep(result)
		//Do we have enough steps accumulate?
		if p.frontier.DoneSearching() {
			//Communicate to all still-running routines that they can stop
			close(done)
			break OuterLoop
		}
	}
}

//TODO: rename this whole thing because it now does much more than frontier.
//TODO: rename basically every single thing I'm adding in this branch once it
//becomes clear exactly how they will end up.
type nextStepFrontier struct {
	//TODO: rename this field itemsToExplore, or 'frontier'
	items          []*potentialNextStep
	grid           *Grid
	CompletedItems []*potentialNextStep
	options        *HumanSolveOptions
}

func newNextStepFrontier(grid *Grid, options *HumanSolveOptions) *nextStepFrontier {
	frontier := &nextStepFrontier{
		grid:    grid,
		options: options,
	}
	heap.Init(frontier)
	initialItem := &potentialNextStep{
		frontier:  frontier,
		heapIndex: -1,
	}
	heap.Push(frontier, initialItem)
	return frontier
}

//DoneSearching will return true when no more items need to be explored
//because we have enough CompletedItems.
func (n *nextStepFrontier) DoneSearching() bool {
	if n.options == nil {
		return true
	}
	//TODO: is this the proper use of NumOptionsToCalculate?
	return n.options.NumOptionsToCalculate <= len(n.CompletedItems)
}

func (n *nextStepFrontier) String() string {
	result := "Items:" + strconv.Itoa(len(n.items)) + "\n"
	result += "Completed:" + strconv.Itoa(len(n.CompletedItems)) + "\n"
	result += "[\n"
	for _, item := range n.items {
		result += item.String() + "\n"
	}
	result += "]\n"
	return result
}

func (n nextStepFrontier) Len() int {
	return len(n.items)
}

func (n nextStepFrontier) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return n.items[i].Goodness() > n.items[j].Goodness()
}

func (n nextStepFrontier) Swap(i, j int) {
	n.items[i], n.items[j] = n.items[j], n.items[i]
	n.items[i].heapIndex = i
	n.items[j].heapIndex = j
}

func (n *nextStepFrontier) Push(x interface{}) {
	length := len(n.items)
	item := x.(*potentialNextStep)
	item.heapIndex = length
	n.items = append(n.items, item)
}

func (n *nextStepFrontier) Pop() interface{} {
	old := n.items
	length := len(old)
	item := old[length-1]
	item.heapIndex = -1 // for safety
	n.items = old[0 : length-1]
	return item
}

//NextPossibleStep pops the best step and returns it.
func (n *nextStepFrontier) NextPossibleStep() *potentialNextStep {
	if n.Len() == 0 {
		return nil
	}
	return n.Pop().(*potentialNextStep)
}

//newHumanSolveSearcher is a new implementation of the core implementation of
//HumanSolve. Mutates the grid.
func newHumanSolveSearcher(grid *Grid, options *HumanSolveOptions) []*CompoundSolveStep {
	//TODO: drop the 'new' from the name.
	var result []*CompoundSolveStep

	for !grid.Solved() {
		newStep := newHumanSolveSearcherSingleStep(grid, options, result)
		if newStep == nil {
			//Sad, guess we failed to solve the puzzle. :-(
			return nil
		}
		result = append(result, newStep)
		newStep.Apply(grid)
	}

	return result
}

//newHumanSolveSearcherSingleStep is the workhorse of the new HumanSolve. It
//searches for the next FillStepChain on the puzzle: a series of steps that
//contains exactly one fill step at its end.
func newHumanSolveSearcherSingleStep(grid *Grid, options *HumanSolveOptions, previousSteps []*CompoundSolveStep) *CompoundSolveStep {

	//TODO: drop the 'new' from the name

	//TODO: with the new approach, we're getting a lot more extreme negative difficulty values. Train a new model!

	//TODO: consider making a special FillStepChain type to use for all of
	//this that asserts in the type system that the chain of solve steps has
	//precisely one fill step and it's at the end of the chain. Conceivably it
	//should have two fields: FillStep and CullSteps.

	frontier := newNextStepFrontier(grid, options)

	step := frontier.NextPossibleStep()

	for step != nil && !frontier.DoneSearching() {
		//Explore step, finding all possible steps that apply from here and
		//adding to the frontier.

		//When adding a step, frontier notes if it's completed (thus going in
		//CompletedItems) or not (thus going in the itemsToExplore)

		//Once frontier.CompletedItems is at least
		//options.NumOptionsToCalculate we can bail out of looking for more
		//steps, shut down other threads, and break out of this loop.

		step.Explore()

		//We do NOT add the explored item back into the frontier.

		step = frontier.NextPossibleStep()

	}

	//Go through possibleCompleteStepsPool and pick one, preferring the lowest valued ones.

	//But first check if we don't have any.
	if len(frontier.CompletedItems) == 0 {
		return nil
	}

	distribution := make(ProbabilityDistribution, len(frontier.CompletedItems))

	for i, item := range frontier.CompletedItems {
		distribution[i] = item.Goodness()
	}

	invertedDistribution := distribution.invert()

	randomIndex := invertedDistribution.RandomIndex()

	return newCompoundSolveStep(frontier.CompletedItems[randomIndex].Steps())

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

	//TODO: figure out how to expose this meaningfully with the new human solve techniques.

	//TODO: hoist this special guess logic out if we decide to commit this.

	/*


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

	*/

	return nil, nil
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
