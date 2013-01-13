package dokugen

type SolveDirections []*SolveStep

type SolveStep struct {
	Row    int
	Col    int
	Num    int
	Reason string
}

func (self *Grid) HumanSolve() *SolveDirections {
	return nil
}
