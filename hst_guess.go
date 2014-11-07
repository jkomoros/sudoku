package sudoku

import (
	"fmt"
	"math/rand"
)

type guessTechnique struct {
	*basicSolveTechnique
}

func (self *guessTechnique) Difficulty() float64 {
	return self.difficultyHelper(1000.0)
}

func (self *guessTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("we have no other moves to make, so we randomly pick a cell with the smallest number of possibilities, %s, and pick one of its possibilities", step.TargetCells.Description())
}

func (self *guessTechnique) Find(grid *Grid) []*SolveStep {

	getter := grid.queue.NewGetter()

	var results []*SolveStep

	for {
		obj := getter.Get()
		if obj == nil {
			break
		}
		//Convert RankedObject to a cell
		cell := obj.(*Cell)
		possibilities := cell.Possibilities()
		num := possibilities[rand.Intn(len(possibilities))]
		step := newFillSolveStep(cell, num, self)

		//We're going to abuse pointerNums and use it to point out the other numbers we COULD have used.
		step.PointerNums = IntSlice(possibilities).Difference(IntSlice{num})
		if step.IsUseful(grid) {
			results = append(results, step)
		}
	}

	return results
}
