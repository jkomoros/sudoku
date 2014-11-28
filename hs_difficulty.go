package sudoku

import (
	"fmt"
	"log"
	"strings"
)

func (self SolveDirections) Stats() []string {
	//TODO: test this.
	techniqueCount := make(map[string]int)
	for _, step := range self {
		techniqueCount[step.Technique.Name()] += 1
	}
	var result []string

	//TODO: use a standard divider across the codebase
	divider := "-------------------------"

	result = append(result, divider)
	result = append(result, fmt.Sprintf("Difficulty : %f", self.Difficulty()))
	result = append(result, divider)
	result = append(result, fmt.Sprintf("Step count: %d", len(self)))
	result = append(result, divider)

	//We want a stable ordering for technique counts.
	for _, technique := range AllTechniques {
		result = append(result, fmt.Sprintf("%s : %d", technique.Name(), techniqueCount[technique.Name()]))
	}

	result = append(result, divider)

	return result
}

func (self SolveDirections) Description() []string {

	if len(self) == 0 {
		return []string{""}
	}

	descriptions := make([]string, len(self))

	for i, step := range self {
		intro := ""
		switch i {
		case 0:
			intro = "First, "
		case len(self) - 1:
			intro = "Finally, "
		default:
			//TODO: switch between "then" and "next" randomly.
			intro = "Next, "
		}
		descriptions[i] = intro + strings.ToLower(step.Description())

	}
	return descriptions
}

func (self SolveDirections) Walkthrough(grid *Grid) string {

	//TODO: test this.

	if len(self) == 0 {
		return "The puzzle could not be solved with any of the techniques we're aware of."
	}

	clone := grid.Copy()
	defer clone.Done()

	DIVIDER := "\n\n--------------------------------------------\n\n"

	intro := fmt.Sprintf("This will take %d steps to solve.", len(self))

	intro += "\nWhen you start, your grid looks like this:\n"

	intro += clone.Diagram()

	intro += "\n"

	intro += DIVIDER

	descriptions := self.Description()

	results := make([]string, len(self))

	for i, description := range descriptions {

		result := description + "\n"
		result += "After doing that, your grid will look like: \n\n"

		self[i].Apply(clone)

		result += clone.Diagram()

		results[i] = result
	}

	return intro + strings.Join(results, DIVIDER) + DIVIDER + "Now the puzzle is solved."
}

func (self SolveDirections) Difficulty() float64 {
	//How difficult the solve directions described are. The measure of difficulty we use is
	//just summing up weights we see; this captures:
	//* Number of steps
	//* Average difficulty of steps
	//* Number of hard steps
	//* (kind of) the hardest step: because the difficulties go up expontentionally.

	//This method assumes the weights have been calibrated empirically to give scores between 0.0 and 1.0
	//without normalization here.

	if len(self) == 0 {
		//The puzzle was not able to be solved, apparently.
		return 1.0
	}

	accum := difficultyConstant
	for _, step := range self {
		accum += step.Technique.Difficulty()
	}

	if accum < 0.0 {
		log.Println("Accumuldated difficulty snapped to 0.0:", accum)
		accum = 0.0
	}

	if accum > 1.0 {
		log.Println("Accumulated difficulty snapped to 1.0:", accum)
		accum = 1.0
	}

	return accum
}
