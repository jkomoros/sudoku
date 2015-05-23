package sudoku

import (
	"container/list"
	"fmt"
	"strconv"
)

type forcingChainsTechnique struct {
	*basicSolveTechnique
}

func (self *forcingChainsTechnique) HumanLikelihood() float64 {
	//TODO: figure out what the baseDifficulty should be, this might be higher than
	//it's actually in practice

	//Note that this number has to be pretty high because it's competing against
	//HiddenSIZEGROUP, which has the k exponential in its favor.
	return self.difficultyHelper(20000.0)
}

func (self *forcingChainsTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("cell %s only has two options, %s, and if you put either one in and see the chain of implications it leads to, both ones end up with %s in cell %s, so we can just fill that number in", step.PointerCells.Description(), step.PointerNums.Description(), step.TargetNums.Description(), step.TargetCells.Description())
}

func (self *forcingChainsTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {
	//TODO: test that this will find multiple if they exist.

	/*
	 * Conceptually this techinque chooses a cell with two possibilities
	 * and explores forward along two branches, seeing what would happen
	 * if it followed the simple implication chains forward to see if any
	 * cells end up set to the same number on both branches, meaning
	 * that no matter what, the cell will end up that value so you can set it
	 * that way now. In some ways it's like a very easy form of guessing.
	 *
	 * This techinque will do a BFS forward from the chosen cell, and won't
	 * explore more than _MAX_IMPLICATION_STEPS steps out from that. It will
	 * stop exploring if it finds one of two types of contradictions:
	 * 1) It notes that down this branch a single cell has had two different numbers
	 * implicated into it, which implies that somewhere earlier we ran into some inconsistency
	 * or
	 * 2) As soon as we note an inconsistency (a cell with no legal values).
	 *
	 * It is important to note that for every sudoku with one solution (that is, all
	 * legal puzzles), one of the two branches MUST lead to an inconsistency somewhere
	 * it's just a matter of how forward you have to go before you find it. That means
	 * that this technique is sensitive to the order in which you explore the frontiers
	 * of implications and when you choose to bail.
	 *
	 */

	getter := grid.queue().DefaultGetter()

	//TODO: investigate bumping this back up when #100 lands.
	_MAX_IMPLICATION_STEPS := 5

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

		firstAccumulator := chainSearcher(_MAX_IMPLICATION_STEPS,
			candidateCell.InGrid(firstGrid),
			firstPossibilityNum)

		secondAccumulator := chainSearcher(_MAX_IMPLICATION_STEPS,
			candidateCell.InGrid(secondGrid),
			secondPossibilityNum)

		//Cells that we've already vended and shouldn't vend again if we find another
		//TODO: figure out a better way to not vend duplicates. this method feels dirty.
		vendedCells := make(map[cellRef]bool)
		//don't vend the candidateCell; obviously both of the two branches will overlap on that one
		//in generation0.
		vendedCells[candidateCell.ref()] = true

		//See if either branch, at some generation, has the same cell forced to the same number in either generation.

		//TODO: visit the pairs of generations in such a way that the sum of the two generation counts
		//goes up linearly, since we're going to skip any that together are too long.. ... but it's
		//probably not a big deal since we'll skip early in the loop anyway.
		for firstGeneration := 0; firstGeneration < len(firstAccumulator); firstGeneration++ {
			for secondGeneration := 0; secondGeneration < len(secondAccumulator); secondGeneration++ {
				firstAffectedCells := firstAccumulator[firstGeneration].numbers
				secondAffectedCells := secondAccumulator[secondGeneration].numbers

				//We calculated up to _MAX_IMPLICATION_STEPS down each branch,
				//but we shouldn't return steps that require more than _MAX_IMPLICATION_STEPS
				//down either branch, total.
				if firstGeneration+secondGeneration > _MAX_IMPLICATION_STEPS+1 {
					continue
				}

				for key, val := range firstAffectedCells {
					//Skip the candidateCell, because that's not a meaningful overlap--we set that one as a way of branching!

					if _, ok := vendedCells[key]; ok {
						//This is a cell we've already vended
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

							if step.IsUseful(grid) {
								vendedCells[key] = true
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
		}

		//TODO: figure out why the tests are coming back with different answers, even when only looking at the key cell
		//that should work from the example.
		//TODO: figure out a way to only compute a generation if required on each branch (don't compute all the way to _MAX_IMPLICATIONS to start)

		//TODO: ideally steps with a higher generation + generation score
		//would be scored as higher diffiuclty maybe include a
		//difficultyMultiplier in SolveStep that we can fill in? Hmmm, but
		//ideally it would factor  in at humanLikelihood level. Having a
		//million different ForcingChainLength techniques would be a
		//nightmare, peformance wise... unless there was a way to pass the
		//work done in one technique to another.

	}
}

type chainSearcherGenerationDetails struct {
	numbers         map[cellRef]int
	firstGeneration map[cellRef]int
}

func (c chainSearcherGenerationDetails) String() string {
	result := "Begin map (length " + strconv.Itoa(len(c.numbers)) + ")\n"
	for cell, num := range c.numbers {
		result += "\t" + cell.String() + " : " + strconv.Itoa(num) + " : " + strconv.Itoa(c.firstGeneration[cell]) + "\n"
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

func (c chainSearcherAccumulator) addGeneration() chainSearcherAccumulator {
	newGeneration := chainSearcherGenerationDetails{make(map[cellRef]int), make(map[cellRef]int)}
	result := append(c, newGeneration)
	if len(result) > 1 {
		oldGeneration := result[len(result)-2]
		//Accumulate forward old generation
		for key, val := range oldGeneration.numbers {
			newGeneration.numbers[key] = val
			newGeneration.firstGeneration[key] = oldGeneration.firstGeneration[key]
		}
	}
	return result
}

func chainSearcher(maxGeneration int, cell *Cell, numToApply int) chainSearcherAccumulator {

	//Chainsearcher implements a BFS over implications forward given the starting point.
	//It collects its results in the provided chainSearcherAccumulator.

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

	var result chainSearcherAccumulator

	workSteps := list.New()

	//Add the first workstep.
	workSteps.PushBack(modificationToMake{
		0,
		cell,
		numToApply,
	})

	var step modificationToMake

	e := workSteps.Front()

	for e != nil {

		workSteps.Remove(e)

		switch t := e.Value.(type) {
		case modificationToMake:
			step = t
		default:
			panic("Found unexpected type in workSteps list")
		}

		if step.generation > maxGeneration {
			break
		}

		for len(result) < step.generation+1 {
			result = result.addGeneration()
		}

		generationDetails := result[step.generation]

		cellsToVisit := step.cell.Neighbors().FilterByPossible(step.numToApply).FilterByNumPossibilities(2)

		step.cell.SetNumber(step.numToApply)

		if cell.grid.Invalid() {
			//Filling that cell make the grid invalid! We found a contradiction, no need to process
			//this branch more.

			//However, this last generation--the one we found the inconsistency in--needs to be
			//thrown out.

			return result[:len(result)-1]
		}

		if currentVal, ok := generationDetails.numbers[step.cell.ref()]; ok {
			if currentVal != step.numToApply {
				//Found a contradiction! We can bail from processing any more because this branch leads inexorably
				//to a contradiction.

				//However, this last generation--the one we found the inconsistency in--needs to be thrown out.

				return result[:len(result)-1]
			}
		}

		generationDetails.numbers[step.cell.ref()] = step.numToApply
		generationDetails.firstGeneration[step.cell.ref()] = step.generation

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
				step.generation + 1,
				cellToVisit,
				forcedNum,
			})

		}

		e = workSteps.Front()

	}

	return result

}
