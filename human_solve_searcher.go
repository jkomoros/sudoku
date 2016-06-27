package sudoku

import (
	"container/heap"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

/*
 * A CompoundSolveStep is a series of 0 or more PrecursorSteps that are cull
 * steps (as opposed to fill steps), terminated with a single, non-optional
 * fill step. This organization reflects the observation that cull steps are
 * only useful if they help advance the grid to a state where a FillStep is
 * obvious in the short-term.
 *
 * The process of finding a HumanSolution to a puzzle reduces down to an
 * iterative search for a series of CompoundSolveSteps that, when applied in
 * order, will cause the puzzle to be solved. Hint is effectively the same,
 * except only searching for one CompoundSolveStep.
 *
 * When trying to discover which CompoundSolveStep to return, we need to
 * generate a number of options and pick the best one. humanSolveSearcher is
 * the struct that contains the information about the search for the current
 * CompoundSolveStep. In keeps track of the valid CompoundSolveSteps that have
 * already been found, and the in-progress CompoundSolveSteps that are not yet
 * complete (that is, that have not yet found their terminating fill step).
 *
 * Each possible CompoundSolveStep that is being considered (both incomplete
 * and complete ones) is represented by a humanSolveItem. Each humanSolveItem
 * has a parent humanSolveItem, except for the special initial item. Most
 * properties about a humanSolveItem are fixed at creation time. Each
 * humanSolveItem has one SolveStep representing the last SolveStep in their
 * chain.
 *
 * Each humanSolveItem has a Goodness() score, reflecting how good of an option
 * it is. Lower scores are better. This score is a function of the Twiddles
 * applied to this item and the twiddles applied up through its ancestor chain.
 * Twiddles between 0.0 and 1.0 make a humanSolveItem more good; values between
 * 1.0 and Infinity make it less good. Normally, the longer the chain of Steps,
 * the higher (worse) the Goodness score.
 *
 * humanSolveSearcher maintains a list of completedItems that it has found--
 * that is, humanSolveItems whose chain of steps represents a valid
 * CompoundSolveStep (0 or more cull steps followed by a single fill step). As
 * soon as a humanSolveItem is found, if it is a valid CompoundSolveStep, it is
 * added to the completedItems list. As soon as the completedItems list is
 * greater than options.NumOptionsToCalculate, we cease looking for more items
 * and move onto step selection phase.
 *
 * humanSolveSearcher maintains a heap of humanSolveItems sorted by their
 * Goodness. It will explore each item in order. When it explores each item, it
 * derives a grid representing the current grid state mutated by all of the
 * steps so far in this humanSolveItem's ancestor chain. It then searches for
 * all SolveSteps that can be found at this grid state and creates
 * humanSolveItems for each one, with this humanSolveItem as their parent. As
 * these items are created they are either put in completedItems or
 * itemsToExplore, depending on if they are complete or not. Once a
 * humanSolveItem is explored it is not put back in the itemsToExplore heap.
 * Goodness inverted and picked.
 *
 * Once humanSolveSearcher has found options.NumOptionsToCalculate
 * completedItems, it goes into selection phase. It creates a
 * ProbabilityDistribution with each item's Goodness() score. Then it inverts
 * that distribution and uses it to pick which CompoundSolveStep to return.
 */

//humanSolveSearcher keeps track of the search for a single new
//CompoundSolveStep. It keeps track of the humanSolveItems that are in-
//progress (itemsToExplore) and the items that are fully complete (that is,
//that are terminated by a FillStep and valid to return as an option).
type humanSolveSearcher struct {
	itemsToExplore []*humanSolveItem
	completedItems []*humanSolveItem
	//Various options frozen in at creation time that various methods need
	//access to.
	grid                  *Grid
	options               *HumanSolveOptions
	previousCompoundSteps []*CompoundSolveStep
}

//humanSolveItem keeps track of in-progress CompoundSolveSteps that we're
//currently building and considering. It also maintains various metadata about
//how this item fits in the searcher. Many things about the item are frozen at
//the time of creation; many of the properties of the humanSolveItem are
//derived recursively from the parents.
type humanSolveItem struct {
	//All humanSolveItem, except the initial in a searcher, must have a parent.
	parent    *humanSolveItem
	step      *SolveStep
	twiddles  map[string]probabilityTweak
	heapIndex int
	searcher  *humanSolveSearcher
}

//humanSolveHelper does most of the basic set up for both HumanSolve and Hint.
func humanSolveHelper(grid *Grid, options *HumanSolveOptions, endConditionSolved bool) *SolveDirections {
	//Short circuit solving if it has multiple solutions.
	if grid.HasMultipleSolutions() {
		return nil
	}

	if options == nil {
		options = DefaultHumanSolveOptions()
	}

	//TODO: we could also do a test for if it's already solved here.
	//(newHumanSolveSearcher implicitly does in the loop, but no harm in
	//checking here once too.

	options.validate()

	snapshot := grid.Copy()

	var steps []*CompoundSolveStep

	if endConditionSolved {
		steps = humanSolveSearch(grid, options)
	} else {
		result := humanSolveSearchSingleStep(grid, options, nil)
		if result != nil {
			steps = []*CompoundSolveStep{result}
		}
	}

	if len(steps) == 0 {
		return nil
	}

	return &SolveDirections{snapshot, steps}
}

//humanSolveSearch is a new implementation of the core implementation of
//HumanSolve. Mutates the grid.
func humanSolveSearch(grid *Grid, options *HumanSolveOptions) []*CompoundSolveStep {
	var result []*CompoundSolveStep

	for !grid.Solved() {
		newStep := humanSolveSearchSingleStep(grid, options, result)
		if newStep == nil {
			//Sad, guess we failed to solve the puzzle. :-(
			return nil
		}
		result = append(result, newStep)
		newStep.Apply(grid)
	}

	return result
}

//humanSolveSearchSingleStep is the workhorse of the new HumanSolve. It
//searches for the next CompoundSolveStep on the puzzle: a series of steps that
//contains exactly one fill step at its end.
func humanSolveSearchSingleStep(grid *Grid, options *HumanSolveOptions, previousSteps []*CompoundSolveStep) *CompoundSolveStep {

	//This function doesn't do much on top of HumanSolvePossibleSteps, but
	//it's worth it to mirror humanSolveSearch

	steps, distribution := grid.HumanSolvePossibleSteps(options, previousSteps)

	if len(steps) == 0 || len(distribution) == 0 {
		return nil
	}

	randomIndex := distribution.RandomIndex()

	return steps[randomIndex]
}

/************************************************************
 *
 * humanSolveItem implementation
 *
 ************************************************************/

//Grid returns a grid with all of this item's steps applied
func (p *humanSolveItem) Grid() *Grid {
	//TODO: it's conceivable that it might be best to memoize this... It's
	//unlikely thoguh, since grid will only be accessed once and many items
	//will never have their grids accessed.
	if p.searcher.grid == nil {
		return nil
	}

	result := p.searcher.grid.Copy()

	for _, step := range p.Steps() {
		step.Apply(result)
	}

	return result
}

//Goodness is how good the next step chain is in total. A LOWER Goodness is better. There's not enough precision between 0.0 and
//1.0 if we try to cram all values in there and they get very small.
func (p *humanSolveItem) Goodness() float64 {
	if p.parent == nil {
		return 1.0
	}
	//TODO: as an optimization we could cache this; each step is immutable basically.
	ownMultiplicationFactor := probabilityTweak(1.0)
	for _, twiddle := range p.twiddles {
		ownMultiplicationFactor *= twiddle
	}
	return p.parent.Goodness() * float64(ownMultiplicationFactor)
}

//explainGoodness returns a string explaining why this item has the goodness
//it does. Primarily useful for debugging.
func (p *humanSolveItem) explainGoodness(startCount int) string {
	if p.parent == nil {
		return ""
	}
	var resultSections []string
	for name, value := range p.twiddles {
		resultSections = append(resultSections, strconv.Itoa(startCount)+":"+name+":"+strconv.FormatFloat(float64(value), 'f', -1, 64))
	}

	return p.parent.explainGoodness(startCount+1) + "\n" + strings.Join(resultSections, "\n")

}

func (p *humanSolveItem) Steps() []*SolveStep {
	//TODO: can memoize this since it will never change
	if p.parent == nil {
		return nil
	}
	return append(p.parent.Steps(), p.step)
}

func (p *humanSolveItem) AddStep(step *SolveStep) *humanSolveItem {
	result := &humanSolveItem{
		parent:    p,
		step:      step,
		twiddles:  map[string]probabilityTweak{},
		heapIndex: -1,
		searcher:  p.searcher,
	}
	inProgressCompoundStep := p.Steps()
	grid := result.Grid()
	for name, twiddler := range twiddlers {
		tweak := twiddler(step, inProgressCompoundStep, p.searcher.previousCompoundSteps, grid)
		result.Twiddle(tweak, name)
	}
	if result.IsComplete() {
		p.searcher.completedItems = append(p.searcher.completedItems, result)
	} else {
		heap.Push(p.searcher, result)
	}
	return result
}

//Twiddle modifies goodness by the given amount and keeps track of the reason
//for debugging purposes. A twiddle of 1.0 has no effect.q A twiddle between
//0.0 and 1.0 increases the goodness. A twiddle of 1.0 or greater decreases
//goodness.
func (p *humanSolveItem) Twiddle(amount probabilityTweak, description string) {
	if amount < 0.0 {
		return
	}
	p.twiddles[description] = amount
	if p.heapIndex >= 0 {
		heap.Fix(p.searcher, p.heapIndex)
	}
}

func (p *humanSolveItem) String() string {
	return fmt.Sprintf("%v %f %d", p.Steps(), p.Goodness(), p.heapIndex)
}

func (p *humanSolveItem) IsComplete() bool {
	steps := p.Steps()
	if len(steps) == 0 {
		return false
	}
	return steps[len(steps)-1].Technique.IsFill()
}

//Explore is the workhorse of HumanSolve; it's the thing that identifies all
//of the new steps rooted from here in parallel (and bails early if we've
//found enough results)
func (p *humanSolveItem) Explore() {

	/*

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

	//TODO: play around with debug hints in i-sudoku ahile to develop an
	//intuition of what's happening in practice.

	//TODO: hidden/naked triple row is WAY too cheap as a cull step. It's the first
	//one that shows up often!

	//TODO: Obvious in collection is never showing up, even when it should be
	//far and away the number 1.

	//TODO: this in practice fills out Guesses ALL of the time, which causes
	//the probability distributions to go really wonky. Maybe only fall back
	//on Guess if no other things come out?
	techniques := p.searcher.options.effectiveTechniquesToUse()

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
		if p.searcher.DoneSearching() {
			//Communicate to all still-running routines that they can stop
			close(done)
			break OuterLoop
		}
	}
}

/************************************************************
 *
 * humanSolveSearcher implementation
 *
 ************************************************************/

func newHumanSolveSearcher(grid *Grid, previousCompoundSteps []*CompoundSolveStep, options *HumanSolveOptions) *humanSolveSearcher {
	searcher := &humanSolveSearcher{
		grid:                  grid,
		options:               options,
		previousCompoundSteps: previousCompoundSteps,
	}
	heap.Init(searcher)
	initialItem := &humanSolveItem{
		searcher:  searcher,
		heapIndex: -1,
	}
	heap.Push(searcher, initialItem)
	return searcher
}

//DoneSearching will return true when no more items need to be explored
//because we have enough CompletedItems.
func (n *humanSolveSearcher) DoneSearching() bool {
	if n.options == nil {
		return true
	}
	return n.options.NumOptionsToCalculate <= len(n.completedItems)
}

//NextPossibleStep pops the best step and returns it.
func (n *humanSolveSearcher) NextPossibleStep() *humanSolveItem {
	if n.Len() == 0 {
		return nil
	}
	return n.Pop().(*humanSolveItem)
}

//String prints out a useful debug output for the searcher's state.
func (n *humanSolveSearcher) String() string {
	result := "Items:" + strconv.Itoa(len(n.itemsToExplore)) + "\n"
	result += "Completed:" + strconv.Itoa(len(n.completedItems)) + "\n"
	result += "[\n"
	for _, item := range n.itemsToExplore {
		result += item.String() + "\n"
	}
	result += "]\n"
	return result
}

//Len is necessary to implement heap.Interface
func (n humanSolveSearcher) Len() int {
	return len(n.itemsToExplore)
}

//Less is necessary to implement heap.Interface
func (n humanSolveSearcher) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return n.itemsToExplore[i].Goodness() > n.itemsToExplore[j].Goodness()
}

//Swap is necessary to implement heap.Interface
func (n humanSolveSearcher) Swap(i, j int) {
	n.itemsToExplore[i], n.itemsToExplore[j] = n.itemsToExplore[j], n.itemsToExplore[i]
	n.itemsToExplore[i].heapIndex = i
	n.itemsToExplore[j].heapIndex = j
}

//Push is necessary to implement heap.Interface. It should not be used
//direclty; instead, use heap.Push()
func (n *humanSolveSearcher) Push(x interface{}) {
	length := len(n.itemsToExplore)
	item := x.(*humanSolveItem)
	item.heapIndex = length
	n.itemsToExplore = append(n.itemsToExplore, item)
}

//Pop is necessary to implement heap.Interface. It should not be used
//directly; instead use heap.Pop()
func (n *humanSolveSearcher) Pop() interface{} {
	old := n.itemsToExplore
	length := len(old)
	item := old[length-1]
	item.heapIndex = -1 // for safety
	n.itemsToExplore = old[0 : length-1]
	return item
}

//HumanSolvePossibleSteps returns a list of CompoundSolveSteps that could
//apply at this state, along with the probability distribution that a human
//would pick each one. The optional previousSteps argument is the list of
//CompoundSolveSteps that have been applied to the grid so far, and is used
//primarily to tweak the probability distribution and make, for example, it
//more likely to pick cells in the same block as the cell that was just
//filled. This method is the workhorse at the core of HumanSolve() and is
//exposed here primarily so users of this library can get a peek at which
//possibilites exist at each step. cmd/i-sudoku is one user of this method.
func (self *Grid) HumanSolvePossibleSteps(options *HumanSolveOptions, previousSteps []*CompoundSolveStep) (steps []*CompoundSolveStep, distribution ProbabilityDistribution) {

	//TODO: with the new approach, we're getting a lot more extreme negative difficulty values. Train a new model!

	searcher := newHumanSolveSearcher(self, previousSteps, options)

	step := searcher.NextPossibleStep()

	for step != nil && !searcher.DoneSearching() {
		//Explore step, finding all possible steps that apply from here and
		//adding to the frontier of itemsToExplore.

		//When adding a step, searcher notes if it's completed (thus going in
		//CompletedItems) or not (thus going in the itemsToExplore)

		//Once searcher.CompletedItems is at least
		//options.NumOptionsToCalculate we can bail out of looking for more
		//steps, shut down other threads, and break out of this loop.

		step.Explore()

		//We do NOT add the explored item back into the frontier.

		step = searcher.NextPossibleStep()

	}

	//Prepare the distribution and list of steps

	//But first check if we don't have any.
	if len(searcher.completedItems) == 0 {
		return nil, nil
	}

	distri := make(ProbabilityDistribution, len(searcher.completedItems))
	var resultSteps []*CompoundSolveStep

	for i, item := range searcher.completedItems {
		distri[i] = item.Goodness()
		resultSteps = append(resultSteps, newCompoundSolveStep(item.Steps()))
	}

	invertedDistribution := distri.invert()

	return resultSteps, invertedDistribution
}
