package sudoku

import (
	"math/rand"
	"sync"
)

func (self *mutableGridImpl) Solve() bool {

	//Special case; check if it's already solved.
	//TODO: removing this causes Solve, when called on an already solved grid, to sometimes fail. Why is that?
	if self.Solved() {
		return true
	}

	//TODO: Optimization: we only need one, so we can bail as soon as we find a single one.
	solutions := self.Solutions()
	if len(solutions) == 0 {
		return false
	}
	self.LoadSDK(solutions[0].DataString())
	return true
}

func (self *gridImpl) NumSolutions() int {
	//TODO: implement this!
	return 0
}

func (self *mutableGridImpl) NumSolutions() int {
	return len(self.Solutions())
}

func (self *gridImpl) HasSolution() bool {
	//TODO: implement this!
	return false
}

func (self *mutableGridImpl) HasSolution() bool {
	//TODO: optimize this to bail as soon as we find a single solution.
	return len(self.nOrFewerSolutions(1)) > 0
}

func (self *gridImpl) HasMultipleSolutions() bool {
	//TODO: implement this!
	return false
}

func (self *mutableGridImpl) HasMultipleSolutions() bool {
	return len(self.nOrFewerSolutions(2)) >= 2
}

func (self *gridImpl) Solutions() []Grid {
	//TODO: implement this
	return nil
}

func (self *mutableGridImpl) Solutions() (solutions []Grid) {
	return self.nOrFewerSolutions(0)
}

//The actual workhorse of solutions generating. 0 means "as many as you can
//find". It might return more than you asked for, if it already had more
//results than requested sitting around.
func (self *mutableGridImpl) nOrFewerSolutions(max int) []Grid {

	self.cachedSolutionsLockRef.RLock()
	hasNoCachedSolutions := self.cachedSolutionsRef == nil
	cachedSolutionsLen := self.cachedSolutionsRequestedLengthRef
	self.cachedSolutionsLockRef.RUnlock()

	if hasNoCachedSolutions || (max == 0 && cachedSolutionsLen != 0) || (max > 0 && cachedSolutionsLen < max && cachedSolutionsLen > 0) {

		queueDone := make(chan bool, 1)

		queue := newSyncedFiniteQueue(0, DIM*DIM, queueDone)

		queue.In <- self.Copy()

		//In the past this wasn't buffered, but then when we finished early other items would try to go into it
		//and block, which prevented them from looping back up and getting the signal to shut down.
		//Since there's only a known number of threads, we'll make sure they all ahve a place to leave their work
		//without blocking so they can get the signal to shut down.
		incomingSolutions := make(chan Grid, _NUM_SOLVER_THREADS)

		//Using a pattern for closing fan in style receivers from http://blog.golang.org/pipelines
		var wg sync.WaitGroup

		var solutions []Grid

		//TODO: figure out a way to kill all of these threads when necessary.

		//TODO: don't use a constant here, use someting around the lines of
		//numCPU

		for i := 0; i < _NUM_SOLVER_THREADS; i++ {
			go func() {
				//Sovler thread.
				firstRun := true
				defer wg.Done()
				for {
					grid, ok := <-queue.Out
					if !ok {
						return
					}
					result := searchGridSolutions(grid.(Grid), queue, firstRun, max)
					if result != nil {
						incomingSolutions <- result
					}
					queue.ItemDone <- true
					firstRun = false
				}
			}()
		}

		wg.Add(_NUM_SOLVER_THREADS)

		go func() {
			wg.Wait()
			close(incomingSolutions)
		}()

	OuterLoop:
		for {
			select {
			case solution := <-incomingSolutions:
				//incomingSolutions must have been closed because no more work to do.
				if solution == nil {
					break OuterLoop
				}
				//Add it to results
				solutions = append(solutions, solution)
				if max > 0 && len(solutions) >= max {
					break OuterLoop
				}
			case <-queueDone:
				//Well, that's as good as it's going to get.
				break OuterLoop
			}
		}

		//In some cases that previous select would have something incoming on
		//incomingsolutions, as well as on queueDone, and queueDone would have
		//just so happened to have won. Check for one last remaining item
		//coming in from incomingSolutions. Technically it's possible (how?)
		//to have multiple items waiting on incomingSolutions, so read as many
		//as we can get without blocking.(Not checking for this was the reason
		//for bug #134.)

	DoneReading:
		for {
			select {
			case solution := <-incomingSolutions:
				if solution != nil {
					solutions = append(solutions, solution)
				}
			default:
				//Nope, guess there wasn't one left.
				break DoneReading
			}
		}

		//There might be some things waiting to go into incomingSolutions here, but because it has a slot
		//for every thread to be buffered, it's OK, we can just stop now.

		//TODO: the grids waiting in the queue will never have their .Done called. This isn't a big deal--GC should reclaim them--
		//but we won't have as many that we could reuse.
		queue.Exit <- true

		self.cachedSolutionsLockRef.Lock()
		self.cachedSolutionsRef = solutions
		self.cachedSolutionsRequestedLengthRef = max
		self.cachedSolutionsLockRef.Unlock()

	}

	//TODO: rejigger this to not need a write lock then a read lock when setting.
	self.cachedSolutionsLockRef.RLock()
	result := self.cachedSolutionsRef
	self.cachedSolutionsLockRef.RUnlock()
	return result

}

