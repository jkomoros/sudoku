package sudoku

import (
	"fmt"
	"math/rand"
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
	techniques = append(techniques, necessaryInRowTechnique{})
}

type onlyLegalNumberTechnique struct {
}

type necessaryInRowTechnique struct {
}

func (self onlyLegalNumberTechnique) Name() string {
	return "Only Legal Number"
}

func (self onlyLegalNumberTechnique) Description(step *SolveStep) string {
	return fmt.Sprintf("%d is the only remaining valid number for that cell", step.Num)
}

func (self onlyLegalNumberTechnique) Apply(grid *Grid) *SolveStep {
	//This will be a random item
	obj := grid.queue.NewGetter().GetSmallerThan(2)
	if obj == nil {
		//There weren't any cells with one option.
		return nil
	}
	cell := obj.(*Cell)

	cell.SetNumber(cell.implicitNumber())
	return &SolveStep{cell.Row, cell.Col, cell.Number(), self}
}

func (self necessaryInRowTechnique) Name() string {
	return "Necessary In Row"
}

func (self necessaryInRowTechnique) Description(step *SolveStep) string {
	//TODO: format the text to say "first/second/third/etc"
	return fmt.Sprintf("%d is required in the %d row, and %d is the only column it fits", step.Num, step.Row+1, step.Col+1)
}

func (self necessaryInRowTechnique) Apply(grid *Grid) *SolveStep {
	//This will be a random item
	indexes := rand.Perm(DIM)

	for _, r := range indexes {
		seenInRow := make([]int, DIM)
		row := grid.Row(r)
		for _, cell := range row {
			for _, possibility := range cell.Possibilities() {
				seenInRow[possibility-1]++
			}
		}
		seenIndexes := rand.Perm(DIM)
		for _, index := range seenIndexes {
			seen := seenInRow[index]
			if seen == 1 {
				//Okay, we know our target number. Which cell was it?
				for _, cell := range row {
					if cell.Possible(index + 1) {
						//Found it!
						cell.SetNumber(index + 1)
						return &SolveStep{cell.Row, cell.Col, cell.Number(), self}
					}
				}
			}
		}
	}
	//Nope.
	return nil
}

func (self *Grid) HumanSolve() *SolveDirections {
	return nil
}
