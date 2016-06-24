package sudoku

import (
	"fmt"
)

type guessTechnique struct {
	*basicSolveTechnique
}

func (self *guessTechnique) humanLikelihood(step *SolveStep) float64 {
	return self.difficultyHelper(1000000000000000000000.0)
}

func (self *guessTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("we have no other moves to make, so we randomly pick a cell with the smallest number of possibilities, %s, and pick one of its possibilities", step.TargetCells.Description())
}

func (self *guessTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {

	//We used to have a very elaborate aparatus for guess logic where we'd
	//earnestly guess and then HumanSolve forward until we discovered a
	//solution or an inconsistency. But that was dumb because we never
	//returned a result down a wrong guess (or otherwise proved that we had
	//done the 'real' work). And it massively complicated the flow. So... just
	//cheat. Brute force solve the grid, pick a cell with small number of
	//possibilities, and then just immediately return the correct value for
	//it. Done!

	solvedGrid := grid.Copy()
	solvedGrid.Solve()

	if !solvedGrid.Solved() {
		return
	}

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

		cellInSolvedGrid := cell.InGrid(solvedGrid)

		num := cellInSolvedGrid.Number()
		step := newFillSolveStep(cell, num, self)

		if step.IsUseful(grid) {
			select {
			case results <- step:
			case <-done:
				return
			}
		}
	}
}
