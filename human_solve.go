package dokugen

type SolveDirections []*SolveStep

type SolveTechnique int

const (
	ONLY_LEGAL_NUMBER = iota
	ONLY_LEGAL_POS_ROW
	ONLY_LEGAL_POS_COL
	ONLY_LEGAL_POS_BLOCK
)

type SolveStep struct {
	Row       int
	Col       int
	Num       int
	Technique SolveTechnique
}

func (self *Grid) HumanSolve() *SolveDirections {
	return nil
}
