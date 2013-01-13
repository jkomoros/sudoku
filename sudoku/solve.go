package sudoku

import (
	"math/rand"
)

//Searches for a solution to the puzzle as it currently exists without
//unfilling any cells. If one exists, it will fill in all cells to fit that
//solution and return true. If there are no solutions the grid will remain
//untouched and it will return false.
func (self *Grid) Solve() bool {
	//TODO: Optimization: we only need one, so we can bail as soon as we find a single one.
	solutions := self.Solutions()
	if len(solutions) == 0 {
		return false
	}
	self.Load(solutions[0].DataString())
	return true
}

//Returns the total number of solutions found in the grid. Does not mutate the grid.
func (self *Grid) NumSolutions() int {
	return len(self.Solutions())
}

//Returns true if the grid has at least one solution. Does not mutate the grid.
func (self *Grid) HasSolution() bool {
	//TODO: optimize this to bail as soon as we find a single solution.
	return len(self.nOrFewerSolutions(1)) > 0
}

func (self *Grid) HasMultipleSolutions() bool {
	return len(self.nOrFewerSolutions(2)) >= 2
}

//Returns a slice of grids that represent possible solutions if you were to solve forward this grid. The current grid is not modified.
//If there are no solutions forward from this location it will return a slice with len() 0.
func (self *Grid) Solutions() (solutions []*Grid) {
	return self.nOrFewerSolutions(0)
}

//The actual workhorse of solutions generating. 0 means "as many as you can find". It might return more than you asked for, if it already had more results than requested sitting around.
func (self *Grid) nOrFewerSolutions(max int) []*Grid {
	if self.cachedSolutions == nil || (max > 0 && len(self.cachedSolutions) < max) {

		//We'll have a thread that's keeping track of how many grids need to be processed and how many have responded.
		inGrids := make(chan *Grid)
		outGrids := make(chan *Grid)

		//The way for us to signify to the worker threads to kill themselves.
		exit := make(chan bool, NUM_SOLVER_THREADS)

		//Where we'll store the grids that are yet to be processed.
		//We used to use a SyncedFiniteQueue here, but it was unnecessary now that we explore a thread before
		//spinning off other threads, since that already approximates a DFS.
		gridsToProcess := make(chan *Grid, NUM_SOLVER_THREADS*DIM*DIM)

		//If you can grab a true from this then you're the first searchSolutions to run. In that case, you should
		//spin off more work for everybody else rather than diving down to the bottom.
		isFirst := make(chan bool, 1)

		counter := 0

		//Kick off NUM_SOLVER_THREADS
		for i := 0; i < NUM_SOLVER_THREADS; i++ {
			go func() {
				var firstRun bool
				for {
					select {
					case <-exit:
						return
					case grid := <-gridsToProcess:
						select {
						case <-isFirst:
							firstRun = true
						default:
							firstRun = false
						}
						outGrids <- grid.searchSolutions(inGrids, max, firstRun)
					}
				}
			}()
		}

		//How the counter loop will tell us that we've met the final conditions.
		results := make(chan []*Grid)

		//Kick off the main counter loop. This will also ensure that we clean up nicely after ourselves and drain all chanenls.
		go func() {
			//This will accept at least one grid, and after that when the counter is 0 it will exit.
			exiting := false
			var workingSolutions []*Grid
			for {
				select {
				case inGrid := <-inGrids:
					//If we're exiting no need to put more work in the queue.
					if !exiting {
						counter++
						//We've already done the critical counting; put another thing on the thread but don't wait for it because it may block.
						gridsToProcess <- inGrid
					}
				case outGrid := <-outGrids:
					counter--
					if outGrid != nil {
						workingSolutions = append(workingSolutions, outGrid)
					}
					if max > 0 && len(workingSolutions) >= max {
						//We can early exit but we need to continue consuming stuff off the channels so we don't leak.
						exiting = true
						results <- workingSolutions
						workingSolutions = nil
					}
					if counter == 0 {
						if !exiting {
							//There wasn't an early exit, so we still need to signal up with the result.
							results <- workingSolutions
						}
						//And now we die.
						return
					}
				}
			}
		}()

		//Feed in the first work item and...
		inGrids <- self.Copy()
		//...wait for the results.
		tempSolutions := <-results

		//Kill NUM_SOLVER_THREADS processes
		for i := 0; i < NUM_SOLVER_THREADS; i++ {
			//Because exit is buffered, we won't have to wait for all threads to acknowledge the kill order before proceeding.
			//...I'm not entirely sure why we don't have the earlier Fill() problem where other threads would try to post
			//after the main thread was done. I think now we just have a garbage collected, constantly stuck problem.
			exit <- true
		}

		self.cachedSolutions = tempSolutions

	}

	return self.cachedSolutions
}

func (self *Grid) searchSolutions(gridsToProcess chan *Grid, numSoughtSolutions int, firstRun bool) *Grid {
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
	rankedObject := self.queue.Get()
	if rankedObject == nil {
		panic("Queue didn't have any cells.")
	}

	cell, ok := rankedObject.(*Cell)
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
		copy.Cell(cell.Row, cell.Col).SetNumber(num)
		if i == 0 && !firstRun {
			//We'll do the last one ourselves
			result = copy.searchSolutions(gridsToProcess, numSoughtSolutions, false)
			if result != nil && numSoughtSolutions == 1 {
				//No need to spin off other branches, just return up.
				return result
			}
		} else {
			//But all of the other ones we'll spin off so other threads can take them.
			gridsToProcess <- copy
		}

	}

	return result

}

//Fills in all of the cells it can without branching or doing any advanced
//techniques that require anything more than a single cell's possibles list.
func (self *Grid) fillSimpleCells() int {
	count := 0
	obj := self.queue.GetSmallerThan(2)
	for obj != nil && !self.cellsInvalid() {
		cell, ok := obj.(*Cell)
		if !ok {
			continue
		}
		cell.SetNumber(cell.implicitNumber())
		count++
		obj = self.queue.GetSmallerThan(2)
	}
	return count
}
