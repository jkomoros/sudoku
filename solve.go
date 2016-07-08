package sudoku

import (
	"math/rand"
	"sync"
)

//Solve searches for a solution to the puzzle as it currently exists without
//unfilling any cells. If one exists, it will fill in all cells to fit that
//solution and return true. If there are no solutions the grid will remain
//untouched and it will return false. If multiple solutions exist, Solve will pick one at random.
func (self *Grid) Solve() bool {

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

//NumSolutions returns the total number of solutions found in the grid when it is solved forward
//from this point. A valid Sudoku puzzle has only one solution. Does not mutate the grid.
func (self *Grid) NumSolutions() int {
	return len(self.Solutions())
}

//HasSolution returns true if the grid has at least one solution. Does not mutate the grid.
func (self *Grid) HasSolution() bool {
	//TODO: optimize this to bail as soon as we find a single solution.
	return len(self.nOrFewerSolutions(1)) > 0
}

//HasMultipleSolutions returns true if the grid has more than one solution. Does not mutate the grid.
func (self *Grid) HasMultipleSolutions() bool {
	return len(self.nOrFewerSolutions(2)) >= 2
}

//Solutions returns a slice of grids that represent possible solutions if you were to solve forward this grid. The current grid is not modified.
//If there are no solutions forward from this location it will return a slice with len() 0. Does not mutate the grid.
func (self *Grid) Solutions() (solutions []*Grid) {
	return self.nOrFewerSolutions(0)
}

//The actual workhorse of solutions generating. 0 means "as many as you can find". It might return more than you asked for, if it already had more results than requested sitting around.
func (self *Grid) nOrFewerSolutions(max int) []*Grid {

	self.cachedSolutionsLock.RLock()
	hasNoCachedSolutions := self.cachedSolutions == nil
	cachedSolutionsLen := self.cachedSolutionsRequestedLength
	self.cachedSolutionsLock.RUnlock()

	if hasNoCachedSolutions || (max == 0 && cachedSolutionsLen != 0) || (max > 0 && cachedSolutionsLen < max && cachedSolutionsLen > 0) {

		queueDone := make(chan bool, 1)

		queue := newSyncedFiniteQueue(0, DIM*DIM, queueDone)

		queue.In <- self.Copy()

		//In the past this wasn't buffered, but then when we finished early other items would try to go into it
		//and block, which prevented them from looping back up and getting the signal to shut down.
		//Since there's only a known number of threads, we'll make sure they all ahve a place to leave their work
		//without blocking so they can get the signal to shut down.
		incomingSolutions := make(chan *Grid, _NUM_SOLVER_THREADS)

		//Using a pattern for closing fan in style receivers from http://blog.golang.org/pipelines
		var wg sync.WaitGroup

		var solutions []*Grid

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
					result := grid.(*Grid).searchSolutions(queue, firstRun, max)
					grid.(*Grid).Done()
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

		self.cachedSolutionsLock.Lock()
		self.cachedSolutions = solutions
		self.cachedSolutionsRequestedLength = max
		self.cachedSolutionsLock.Unlock()

	}

	//TODO: rejigger this to not need a write lock then a read lock when setting.
	self.cachedSolutionsLock.RLock()
	result := self.cachedSolutions
	self.cachedSolutionsLock.RUnlock()
	return result

}

func (self *Grid) searchSolutions(queue *syncedFiniteQueue, isFirstRun bool, numSoughtSolutions int) *Grid {
	//This will only be called by Solutions.
	//We will return ourselves if we are a solution, and if not we will return nil.
	//If there are any sub children, we will send them to counter before we're done.

	if self.Invalid() {
		return nil
	}

	self.fillSimpleCells()
	//Have any cells noticed they were invalid while solving forward?
	if self.cellsInvalid() {
		return nil
	}

	if self.Solved() {
		return self
	}

	//Well, looks like we're going to have to branch.
	rankedObject := self.queue().NewGetter().Get()
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

	var result *Grid

	for i, num := range possibilities {
		copy := self.Copy()
		copy.Cell(cell.Row(), cell.Col()).SetNumber(num)
		//As an optimization for cases where there are many solutions, we'll just continue a DFS until we barf then unroll back up.
		//It doesn't appear to slow things down in the general case
		if i == 0 && !isFirstRun {
			result = copy.searchSolutions(queue, false, numSoughtSolutions)
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

//Fills in all of the cells it can without branching or doing any advanced
//techniques that require anything more than a single cell's possibles list.
func (self *Grid) fillSimpleCells() int {
	count := 0
	getter := self.queue().NewGetter()
	obj := getter.GetSmallerThan(2)
	for obj != nil && !self.cellsInvalid() {
		cell, ok := obj.(Cell)
		if !ok {
			continue
		}
		cell.SetNumber(cell.implicitNumber())
		count++
		obj = getter.GetSmallerThan(2)
	}
	return count
}
