package sudoku

import (
	"log"
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

		log.Println(firstAccumulator)
		log.Println(secondAccumulator)

		/*
			//Check pairwise through each of the result sets for each side and see if any overlap in an interestin way

			for i, theSet := range firstAffectedCellSets {
				theCellMapping := firstAffectedCellNums[i]
				for j, theSecondSet := range secondAffectedCellSets {
					theSecondCellMapping := secondAffectedCellNums[j]

					intersection := theSet.intersection(theSecondSet)
					if len(intersection) > 0 {
						//Okay, a cell overlapped... did they both set the same number?
						//TODO: should we look at all items that overlap if it's greater than 1?

						cell := intersection.toSlice()[0]

						if theCellMapping[cell] == theSecondCellMapping[cell] {
							//Booyah, found a step.

							step := &SolveStep{self,
								CellSlice{cell},
								IntSlice{theCellMapping[cell]},
								CellSlice{candidateCell},
								candidateCell.Possibilities(),
							}

							if step.IsUseful(grid) {
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
		*/

	}
}

type chainSearcherGenerationDetails struct {
	affectedCells cellSet
	filledNumbers map[*Cell]int
}

type chainSearcherAccumulator []*chainSearcherGenerationDetails

func makeChainSeacherAccumulator(size int) chainSearcherAccumulator {
	result := make(chainSearcherAccumulator, size)
	for i := 0; i < size; i++ {
		result[i] = &chainSearcherGenerationDetails{
			affectedCells: make(cellSet),
			filledNumbers: make(map[*Cell]int),
		}
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

	generationDetails.affectedCells[cell] = true
	generationDetails.filledNumbers[cell] = numToApply

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
