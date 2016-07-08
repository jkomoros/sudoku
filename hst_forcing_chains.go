package sudoku

import (
	"fmt"
	"strconv"
)

//TODO: investigate bumping this back up when #100 lands
const _MAX_IMPLICATION_STEPS = 5

type forcingChainsTechnique struct {
	*basicSolveTechnique
}

func (self *forcingChainsTechnique) numImplicationSteps(step *SolveStep) int {

	if step == nil {
		return 1
	}

	//Verify that the information we're unpacking is what we expect
	numImplicationSteps, ok := step.extra.(int)

	if !ok {
		numImplicationSteps = 1
	}
	return numImplicationSteps
}

func (self *forcingChainsTechnique) Variants() []string {
	var result []string
	for i := 1; i <= _MAX_IMPLICATION_STEPS+1; i++ {
		result = append(result, self.Name()+" ("+strconv.Itoa(i)+" steps)")
	}
	return result
}

func (self *forcingChainsTechnique) variant(step *SolveStep) string {
	return self.basicSolveTechnique.variant(step) + " (" + strconv.Itoa(self.numImplicationSteps(step)) + " steps)"
}

func (self *forcingChainsTechnique) humanLikelihood(step *SolveStep) float64 {
	//TODO: figure out what the baseDifficulty should be, this might be higher than
	//it's actually in practice

	//Note that this number has to be pretty high because it's competing against
	//HiddenSIZEGROUP, which has the k exponential in its favor.
	return float64(self.numImplicationSteps(step)) * self.difficultyHelper(20000.0)
}

func (self *forcingChainsTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("cell %s only has two options, %s, and if you put either one in and see the chain of implications it leads to, both ones end up with %s in cell %s, so we can just fill that number in", step.PointerCells.Description(), step.PointerNums.Description(), step.TargetNums.Description(), step.TargetCells.Description())
}

func (self *forcingChainsTechnique) Candidates(grid *Grid, maxResults int) []*SolveStep {
	return self.candidatesHelper(self, grid, maxResults)
}

