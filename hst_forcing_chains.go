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

		firstAffectedCellSet := cellSet{}
		secondAffectedCellSet := cellSet{}

		//What numbers were put into the cells on the two branches
		//(We can't use the cellSet because that only store bools)
		firstAffectedCellMapping := make(map[*Cell]int)
		secondAffectedCellMapping := make(map[*Cell]int)

		firstCurrentCell := candidateCell.InGrid(firstGrid)
		secondCurrentCell := candidateCell.InGrid(secondGrid)

		for i := 0; i < _MAX_IMPLICATION_STEPS; i++ {

			//Don't do work if this branch is done
			if firstCurrentCell != nil {

				//Before we make the change, make sure the change will force
				//another cell. (We don't just set it and see if there are any
				//singleLegals in neigbhors because this technique is only
				//concerned with CHAINS of cells that we bring into existence.

				//Find the nextCells that will have their numbers forced by the cell we're thinking of fillint.
				nextCells := firstCurrentCell.Neighbors().FilterByPossible(firstPossibilityNum).FilterByNumPossibilities(2)

				if len(nextCells) == 1 {

					cell := nextCells[0]
					//TODO: really if more than one cell is affected that's fine, we should recursively handle all of those cells.
					//(Although of course if no cells are affected then we should abandon this chain)

					//Okay, looks like we found a cell we're going to affect.
					//Note that we're going to affect it.

					firstAffectedCellSet = firstAffectedCellSet.union(cellSet{cell: true})

					forcedNum := cell.Possibilities().Difference(IntSlice{firstPossibilityNum})[0]
					firstAffectedCellMapping[cell] = forcedNum

					//Make the change to the original set
					firstCurrentCell.SetNumber(firstPossibilityNum)

					//And update the one we're going to do next step
					firstPossibilityNum = forcedNum
					firstCurrentCell = cell

				} else {
					//Stop searching forward on this chain
					firstCurrentCell = nil
				}
			}
			if secondCurrentCell != nil {

				//Before we make the change, make sure the change will force
				//another cell. (We don't just set it and see if there are any
				//singleLegals in neigbhors because this technique is only
				//concerned with CHAINS of cells that we bring into existence.

				//Find the nextCells that will have their numbers forced by the cell we're thinking of fillint.
				nextCells := secondCurrentCell.Neighbors().FilterByPossible(secondPossibilityNum).FilterByNumPossibilities(2)

				if len(nextCells) == 1 {

					cell := nextCells[0]
					//TODO: really if more than one cell is affected that's fine, we should recursively handle all of those cells.
					//(Although of course if no cells are affected then we should abandon this chain)

					//Okay, looks like we found a cell we're going to affect.
					//Note that we're going to affect it.

					secondAffectedCellSet = secondAffectedCellSet.union(cellSet{cell: true})

					forcedNum := cell.Possibilities().Difference(IntSlice{secondPossibilityNum})[0]
					secondAffectedCellMapping[cell] = forcedNum

					//Make the change to the original set
					secondCurrentCell.SetNumber(secondPossibilityNum)

					//And update the one we're going to do next step
					secondPossibilityNum = forcedNum
					secondCurrentCell = cell

				} else {
					//Stop searching forward on this chain
					secondCurrentCell = nil
				}
			}

			//Check if the sets overlap.

			intersection := firstAffectedCellSet.intersection(secondAffectedCellSet)
			if len(intersection) > 0 {
				//Okay, a cell overlapped... did they both set the same number?
				//TODO: should we look at all items that overlap if it's greater than 1?

				cell := intersection.toSlice()[0]

				if firstAffectedCellMapping[cell] == secondAffectedCellMapping[cell] {
					//Booyah, found a step.

					step := &SolveStep{self,
						CellSlice{cell},
						IntSlice{firstAffectedCellMapping[cell]},
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
