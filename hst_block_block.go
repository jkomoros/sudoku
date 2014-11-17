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