func (self *forcingChainsTechnique) find(grid *Grid, coordinator findCoordinator) {
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

	getter := grid.queue().NewGetter()

	for {

		//Check if it's time to stop.
		if coordinator.shouldExitEarly() {
			return
		}

		candidate := getter.GetSmallerThan(3)

		if candidate == nil {
			break
		}

		candidateCell := candidate.(Cell)

		if len(candidateCell.Possibilities()) != 2 {
			//We found one with 1 possibility, which isn't interesting for us--nakedSingle should do that one.
			continue
		}

		firstPossibilityNum := candidateCell.Possibilities()[0]
		secondPossibilityNum := candidateCell.Possibilities()[1]

		firstGrid := grid.Copy()
		secondGrid := grid.Copy()

		//Check that the neighbor isn't just already having a single possibility, because then this technique is overkill.

		firstAccumulator := &chainSearcherAccumulator{make(map[cellRef]IntSlice), make(map[cellRef]IntSlice)}
		secondAccumulator := &chainSearcherAccumulator{make(map[cellRef]IntSlice), make(map[cellRef]IntSlice)}

		chainSearcher(0, _MAX_IMPLICATION_STEPS,
			candidateCell.InGrid(firstGrid).Mutable(),
			firstPossibilityNum, firstAccumulator)

		chainSearcher(0, _MAX_IMPLICATION_STEPS,
			candidateCell.InGrid(secondGrid).Mutable(),
			secondPossibilityNum, secondAccumulator)

		firstAccumulator.reduce()
		secondAccumulator.reduce()

		//See if either branch, at some generation, has the same cell forced to the same number in either generation.

		//We're just going to look at the last generation for each and compare
		//when each cell was setœ instead of doing (expensive!) pairwise
		//comparison across all of them

		for cell, numSlice := range firstAccumulator.numbers {
			if secondNumSlice, ok := secondAccumulator.numbers[cell]; ok {

				//Found two cells that overlap in terms of both being affected.
				//We're only interested in them if they are both set to exactly one item, which is the
				//same number.
				if len(numSlice) != 1 || len(secondNumSlice) != 1 || !numSlice.SameContentAs(secondNumSlice) {
					continue
				}

				numImplicationSteps := firstAccumulator.firstGeneration[cell][0] + secondAccumulator.firstGeneration[cell][0]

				//Is their combined generation count lower than _MAX_IMPLICATION_STEPS?
				if numImplicationSteps > _MAX_IMPLICATION_STEPS+1 {
					//Too many implication steps. :-(
					continue
				}

				//Okay, we have a candidate step. Is it useful?
				step := &SolveStep{self,
					CellSlice{cell.Cell(grid)},
					IntSlice{numSlice[0]},
					CellSlice{candidateCell},
					candidateCell.Possibilities(),
					numImplicationSteps,
				}

				if step.IsUseful(grid) {
					if coordinator.foundResult(step) {
						return
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

type chainSearcherAccumulator struct {
	numbers         map[cellRef]IntSlice
	firstGeneration map[cellRef]IntSlice
}

func (c *chainSearcherAccumulator) String() string {
	result := "Begin map (length " + strconv.Itoa(len(c.numbers)) + ")\n"
	for cell, numSlice := range c.numbers {
		result += "\t" + cell.String() + " : " + numSlice.Description() + " : " + c.firstGeneration[cell].Description() + "\n"
	}
	result += "End map\n"
	return result
}

//Goes through each item in the map and removes duplicates, keeping the smallest generation seen for each unique number.
func (c *chainSearcherAccumulator) reduce() {
	for cell, numList := range c.numbers {
		output := make(map[int]int)
		generationList, ok := c.firstGeneration[cell]
		if !ok {
			panic("numbers and firstGeneration were out of sync")
		}
		for i, num := range numList {
			generation := generationList[i]
			currentGeneration := output[num]
			if currentGeneration == 0 || generation < currentGeneration {
				output[num] = generation
			}
		}
		var resultNumbers IntSlice
		var resultGenerations IntSlice
		for num, generation := range output {
			resultNumbers = append(resultNumbers, num)
			resultGenerations = append(resultGenerations, generation)
		}
		c.numbers[cell] = resultNumbers
		c.firstGeneration[cell] = resultGenerations
	}
}

func chainSearcher(generation int, maxGeneration int, cell MutableCell, numToApply int, accum *chainSearcherAccumulator) {

	/*
	 * chainSearcher implements a DFS to search forward through implication chains to
	 * fill out accum with details about cells it sees and sets.
	 * The reason a DFS and not a BFS is called for is because with forcing chains, we
	 * KNOW that either the left or right branch will lead to an inconsistency at some point
	 * (as long as the sudoku has only one valid solution). We want to IGNORE that
	 * inconsistency for as long as possible to follow the implication chains as deep as we can go.
	 * By definition, the end of the DFS will be the farthest a given implication chain can go
	 * towards setting that specific cell to the forced value. This means that we have the maximum
	 * density of implication chain results to sift through to find cells forced to the same value.
	 */

	if generation > maxGeneration {
		//base case
		return
	}

	//Becuase this is a DFS, if we see an invalidity in this grid, it's a meaningful invalidity
	//and we should avoid it.
	if cell.grid().Invalid() {
		return
	}

	cellsToVisit := cell.Neighbors().FilterByPossible(numToApply).FilterByNumPossibilities(2)

	cell.SetNumber(numToApply)

	//Accumulate information about this cell being set. We'll reduce out duplicates later.
	accum.numbers[cell.ref()] = append(accum.numbers[cell.ref()], numToApply)
	accum.firstGeneration[cell.ref()] = append(accum.firstGeneration[cell.ref()], generation)

	for _, cellToVisit := range cellsToVisit {
		possibilities := cellToVisit.Possibilities()

		if len(possibilities) != 1 {
			panic("Expected the cell to have one possibility")
		}

		forcedNum := possibilities[0]

		//recurse
		chainSearcher(generation+1, maxGeneration, cellToVisit.Mutable(), forcedNum, accum)
	}

	//Undo this number and return
	cell.SetNumber(0)
}
