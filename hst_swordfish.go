package sudoku

type swordfishTechnique struct {
	*basicSolveTechnique
}

func (self *swordfishTechnique) humanLikelihood(step *SolveStep) float64 {
	//TODO: reason more carefully about how hard this technique is.
	return self.difficultyHelper(70.0)
}

func (self *swordfishTechnique) Description(step *SolveStep) string {
	//TODO: Implement this
	return "TODO: IMPLEMENT ME"
}

func (self *swordfishTechnique) Find(grid *Grid, results chan *SolveStep, done chan bool) {

	//TODO: implement me

}
