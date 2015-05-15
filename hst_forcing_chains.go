package sudoku

import (
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

		//accumulate forward, so the last generation has ALL cells affected in any generation on that branch
		firstAccumulator.accumulateGenerations()
		secondAccumulator.accumulateGenerations()

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

//accumulateGenerations goes through each generation (youngest to newest)
//and squaches older generation maps into each generation, so each
//generation's map represents the totality of all cells seen at that point.
func (c chainSearcherAccumulator) accumulateGenerations() {
	for i := len(c) - 2; i >= 0; i-- {
		lastGeneration := c[i+1]
		currentGeneration := c[i]
		for key, val := range lastGeneration {
			currentGeneration[key] = val
		}
	}
}

func makeChainSeacherAccumulator(size int) chainSearcherAccumulator {
	result := make(chainSearcherAccumulator, size)
	for i := 0; i < size; i++ {
		result[i] = make(map[cellRef]int)
	}
	return result
}

func chainSearcher(i int, cell *Cell, numToApply int, accumulator chainSearcherAccumulator) {
	if i <= 0 || cell == nil {
		//Base case
		return
	}

	if i-1 >= len(accumulator) {
		panic("The accumulator provided was not big enough for the i provided.")
	}

	generationDetails := accumulator[i-1]

	//Find the nextCells that WILL have their numbers forced by the cell we're thinking of fillint.
	cellsToVisit := cell.Neighbors().FilterByPossible(numToApply).FilterByNumPossibilities(2)

	//Now that we know which cells will be affected and what their next number will be,
	//set the number in the given cell and then recurse downward down each branch.
	cell.SetNumber(numToApply)

	generationDetails[cell.ref()] = numToApply

	for _, cellToVisit := range cellsToVisit {

		possibilities := cellToVisit.Possibilities()

		if len(possibilities) != 1 {
			panic("Expected the cell to have one possibility")
		}

		forcedNum := possibilities[0]

		//Each branch modifies the grid, so create a new copy
		newGrid := cellToVisit.grid.Copy()
		cellToVisit = cellToVisit.InGrid(newGrid)

		//Recurse downward
		chainSearcher(i-1, cellToVisit, forcedNum, accumulator)

	}

}
