package sudoku

import (
	"fmt"
)

type SolveDirections []*SolveStep

const (
	ONLY_LEGAL_NUMBER = iota
)

type SolveStep struct {
	Row       int
	Col       int
	Num       int
	Technique SolveTechnique
}

type SolveTechnique interface {
	Name() string
	Description(*SolveStep) string
	Apply(*Grid) *SolveStep
}

var techniques []SolveTechnique

func init() {
	//TODO: init techniques with enough space
	techniques = append(techniques, onlyLegalNumberTechnique{})
}

type onlyLegalNumberTechnique struct {
}

func (self onlyLegalNumberTechnique) Name() string {
	return "Only Legal Number"
}

func (self onlyLegalNumberTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("%d is the only remaining valid number for that cell", step.Num)
}

func (self onlyLegalNumberTechnique) Apply(grid *Grid) *SolveStep {
	grid.refillQueue()
	//This will be a random item
	obj := grid.queue.GetSmallerThan(2)
	if obj == nil {
		//There weren't any cells with one option.
		return nil
	}
	cell := obj.(*Cell)

	cell.SetNumber(cell.implicitNumber())
	return &SolveStep{cell.Row, cell.Col, cell.Number(), self}
}

func (self *Grid) HumanSolve() *SolveDirections {
	return nil
}
