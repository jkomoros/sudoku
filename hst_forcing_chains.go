package sudoku

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

		firstAffectedCellSets, firstAffectedCellNums := chainSearcher(_MAX_IMPLICATION_STEPS,
			candidateCell.InGrid(firstGrid),
			firstPossibilityNum,
			[]cellSet{cellSet{}},
			[]map[*Cell]int{make(map[*Cell]int)})

		secondAffectedCellSets, secondAffectedCellNums := chainSearcher(_MAX_IMPLICATION_STEPS,
			candidateCell.InGrid(secondGrid),
			secondPossibilityNum,
			[]cellSet{cellSet{}},
			[]map[*Cell]int{make(map[*Cell]int)})

		//Check if the sets overlap.

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

	}
}

func chainSearcher(i int, cell *Cell, numToApply int, affectedCellSetsSoFar []cellSet, affectedCellNumsSoFar []map[*Cell]int) (affectedCellSets []cellSet, affectedCellNums []map[*Cell]int) {
	if i <= 0 || cell == nil {
		//Base case
		return affectedCellSetsSoFar, affectedCellNumsSoFar
	}

	if len(affectedCellSetsSoFar) != len(affectedCellNumsSoFar) {
		panic("Mismatched sizes of arrays to chainSearcher")
	}

	var resultAffectedCellSets []cellSet
	var resultAffectedCellNums []map[*Cell]int

	for k, affectedCellSet := range affectedCellSetsSoFar {

		//Before we make the change, make sure the change will force
		//another cell. (We don't just set it and see if there are any
		//singleLegals in neigbhors because this technique is only
		//concerned with CHAINS of cells that we bring into existence.

		affectedCellNums := affectedCellNumsSoFar[k]

		//Find the nextCells that will have their numbers forced by the cell we're thinking of fillint.
		for j, nextCell := range cell.Neighbors().FilterByPossible(numToApply).FilterByNumPossibilities(2) {
			forcedNum := nextCell.Possibilities().Difference(IntSlice{numToApply})[0]

			var newAffectedNums map[*Cell]int

			//If it's not the first one, we're effectively branching.
			if j == 0 {
				newAffectedNums = affectedCellNums

			} else {
				newGrid := nextCell.grid.Copy()
				nextCell = nextCell.InGrid(newGrid)

				newAffectedNums = make(map[*Cell]int)

				for key, val := range affectedCellNums {
					newAffectedNums[key] = val
				}

			}

			newAffectedNums[nextCell] = forcedNum

			nextCell.SetNumber(forcedNum)

			theAffectedCellSets, theAffectedCellNums := chainSearcher(i-1,
				nextCell,
				forcedNum,
				[]cellSet{affectedCellSet.union(cellSet{nextCell: true})},
				[]map[*Cell]int{newAffectedNums})

			resultAffectedCellSets = append(resultAffectedCellSets, theAffectedCellSets...)
			resultAffectedCellNums = append(resultAffectedCellNums, theAffectedCellNums...)
		}
	}
	return resultAffectedCellSets, resultAffectedCellNums

}
