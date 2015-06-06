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

	//TODO: implement this.
	return
}
