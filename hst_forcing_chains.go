package sudoku

import (
	"container/list"
	"fmt"
	"log"
	"strconv"
)

type forcingChainsTechnique struct {
	*basicSolveTechnique
}

func (self *forcingChainsTechnique) HumanLikelihood() float64 {
	//TODO: figure out what the baseDifficulty should be
	return self.difficultyHelper(200.0)
}

func (self *forcingChainsTechnique) Description(step *SolveStep) string {
	//TODO: implement this
	return "ERROR: NOT IMPLEMENTED"
}

func (self *forcingChainsTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {
	//TODO: test that this will find multiple if they exist.
	//TODO: Implement this.

	getter := grid.queue().DefaultGetter()

	_MAX_IMPLICATION_STEPS := 6

	for {

		//Check if it's time to stop.
		select {
		case <-done:
			return
		default:
		}

		candidate := getter.GetSmallerThan(3)

		if candidate == nil {
			break
		}

		candidateCell := candidate.(*Cell)

		if len(candidateCell.Possibilities()) != 2 {
			//We found one with 1 possibility, which isn't interesting for us--nakedSingle should do that one.
			continue
		}

		firstPossibilityNum := candidateCell.Possibilities()[0]
		secondPossibilityNum := candidateCell.Possibilities()[1]

		firstGrid := grid.Copy()
		secondGrid := grid.Copy()

		//Check that the neighbor isn't just already having a single possibility, because then this technique is overkill.

		firstAccumulator := makeChainSeacherAccumulator(_MAX_IMPLICATION_STEPS)
		secondAccumulator := makeChainSeacherAccumulator(_MAX_IMPLICATION_STEPS)

		chainSearcher(_MAX_IMPLICATION_STEPS,
			candidateCell.InGrid(firstGrid),
			firstPossibilityNum,
			firstAccumulator)

		chainSearcher(_MAX_IMPLICATION_STEPS,
			candidateCell.InGrid(secondGrid),
			secondPossibilityNum,
			secondAccumulator)

		//TODO:Check if the sets overlap.

		doPrint := candidateCell.Row() == 1 && candidateCell.Col() == 0

		//For these debugging purposes, only print out the candidateCell we know to be interesting in the test case.
		if doPrint {
			log.Println(firstAccumulator)
			log.Println(secondAccumulator)
		}

		//See if either branch, at some generation, has the same cell forced to the same number in either generation.

		if doPrint {
			log.Println("Accumulators after accumulating generations:")
			log.Println(firstAccumulator)
			log.Println(secondAccumulator)
		}

		foundOne := false

		for generation := _MAX_IMPLICATION_STEPS - 1; generation >= 0 && !foundOne; generation-- {

			//Check for any overlap at the last generation
			firstAffectedCells := firstAccumulator[generation]
			secondAffectedCells := secondAccumulator[generation]

			for key, val := range firstAffectedCells {

				//Skip the candidateCell, because that's not a meaningful overlap--we set that one as a way of branching!
				if key == candidateCell.ref() {
					continue
				}

				if num, ok := secondAffectedCells[key]; ok {
					//Found cell overlap! ... is the forced number the same?
					if val == num {
						//Yup, seems like we've found a cell that is forced to the same value on either branch.
						step := &SolveStep{self,
							CellSlice{key.Cell(grid)},
							IntSlice{val},
							CellSlice{candidateCell},
							candidateCell.Possibilities(),
						}

						if doPrint {
							log.Println(step)
							log.Println("Candidate Cell", candidateCell.ref())
						}

						if step.IsUseful(grid) {
							foundOne = true
							if doPrint {
								log.Println("Found solution on generation: ", generation)
							}
							select {
							case results <- step:
							case <-done:
								return
							}
						}
					}
				}
			}
		}

		//TODO: figure out why the tests are coming back with different answers, even when only looking at the key cell
		//that should work from the example.
		//TODO: we should prefer solutions where the total implications on both branches are minimized.
		//For example, if only one implication is requried on left, but 4 are on right, that's preferable to one where
		//three implications are required on both sides.
		//TODO: figure out a way to only compute a generation if required on each branch (don't compute all the way to _MAX_IMPLICATIONS to start)

	}
}

type chainSearcherGenerationDetails map[cellRef]int

func (c chainSearcherGenerationDetails) String() string {
	result := "Begin map\n"
	for cell, num := range c {
		result += "\t" + cell.String() + " : " + strconv.Itoa(num) + "\n"
	}
	result += "End map\n"
	return result
}

type chainSearcherAccumulator []chainSearcherGenerationDetails

func (c chainSearcherAccumulator) String() string {
	result := "Accumulator[\n"
	for _, rec := range c {
		result += fmt.Sprintf("%s\n", rec)
	}
	result += "]\n"
	return result
}

func makeChainSeacherAccumulator(size int) chainSearcherAccumulator {
	result := make(chainSearcherAccumulator, size)
	for i := 0; i < size; i++ {
		result[i] = make(map[cellRef]int)
	}
	return result
}

func chainSearcher(i int, cell *Cell, numToApply int, accumulator chainSearcherAccumulator) {

	//TODO: rename the i paramater to max generations
	//TODO: generations should count UP, not down.

	//Chainsearcher implements a BFS over implications forward given the starting point.
	//It collects its results in the provided chainSearcherAccumulator.

	//TODO: why doesn't this just return its own chainSearcherAccumulator.

	//the first time we cross over into a new generation, we should do a one-time copy of the old generation
	//into the new.
	//At any write, if we notice that we'd be overwriting to a different value, we can bail out (how would
	//we mark that we bailed early), since we've run into an inconsistency down this branch and following
	//it further is not useful.

	type modificationToMake struct {
		generation int
		cell       *Cell
		numToApply int
	}

	workSteps := list.New()

	//Add the first workstep.
	workSteps.PushBack(modificationToMake{
		i,
		cell,
		numToApply,
	})

	var step modificationToMake

	e := workSteps.Front()

	lastSeenGeneration := i

	for e != nil {

		workSteps.Remove(e)

		switch t := e.Value.(type) {
		case modificationToMake:
			step = t
		default:
			panic("Found unexpected type in workSteps list")
		}

		if step.generation <= 0 {
			break
		}

		generationDetails := accumulator[step.generation-1]

		if step.generation != lastSeenGeneration {
			//We just crossed the generation boundary. Copy the generation above's data into ours before we start populating it.
			lastGenerationDetails := accumulator[step.generation]
			for key, val := range lastGenerationDetails {
				generationDetails[key] = val
			}
		}

		cellsToVisit := step.cell.Neighbors().FilterByPossible(step.numToApply).FilterByNumPossibilities(2)

		step.cell.SetNumber(step.numToApply)

		//TODO: check here if we're overwriting a value; if so, don't process anymore work steps.
		generationDetails[step.cell.ref()] = step.numToApply

		for _, cellToVisit := range cellsToVisit {
			possibilities := cellToVisit.Possibilities()

			if len(possibilities) != 1 {
				panic("Expected the cell to have one possibility")
			}

			forcedNum := possibilities[0]

			//Each branch modifies the grid, so create a new copy
			newGrid := cellToVisit.grid.Copy()
			cellToVisit = cellToVisit.InGrid(newGrid)

			workSteps.PushBack(modificationToMake{
				step.generation - 1,
				cellToVisit,
				forcedNum,
			})

		}

		e = workSteps.Front()

	}

}