func (self *gridImpl) searchSolutions(queue *syncedFiniteQueue, isFirstRun bool, numSoughtSolutions int) Grid {
	//TODO: implement this!
	return nil
}

func searchGridSolutions(grid Grid, queue *syncedFiniteQueue, isFirstRun bool, numSoughtSolutions int) Grid {
	//This will only be called by Solutions.
	//We will return ourselves if we are a solution, and if not we will return nil.
	//If there are any sub children, we will send them to counter before we're done.

	if grid.Invalid() {
		return nil
	}

	grid = withSimpleCellsFilled(grid)
	//Have any cells noticed they were invalid while solving forward?
	if grid.basicInvalid() {
		return nil
	}

	if grid.Solved() {
		return grid
	}

	//Well, looks like we're going to have to branch.
	rankedObject := grid.queue().NewGetter().Get()
	if rankedObject == nil {
		panic("Queue didn't have any cells.")
	}

	cell, ok := rankedObject.(Cell)
	if !ok {
		panic("We got back a non-cell from the grid's queue")
	}

	unshuffledPossibilities := cell.Possibilities()

	possibilities := make([]int, len(unshuffledPossibilities))

	for i, j := range rand.Perm(len(unshuffledPossibilities)) {
		possibilities[i] = unshuffledPossibilities[j]
	}

	var result Grid

	for i, num := range possibilities {
		//TODO: this seems like a natural place to use CopyWithModifications,
		//but gridImpl.fillSimpleCells will be called on it.
		copy := grid.MutableCopy()
		cell.MutableInGrid(copy).SetNumber(num)
		//As an optimization for cases where there are many solutions, we'll just continue a DFS until we barf then unroll back up.
		//It doesn't appear to slow things down in the general case
		if i == 0 && !isFirstRun {
			result = searchGridSolutions(copy, queue, false, numSoughtSolutions)
			if result != nil && numSoughtSolutions == 1 {
				//No need to spin off other branches, just return up.
				return result
			}
		} else {
			queue.In <- copy
		}
	}

	return result

}

//Returns a copy of Grid that has filled in all of the cells it can without
//branching or doing any advanced techniques that require anything more than a
//single cell's possibles list.
func withSimpleCellsFilled(grid Grid) Grid {

	//We fetch all of the cells that have a single possibility, then create a
	//copy with all of those filled. Then we repeat, because filling those
	//cells may have set other cells to now only have one possibility. Repeat
	//until no more cells with one possibility found.

	changesMade := true

	for changesMade {
		changesMade = false
		getter := grid.queue().NewGetter()
		obj := getter.GetSmallerThan(2)

		var modifications GridModifcation
		for obj != nil && !grid.basicInvalid() {
			cell, ok := obj.(MutableCell)
			if !ok {
				continue
			}
			changesMade = true
			modification := &CellModification{
				Cell:   cell,
				Number: cell.implicitNumber(),
			}
			modifications = append(modifications, modification)
			obj = getter.GetSmallerThan(2)
		}

		grid = grid.CopyWithModifications(modifications)

	}
	return grid
}
