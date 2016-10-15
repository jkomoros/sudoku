package sudoku

import (
	"fmt"
	"math"
)

type guessTechnique struct {
	*basicSolveTechnique
}

func (self *guessTechnique) humanLikelihood(step *SolveStep) float64 {
	//Guess is so horribly terrible that we want it to NEVER come in lower
	//than even a chain of cullSteps. Making it postive infinity means that no
	//twiddles will ever make it less impossibly bad.
	return self.difficultyHelper(math.Inf(1))
}

func (self *guessTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("we have no other moves to make, so we randomly pick a cell with the smallest number of possibilities, %s, and pick one of its possibilities", step.TargetCells.Description())
}

func (self *guessTechnique) Candidates(grid Grid, maxResults int) []*SolveStep {
	return self.candidatesHelper(self, grid, maxResults)
}

func (self *guessTechnique) find(grid Grid, coordinator findCoordinator) {

	//We used to have a very elaborate aparatus for guess logic where we'd
	//earnestly guess and then HumanSolve forward until we discovered a
	//solution or an inconsistency. But that was dumb because we never
	//returned a result down a wrong guess (or otherwise proved that we had
	//done the 'real' work). And it massively complicated the flow. So... just
	//cheat. Brute force solve the grid, pick a cell with small number of
	//possibilities, and then just immediately return the correct value for
	//it. Done!

	solvedGrid := grid.MutableCopy()
	solvedGrid.Solve()

	if !solvedGrid.Solved() {
		return
	}

	getter := grid.queue().NewGetter()

	for {

		if coordinator.shouldExitEarly() {
			return
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
		cell := obj.(Cell)

		cellInSolvedGrid := cell.InGrid(solvedGrid)

		num := cellInSolvedGrid.Number()
		step := &SolveStep{
			Technique:   self,
			TargetCells: CellRefSlice{cell.Reference()},
			TargetNums:  IntSlice{num},
		}

		if step.IsUseful(grid) {
			if coordinator.foundResult(step) {
				return
			}
		}
	}
}
