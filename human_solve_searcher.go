package sudoku

import (
	"container/heap"
	"fmt"
	"runtime"
	"strconv"
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

//humanSolveSearcherHeap is what we will use for the heap implementation in
//searcher. We put it as a seaprate time to avoid having to have
//heap.Interface methods on searcher itself, since for proper use you're not
//supposed to call those directly. So putting them on a sub-struct helps hide
//them a bit.
type humanSolveSearcherHeap []*humanSolveItem

//humanSolveSearcher keeps track of the search for a single new
//CompoundSolveStep. It keeps track of the humanSolveItems that are in-
//progress (itemsToExplore) and the items that are fully complete (that is,
//that are terminated by a FillStep and valid to return as an option).
type humanSolveSearcher struct {
	itemsToExplore humanSolveSearcherHeap
	completedItems []*humanSolveItem
	//The number of straightforward completed items. That is,
	//CompoundSolveSteps with no precursorSteps that are not Guesses. We keep
	//track of this to figure out when we can early bail if we have enough of
	//them.
	straightforwardItemsCount int
	//TODO: keep track of stats: how big the frontier was at the end of each
	//CompoundSolveStep. Then provide max/mean/median.

	//Cache of steps that have been found. They will be added to the queue, so
	//after searching for this step we can add them.
	stepsCache *foundStepCache

	//done will be closed when DoneSearching will return true. A convenient
	//way for people to check DoneSearching without checking in a tight loop.
	done chan bool
	//... The hacky way to make sure we don't close an already-closed channel.
	channelClosed bool

	//itemsLock controls access to itemsToExplore, completedItems,
	//straightforwardItemsCount, stepsCache, etc.
	//TODO: consider having more fine-grained locks for performance.
	itemsLock sync.Mutex

	//Various options frozen in at creation time that various methods need
	//access to.
	grid                  Grid
	options               *HumanSolveOptions
	previousCompoundSteps []*CompoundSolveStep
}

//twiddleRecord is a key/value pair in twiddles. We want to preserve ordering,
//so we can't use a map.
type twiddleRecord struct {
	name  string
	value probabilityTweak
}

//humanSolveItem keeps track of in-progress CompoundSolveSteps that we're
//currently building and considering. It also maintains various metadata about
//how this item fits in the searcher. Many things about the item are frozen at
//the time of creation; many of the properties of the humanSolveItem are
//derived recursively from the parents.
type humanSolveItem struct {
	//All humanSolveItem, except the initial in a searcher, must have a parent.
	parent         *humanSolveItem
	step           *SolveStep
	twiddles       []twiddleRecord
	heapIndex      int
	searcher       *humanSolveSearcher
	cachedGrid     Grid
	cachedGoodness float64
	//The index of the next techinque to return
	techniqueIndex int
	added          bool
	doneTwiddling  bool
}

//humanSolveWorkItem represents a unit of work that should be done during the
//search.
type humanSolveWorkItem struct {
	grid        Grid
	technique   SolveTechnique
	coordinator findCoordinator
}

//channelFindCoordinator implements the findCoordinator interface. It's a
//simple wrapper around the basic channel logic currently used.
type channelFindCoordinator struct {
	results chan *SolveStep
	done    chan bool
}

//synchronousFindCoordinator implements the findCoordinator interface. It's
//basically just a thin wrapper around humanSolveSearcher. Desigend for use in
//NewSearch.
type synchronousFindCoordinator struct {
	searcher *humanSolveSearcher
	baseItem *humanSolveItem
}

//humanSolveHelper does most of the basic set up for both HumanSolve and Hint.
func humanSolveHelper(grid Grid, options *HumanSolveOptions, previousSteps []*CompoundSolveStep, endConditionSolved bool) *SolveDirections {
	//Short circuit solving if it has multiple solutions.
	if grid.HasMultipleSolutions() {
		return nil
	}

	if options == nil {
		options = DefaultHumanSolveOptions()
	}

	//To shave off a bit more performance, quickly check if the grid is
	//already solved.
	if grid.Solved() {
		return nil
	}

	options.validate()

	snapshot := grid.Copy()

	var steps []*CompoundSolveStep

	if endConditionSolved {
		steps = humanSolveSearch(grid, options)
	} else {
		result := humanSolveSearchSingleStep(grid, options, previousSteps, nil)
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
func humanSolveSearch(grid Grid, options *HumanSolveOptions) []*CompoundSolveStep {
	var result []*CompoundSolveStep

	isMutableGrid := false

	mGrid, ok := grid.(MutableGrid)

	if ok {
		isMutableGrid = true
	}

	stepsCache := &foundStepCache{}

	//TODO: it FEELS like here we should be using read only grids. Test what
	//happens if we get rid of the mutablegrid path (and modify the callers
	//who expect us to mutate the grid). however, we tried doing this, and it
	//added 90% to BenchmarkHumansolve. Presumably it's because we are
	//creating tons of extra grids when we can just accumulate the results in
	//the one item otherwise.
	for !grid.Solved() {
		newStep := humanSolveSearchSingleStep(grid, options, result, stepsCache)
		if newStep == nil {
			//Sad, guess we failed to solve the puzzle. :-(
			return nil
		}
		result = append(result, newStep)
		if isMutableGrid {
			newStep.Apply(mGrid)
		} else {
			grid = grid.CopyWithModifications(newStep.Modifications())
		}
		stepsCache.AddQueue()
		for _, step := range newStep.Steps() {
			stepsCache.RemoveStepsWithCells(step.TargetCells)
		}
	}

	return result
}

//humanSolveSearchSingleStep is the workhorse of the new HumanSolve. It
//searches for the next CompoundSolveStep on the puzzle: a series of steps that
//contains exactly one fill step at its end.
func humanSolveSearchSingleStep(grid Grid, options *HumanSolveOptions, previousSteps []*CompoundSolveStep, stepsCache *foundStepCache) *CompoundSolveStep {

	//This function doesn't do much on top of HumanSolvePossibleSteps, but
	//it's worth it to mirror humanSolveSearch

	steps, distribution := grid.humanSolvePossibleStepsWithCache(options, previousSteps, stepsCache)

	if len(steps) == 0 || len(distribution) == 0 {
		return nil
	}

	randomIndex := distribution.RandomIndex()

	return steps[randomIndex]
}

/************************************************************
 *
 * channelFindCoordinator implementation
 *
 ************************************************************/

func (c *channelFindCoordinator) shouldExitEarly() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *channelFindCoordinator) foundResult(step *SolveStep) bool {
	select {
	case c.results <- step:
		return false
	case <-c.done:
		return true
	}
}

/************************************************************
 *
 * synchronousFindCoordinator implementation
 *
 ************************************************************/

func (s *synchronousFindCoordinator) shouldExitEarly() bool {
	return s.searcher.DoneSearching()
}

func (s *synchronousFindCoordinator) foundResult(step *SolveStep) bool {
	s.baseItem.AddStep(step)
	return s.shouldExitEarly()
}

/************************************************************
 *
 * humanSolveItem implementation
 *
 ************************************************************/

//PreviousGrid returns a grid with all of the steps applied up to BUT NOT
//INCLUDING this items' step.
func (p *humanSolveItem) PreviousGrid() Grid {
	if p.parent == nil {
		return p.searcher.grid
	}
	return p.parent.Grid()
}

//Grid returns a grid with all of this item's steps applied
func (p *humanSolveItem) Grid() Grid {

	if p.cachedGrid == nil {

		var result Grid

		if p.searcher.grid == nil {
			result = nil
		} else if p.parent == nil {
			result = p.searcher.grid
		} else {
			result = p.parent.Grid().CopyWithModifications(p.step.Modifications())
		}

		p.cachedGrid = result
	}

	return p.cachedGrid

}

//Goodness is how good the next step chain is in total. A LOWER Goodness is better. There's not enough precision between 0.0 and
//1.0 if we try to cram all values in there and they get very small.
func (p *humanSolveItem) Goodness() float64 {
	if p.parent == nil {
		return 1.0
	}
	if p.doneTwiddling && p.cachedGoodness != 0 {
		return p.cachedGoodness
	}
	ownAdditionFactor := probabilityTweak(0.0)
	for _, twiddle := range p.twiddles {
		ownAdditionFactor += twiddle.value
	}
	//p.cachedGoodness will be overwritten in the future if doneTwiddling is
	//not yet true.
	p.cachedGoodness = p.parent.Goodness() + float64(ownAdditionFactor)
	return p.cachedGoodness
}

//explainGoodness returns a string explaining why this item has the goodness
//it does. Primarily useful for debugging.
func (p *humanSolveItem) explainGoodness() []string {
	result := []string{
		fmt.Sprintf("G:%f", p.Goodness()),
	}
	return append(result, p.explainGoodnessRecursive(0)...)
}

func (p *humanSolveItem) explainGoodnessRecursive(startCount int) []string {
	if p.parent == nil {
		return nil
	}
	var resultSections []string
	for _, twiddle := range p.twiddles {
		//0.0 values are boring, so skip them.
		if twiddle.value == 0.0 {
			continue
		}
		resultSections = append(resultSections, strconv.Itoa(startCount)+":"+twiddle.name+":"+strconv.FormatFloat(float64(twiddle.value), 'f', 4, 64))
	}
	parents := p.parent.explainGoodnessRecursive(startCount + 1)
	if parents == nil {
		return resultSections
	}
	if resultSections == nil {
		return parents
	}
	return append(parents, resultSections...)

}

func (p *humanSolveItem) Steps() []*SolveStep {
	//Memoizing this seems like it makes sense, but it actually leads to a ~1%
	//INCREASE in HumanSolve.
	if p.parent == nil {
		return nil
	}
	return append(p.parent.Steps(), p.step)
}

//createNewItem creates a new humanSolveItem based on this step, but NOT YET
//ADDED to searcher. call item.Add() to do that.
func (p *humanSolveItem) CreateNewItem(step *SolveStep) *humanSolveItem {
	result := &humanSolveItem{
		parent:    p,
		step:      step,
		twiddles:  nil,
		heapIndex: -1,
		searcher:  p.searcher,
	}
	inProgressCompoundStep := p.Steps()
	previousGrid := result.PreviousGrid()
	for _, twiddler := range twiddlers {
		tweak := twiddler.Twiddle(step, inProgressCompoundStep, p.searcher.previousCompoundSteps, previousGrid)
		result.Twiddle(tweak, twiddler.name)
	}
	result.DoneTwiddling()
	return result
}

//AddStep basically just does p.CreateNewItem, then item.Add()
func (p *humanSolveItem) AddStep(step *SolveStep) *humanSolveItem {
	result := p.CreateNewItem(step)
	p.searcher.AddItem(result)
	return result
}

//DoneTwiddling should be called once no more twiddles are expected. That's
//the signal that it's OK for us to cache twiddles.
func (p *humanSolveItem) DoneTwiddling() {
	p.doneTwiddling = true
}

//Twiddle modifies goodness by the given amount and keeps track of the reason
//for debugging purposes. A twiddle of 1.0 has no effect.q A twiddle between
//0.0 and 1.0 increases the goodness. A twiddle of 1.0 or greater decreases
//goodness.
func (p *humanSolveItem) Twiddle(amount probabilityTweak, description string) {
	if amount < 0.0 {
		return
	}
	//Ignore twiddles once we've been told to not expect any more.
	if p.doneTwiddling {
		return
	}
	p.twiddles = append(p.twiddles, twiddleRecord{description, amount})
	p.searcher.ItemValueChanged(p)
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

//NextSearchWorkItem returns the next humanSolveWorkItem in this item to do:
//the techinque to run on a given grid. If no more work is left to be done,
//returns nil.
func (p *humanSolveItem) NextSearchWorkItem() *humanSolveWorkItem {
	//TODO: the use of effectiveTechniquesToUse here is another nail in the
	//coffin for treaing guess specially.

	techniquesToUse := p.searcher.options.effectiveTechniquesToUse()

	if p.techniqueIndex >= len(techniquesToUse) {
		return nil
	}

	result := &humanSolveWorkItem{
		grid:      p.Grid(),
		technique: techniquesToUse[p.techniqueIndex],
	}
	p.techniqueIndex++
	return result

}

/************************************************************
 *
 * humanSolveSearcher implementation
 *
 ************************************************************/

func newHumanSolveSearcher(grid Grid, previousCompoundSteps []*CompoundSolveStep, options *HumanSolveOptions, stepsCache *foundStepCache) *humanSolveSearcher {
	searcher := &humanSolveSearcher{
		grid:                  grid,
		options:               options,
		previousCompoundSteps: previousCompoundSteps,
		done:       make(chan bool),
		stepsCache: stepsCache,
	}
	heap.Init(&searcher.itemsToExplore)
	initialItem := &humanSolveItem{
		searcher:  searcher,
		heapIndex: -1,
	}
	heap.Push(&searcher.itemsToExplore, initialItem)
	return searcher
}

//Injects this item into to the searcher.
func (n *humanSolveSearcher) AddItem(item *humanSolveItem) {
	//make sure we only add items once.
	if item.added {
		return
	}
	item.added = true
	n.itemsLock.Lock()

	if n.stepsCache != nil {

		//Only add items to the cache that are in the first round, because
		//later rounds could depend on earlier rounds that might actually be
		//applied.

		if item.parent != nil && item.parent.parent == nil {
			n.stepsCache.AddStepToQueue(item.step)
		}

	}

	if item.IsComplete() {
		n.completedItems = append(n.completedItems, item)
		if item.step.Technique != GuessTechnique {
			n.straightforwardItemsCount++
		}
	} else {
		heap.Push(&n.itemsToExplore, item)
	}
	n.itemsLock.Unlock()
}

func (n *humanSolveSearcher) ItemValueChanged(item *humanSolveItem) {
	if item.heapIndex < 0 {
		return
	}
	n.itemsLock.Lock()
	heap.Fix(&n.itemsToExplore, item.heapIndex)
	n.itemsLock.Unlock()
}

//DoneSearching will return true when no more items need to be explored
//because we have enough CompletedItems.
func (n *humanSolveSearcher) DoneSearching() bool {
	if n.options == nil {
		close(n.done)
		return true
	}

	n.itemsLock.Lock()
	lenCompletedItems := len(n.completedItems)
	n.itemsLock.Unlock()

	result := n.options.NumOptionsToCalculate <= lenCompletedItems
	if result {
		n.signalDone()
	}
	return result
}

func (n *humanSolveSearcher) signalDone() {
	n.itemsLock.Lock()
	if !n.channelClosed {
		close(n.done)
		n.channelClosed = true
	}
	n.itemsLock.Unlock()
}

//NextPossibleStep pops the best step and returns it.
func (n *humanSolveSearcher) NextPossibleStep() *humanSolveItem {
	if n.itemsToExplore.Len() == 0 {
		return nil
	}

	n.itemsLock.Lock()
	result := heap.Pop(&n.itemsToExplore).(*humanSolveItem)
	n.itemsLock.Unlock()

	return result
}

//Search is the main workhorse of HumanSolve Search, which explores all of
//the itemsToExplore (potentially bailing early if enough completed items are
//found). When Search is done, searcher.completedItems will contain the
//possibilities to choose from.
func (n *humanSolveSearcher) Search() {

	/*
		The pipeline starts by generating humanSolveWorkItems, and at the
		end collects generated CompoundSolveSteps and puts them in
		searcher.completedItems.

		The pipeline continues until one of the following things are true:

		1) No more work items will be generated. This is reasonably rare
		in practice, because as long as Guess is in the set of
		TechniquesToUse there will almost always be SOME item. When this
		shuts down the pipeline is already mostly idle anyway so it's just
		a matter of tidying up. However, this will always happen in the
		last few steps of solving a puzzle when there's only one move to
		make anyway.

		2) We have at least NumItemsToCompute items in
		searcher.completedItems (or some other more complex early exit
		logic is true) and thus can exit early. When this happens the
		pipeline is roaring through all of the work and needs to signal
		all pieces to shut down. We handle this by defering a close to
		allDone in this method and then just returning.

		The pipeline consists of the following go Routines:

		1) A routine to generate humanSolveWorkItems. It loops through
		searcher.NextPossibleStep, and for each one of those loops through
		NextWorkItem until none are left. It sends workItems down the
		channel to the next stage. Once there are no more steps it closes
		the outbound channel, signalling to the rest of the pipeline to
		exit in Exit Condition #1. If it can receive from the allDone
		channel, that means that Exit Condition #2 is met and it should
		begin an early shutdown and close its output channel.

		Each work item contains the grid to operate on, the technique, and
		a coordinator that is specific to this baseItem.

		2) A series of N worker threads that take an item off of
		workItems, run the technique, and then run another one. The
		coordinator in the work item synchronously adds the result to the
		searcher . If workItems is closed they immediately exit. The
		Technique.Find() methods will early exit if the coordinator tells
		them to (and it will do so faster given its synchronous nature)

		On the main thread we watch for either the searcher to signal that
		DoneSearching is called (by closing its done channel), or for all
		of the solver threads to have quit, signaling that all of the
		techniques have been exhausted.
	*/

	//TODO: make sure that Guess will return at least one guess item in all
	//cases, but never will go above the normal rank of 2 unless there are
	//none of size 2. This will require a new test. Note that all guesses are
	//infinitely bad, which means that guess on a cell of rank 2 and guess on
	//a cell of rank 3 will be equally bad, making it more important to only
	//return cells of the lowest rank.

	//TODO: make this configurable

	//TODO: test if this is faster on devices with other numbers of cores or
	//if it's just tuned to the Mac Pro.
	runtime.GOMAXPROCS(runtime.NumCPU())
	numFindThreads := runtime.GOMAXPROCS(0)/2 - 1

	if numFindThreads < 2 {
		numFindThreads = 2
	}

	workItems := make(chan *humanSolveWorkItem)

	//The thread to generate work items
	go humanSolveSearcherWorkItemGenerator(n, workItems)

	var solveThreadsDone sync.WaitGroup

	solveThreadsDone.Add(numFindThreads)

	for i := 0; i < numFindThreads; i++ {
		go humanSolveSearcherFindThread(workItems, &solveThreadsDone)
	}

	allSolveThreadsDone := make(chan bool)

	//Convert the wait group into a channel send for convenience of using
	//select{} below.
	go func() {
		solveThreadsDone.Wait()
		allSolveThreadsDone <- true
	}()

	//Wait for either all solve threads to have finished (meaning they have
	//found everything they're going to find) or for ourselves to have
	//signaled that we reached DoneSearching().
	select {
	case <-allSolveThreadsDone:
	case <-n.done:
	}

}

//humanSolveSearcherFindThread is a thread that takes in workItems and runs
//the specified technique on the specified grid.
func humanSolveSearcherFindThread(workItems chan *humanSolveWorkItem, wg *sync.WaitGroup) {
	for workItem := range workItems {
		workItem.technique.find(workItem.grid, workItem.coordinator)
	}
	wg.Done()
}

//humanSolveSearcherWorkItemGenerator is used in searcher.Search to generate
//the stream of WorkItems.
func humanSolveSearcherWorkItemGenerator(searcher *humanSolveSearcher, workItems chan *humanSolveWorkItem) {
	//When we return close down workItems to signal downstream things to
	//close.
	defer close(workItems)

	//We'll loop through each step in searcher, and then for each step
	//generate a work item per technique.

	item := searcher.NextPossibleStep()

	for item != nil {

		coordinator := &synchronousFindCoordinator{
			searcher: searcher,
			baseItem: item,
		}

		workItem := item.NextSearchWorkItem()

		for workItem != nil {

			//Tell each workItem where to send its results
			workItem.coordinator = coordinator

			select {
			case workItems <- workItem:
			case <-searcher.done:
				return
			}

			workItem = item.NextSearchWorkItem()
		}

		item = searcher.NextPossibleStep()

	}
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

/************************************************************
 *
 * humanSolveSearcherHeap implementation
 *
 ************************************************************/

//Len is necessary to implement heap.Interface
func (n humanSolveSearcherHeap) Len() int {
	return len(n)
}

//Less is necessary to implement heap.Interface
func (n humanSolveSearcherHeap) Less(i, j int) bool {
	return n[i].Goodness() < n[j].Goodness()
}

//Swap is necessary to implement heap.Interface
func (n humanSolveSearcherHeap) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
	n[i].heapIndex = i
	n[j].heapIndex = j
}

//Push is necessary to implement heap.Interface. It should not be used
//direclty; instead, use heap.Push()
func (n *humanSolveSearcherHeap) Push(x interface{}) {
	length := len(*n)
	item := x.(*humanSolveItem)
	item.heapIndex = length
	*n = append(*n, item)
}

//Pop is necessary to implement heap.Interface. It should not be used
//directly; instead use heap.Pop()
func (n *humanSolveSearcherHeap) Pop() interface{} {
	old := *n
	length := len(old)
	item := old[length-1]
	item.heapIndex = -1 // for safety
	*n = old[0 : length-1]
	return item
}

func humanSolvePossibleStepsImpl(grid Grid, options *HumanSolveOptions, previousSteps []*CompoundSolveStep, stepsCache *foundStepCache) (steps []*CompoundSolveStep, distribution ProbabilityDistribution) {
	//TODO: with the new approach, we're getting a lot more extreme negative difficulty values. Train a new model!

	//TODO: do something with stepsCache.

	//We send a copy here because our own selves will likely be modified soon
	//after returning from this, and if the other threads haven't gotten the
	//signal yet to shut down they might get in a weird state.
	searcher := newHumanSolveSearcher(grid, previousSteps, options, stepsCache)

	searcher.Search()

	//Prepare the distribution and list of steps

	//But first check if we don't have any.
	if len(searcher.completedItems) == 0 {
		return nil, nil
	}

	//Get a consistent snapshot of completedItems; its length might change.
	completedItems := searcher.completedItems

	distri := make(ProbabilityDistribution, len(completedItems))
	var resultSteps []*CompoundSolveStep

	for i, item := range completedItems {
		distri[i] = item.Goodness()
		compoundStep := newCompoundSolveStep(item.Steps())
		compoundStep.explanation = item.explainGoodness()
		resultSteps = append(resultSteps, compoundStep)
	}

	invertedDistribution := distri.invert()

	return resultSteps, invertedDistribution

}

func (self *gridImpl) humanSolvePossibleStepsWithCache(options *HumanSolveOptions, previousSteps []*CompoundSolveStep, stepsCache *foundStepCache) (steps []*CompoundSolveStep, distribution ProbabilityDistribution) {
	return humanSolvePossibleStepsImpl(self, options, previousSteps, stepsCache)
}

func (self *mutableGridImpl) humanSolvePossibleStepsWithCache(options *HumanSolveOptions, previousSteps []*CompoundSolveStep, stepsCache *foundStepCache) (steps []*CompoundSolveStep, distribution ProbabilityDistribution) {
	return humanSolvePossibleStepsImpl(self.Copy(), options, previousSteps, stepsCache)
}

func (self *gridImpl) HumanSolvePossibleSteps(options *HumanSolveOptions, previousSteps []*CompoundSolveStep) (steps []*CompoundSolveStep, distribution ProbabilityDistribution) {
	return self.humanSolvePossibleStepsWithCache(options, previousSteps, nil)
}

func (self *mutableGridImpl) HumanSolvePossibleSteps(options *HumanSolveOptions, previousSteps []*CompoundSolveStep) (steps []*CompoundSolveStep, distribution ProbabilityDistribution) {
	return self.humanSolvePossibleStepsWithCache(options, previousSteps, nil)
}
