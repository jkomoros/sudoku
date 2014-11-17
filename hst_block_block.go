package sudoku

type blockBlockInteractionTechnique struct {
	*basicSolveTechnique
}

func (self *blockBlockInteractionTechnique) Difficulty() float64 {
	//TODO: think more about how difficult this technique is.
	return self.difficultyHelper(2.5)
}

func (self *blockBlockInteractionTechnique) Find(grid *Grid) []*SolveStep {
	//TODO implement and test this.
	return nil
}

func (self *blockBlockInteractionTechnique) Description(step *SolveStep) string {
	//TODO: implement and test this.
	return ""
}

//Technically in the future different grids could have different blcok partioning schemes
func pairwiseBlocks(grid *Grid) [][]int {
	//Returns a list of pairs of block IDs, where the blocks are in either the same row or column.

	//TODO: implement this in a way that doesn't generate all of the pairs and cull.
	var result [][]int

	for _, pair := range subsetIndexes(DIM, 2) {
		if len(pair) != 2 {
			return nil
		}
		//Get the upper left cell cooordinates for both blocks.
		rowOne, colOne, _, _ := grid.blockExtents(pair[0])
		rowTwo, colTwo, _, _ := grid.blockExtents(pair[1])
		//Then see if thw coords are the same.
		if rowOne == rowTwo || colOne == colTwo {
			//Found one that's the same, keep it
			result = append(result, pair)

			//Note: we don't have to worry about getting the same blocks back for the front and back of the pair, thanks
			//to how subsetIndexes works.
		}
	}
	return result

}
