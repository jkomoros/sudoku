package sudoku

//TODO: register this with techniques.go
type xywingTechnique struct {
	*basicSolveTechnique
}

func (self *xywingTechnique) humanLikelihood(step *SolveStep) float64 {
	//TODO: reason about what the proper value should be for this.
	return self.difficultyHelper(60.0)
}

func (self *xywingTechnique) Description(step *SolveStep) string {
	return "TODO: IMPLEMENT THIS"
}

func (self *xywingTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {

	getter := grid.queue().NewGetter()

	for {

		//Check if it's time to stop.
		select {
		case <-done:
			return
		default:
		}

		pivot := getter.GetSmallerThan(3)

		if pivot == nil {
			break
		}

		pivotCell := pivot.(*Cell)

		possibilities := pivotCell.Possibilities()

		if len(possibilities) != 2 {
			//We found one with 1 possibility, which isn't interesting for us--nakedSingle should do that one.
			continue
		}

		//OK, this is our pivot cell. Let's consider its two possibilities to
		//be X and Y.

		x := possibilities[0]
		y := possibilities[1]

		//We want to find two neighbors of the pivot cell that have two
		//possibilities that are XZ and YZ. So we're looking for neighbors
		//that have two possibilities, one of which is either x or y.

		xList := pivotCell.Neighbors().FilterByPossible(x).FilterByNumPossibilities(2)
		yList := pivotCell.Neighbors().FilterByPossible(y).FilterByNumPossibilities(2)

		//Now, we'll check for each possible value of Z
		for z := 1; z <= DIM; z++ {
			//z can't be either x or y, so don't do that work
			if z == x || z == y {
				continue
			}

			zFilteredXList := xList.FilterByPossible(z)
			zFilteredYList := yList.FilterByPossible(z)

			//If there area any items left, we've found candidates.
			if len(zFilteredXList) == 0 || len(zFilteredYList) == 0 {
				//TOOD: we can save a tiny bit of work by checking if
				//zFilteredXList is len() 0 before calculating zFilteredYList
				continue
			}

			//Okay, now the two-up combinations of all cells in the
			//two lists are all candidates. Often it will only be 1
			//cell in each list, but we'll general case this.
			for _, xCell := range zFilteredXList {
				for _, yCell := range zFilteredYList {
					//xCell, yCell is a possible pointing pair group.

					//find cells that are in both neighbor lists
					intersection := xCell.Neighbors().toCellSet().intersection(yCell.Neighbors().toCellSet()).toSlice(grid)

					//TODO: technically the filterByPossible(z) below will
					//also filter out cells that are already set, but this is
					//more semanitcally clear.
					intersection = intersection.FilterByHasPossibilities()

					//TODO: consider if we actually need to remove all of these cells;
					//it might never be able to be in the list anyway.
					//TODO: xCell, yCell can't be in there, because xCell is not in its
					//neighbors and yCell isn't in its neighbors. But it's semantically
					//clearer to do it here, maybe, to make clear the intention?
					intersection = intersection.RemoveCells(CellSlice{pivotCell, xCell, yCell})

					//TODO: consider chunking up this list of affectedCells by
					//the block the cell is in; this fits more closely with
					//how humans actually apply the technique in practice.
					//(Hmm, possibly there should be chunk and unchunked
					//variants?)
					affectedCells := intersection.FilterByPossible(z)

					if len(affectedCells) == 0 {
						continue
					}

					//Okay, we have a candidate step. Is it useful?
					step := &SolveStep{self,
						affectedCells,
						IntSlice{z},
						CellSlice{pivotCell, xCell, yCell},
						IntSlice{x, y, z},
						nil,
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
