package sudoku

import (
	"fmt"
	"math/rand"
)

type blockBlockInteractionTechnique struct {
	*basicSolveTechnique
}

func (self *blockBlockInteractionTechnique) humanLikelihood(step *SolveStep) float64 {
	//TODO: think more about how difficult this technique is.
	return self.difficultyHelper(60.0)
}

func (self *blockBlockInteractionTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {

	defer close(results)

	pairs := pairwiseBlocks(grid)

	//We're going to be looking at a lot of blocks again and again, so might as well cache this.
	unfilledCellsForBlock := make([]CellSlice, DIM)
	filledNumsForBlock := make([]IntSlice, DIM)
	for i := 0; i < DIM; i++ {
		unfilledCellsForBlock[i] = grid.Block(i).FilterByHasPossibilities()
		filledNumsForBlock[i] = grid.Block(i).FilledNums()
	}

	//For each pair of blocks (in random order)
	for _, pairIndex := range rand.Perm(len(pairs)) {

		pair := pairs[pairIndex]
		excludeNums := filledNumsForBlock[pair[0]].toIntSet().union(filledNumsForBlock[pair[1]].toIntSet())
		for _, i := range rand.Perm(DIM) {

			//See if we should stop doing work
			select {
			case <-done:
				return
			default:
			}

			//Skip numbers entirely where either of the blocks has a cell with it set, since there obviously
			//won't be any cells in both blocks that have that possibility.
			if _, ok := excludeNums[i]; ok {
				continue
			}
			//Find cells in each block that have that possibility.
			firstBlockCells := unfilledCellsForBlock[pair[0]].FilterByPossible(i)
			secondBlockCells := unfilledCellsForBlock[pair[1]].FilterByPossible(i)

			//Now we need to figure out if these blocks are in the same row or same col
			var majorAxisIsRow bool
			rowOne, colOne, _, _ := grid.blockExtents(pair[0])
			rowTwo, colTwo, _, _ := grid.blockExtents(pair[1])

			if rowOne == rowTwo {
				majorAxisIsRow = true
			} else if colOne == colTwo {
				majorAxisIsRow = false
			} else {
				panic("This shouldn't happen")
			}

			var blockOneIndexes IntSlice
			var blockTwoIndexes IntSlice

			if majorAxisIsRow {
				blockOneIndexes = firstBlockCells.CollectNums(getRow).Unique()
				blockTwoIndexes = secondBlockCells.CollectNums(getRow).Unique()
			} else {
				blockOneIndexes = firstBlockCells.CollectNums(getCol).Unique()
				blockTwoIndexes = secondBlockCells.CollectNums(getCol).Unique()
			}

			if len(blockOneIndexes) != 2 || len(blockTwoIndexes) != 2 {
				//There can only be two rows or columns in play for this technique to work.
				continue
			}

			if !blockOneIndexes.SameContentAs(blockTwoIndexes) {
				//Meh, they didn't line up into the same major axis cells.
				continue
			}

			var targetCells CellSlice

			if majorAxisIsRow {
				targetCells = grid.Row(blockOneIndexes[0])
				targetCells = append(targetCells, grid.Row(blockOneIndexes[1])...)
				targetCells = targetCells.RemoveCells(grid.Block(pair[0])).RemoveCells(grid.Block(pair[1]))
			} else {
				targetCells = grid.Col(blockOneIndexes[0])
				targetCells = append(targetCells, grid.Col(blockOneIndexes[1])...)
				targetCells = targetCells.RemoveCells(grid.Block(pair[0])).RemoveCells(grid.Block(pair[1]))
			}

			//Okay, we have a possible set. Now we need to create a step.
			step := &SolveStep{
				self,
				targetCells,
				[]int{i},
				append(firstBlockCells, secondBlockCells...),
				nil,
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

func (self *blockBlockInteractionTechnique) Description(step *SolveStep) string {
	if len(step.TargetNums) != 1 {
		return ""
	}

	blockNums := step.PointerCells.CollectNums(getBlock).Unique()
	if len(blockNums) != 2 {
		return ""
	}
	//make sure we get a stable order
	blockNums.Sort()

	grid := step.TargetCells[0].grid
	var majorAxisIsRow bool
	rowOne, colOne, _, _ := grid.blockExtents(blockNums[0])
	rowTwo, colTwo, _, _ := grid.blockExtents(blockNums[1])

	if rowOne == rowTwo {
		majorAxisIsRow = true
	} else if colOne == colTwo {
		majorAxisIsRow = false
	} else {
		panic(1)
	}

	var groupName string

	if majorAxisIsRow {
		groupName = "rows"
	} else {
		groupName = "columns"
	}

	//TODO: explain this better. It's a confusing technique, and this description could be clearer.
	return fmt.Sprintf("%d can only be in two different %s in blocks %s, which means that %d can't be in any other cells in those %s that aren't in blocks %s", step.TargetNums[0], groupName, blockNums.Description(), step.TargetNums[0], groupName, blockNums.Description())
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
