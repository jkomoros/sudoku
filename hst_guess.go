package sudoku

import (
	"fmt"
	"math/rand"
)

type guessTechnique struct {
	*basicSolveTechnique
}

func (self *guessTechnique) HumanLikelihood() float64 {
	return self.difficultyHelper(100000000.0)
}

func (self *guessTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("we have no other moves to make, so we randomly pick a cell with the smallest number of possibilities, %s, and pick one of its possibilities", step.TargetCells.Description())
}

func (self *guessTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {

	getter := grid.queue().NewGetter()

	for {

		select {
		case <-done:
			return
		default:
		}

		obj := getter.Get()
		if obj == nil {
			break
		}

		//This WILL happen, since guess will return a bunch of possible guesses you could make.
		if obj.rank() > 3 {
			//Given that this WILL happen, it's important to return results so far, whatever they are.
			break
		}

		//Convert RankedObject to a cell
		cell := obj.(*Cell)
		possibilities := cell.Possibilities()

		if len(possibilities) == 0 {
			//Not entirely sure why this would happen, but it can...
			continue
		}

		num := possibilities[rand.Intn(len(possibilities))]
		step := newFillSolveStep(cell, num, self)

		//We're going to abuse pointerNums and use it to point out the other numbers we COULD have used.
		step.PointerNums = IntSlice(possibilities).Difference(IntSlice{num})
		if step.IsUseful(grid) {
			select {
			case results <- step:
			case <-done:
				return
			}
		}
	}
}
