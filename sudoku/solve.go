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
		stackDone := make(chan bool, 1)

		stack := NewChanSyncedStack(stackDone)

		stack.Insert(self.Copy())

		incomingSolutions := make(chan *Grid)

		var solutions []*Grid

		for i := 0; i < NUM_SOLVER_THREADS; i++ {
			go func() {
				//Sovler thread.
				for {
					grid, ok := <-stack.Output
					if !ok {
						return
					}
					result := grid.(*Grid).searchSolutions(stack)
					if result != nil {
						incomingSolutions <- result
					}
					stack.ItemDone()
				}
			}()
		}

	OuterLoop:
		for {
			select {
			case solution := <-incomingSolutions:
				//Add it to results
				solutions = append(solutions, solution)
				if len(solutions) >= max {
					stack.Dispose()
					break OuterLoop
				}
			case <-stackDone:
				//Well, that's as good as it's going to get.
				break OuterLoop
			}
		}

		self.cachedSolutions = solutions

	}

	return self.cachedSolutions

}

func (self *Grid) searchSolutions(stack *ChanSyncedStack) *Grid {
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
	rankedObject := self.queue.DefaultGetter().Get()
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

	for _, num := range possibilities {
		copy := self.Copy()
		copy.Cell(cell.Row, cell.Col).SetNumber(num)
		stack.Insert(copy)
	}

	return nil

}

//Fills in all of the cells it can without branching or doing any advanced
//techniques that require anything more than a single cell's possibles list.
func (self *Grid) fillSimpleCells() int {
	count := 0
	getter := self.queue.DefaultGetter()
	obj := getter.GetSmallerThan(2)
	for obj != nil && !self.cellsInvalid() {
		cell, ok := obj.(*Cell)
		if !ok {
			continue
		}
		cell.SetNumber(cell.implicitNumber())
		count++
		obj = getter.GetSmallerThan(2)
	}
	return count
}
