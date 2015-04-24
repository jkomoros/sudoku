package sudoku

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type difficultySignals map[string]float64

//A difficulty signal generator can return more than one difficutly signal, so it doesn't just return float64
//Each signal generator should always return a map with the SAME keys--so if you've called it once you know what the
//next calls will have as keys.
type difficultySignalGenerator func(directions SolveDirections) difficultySignals

const _DIFFICULTY_WEIGHT_FILENAME = "difficulties.csv"

var difficultySignalGenerators []difficultySignalGenerator
var difficultySignalWeights map[string]float64

func init() {
	difficultySignalGenerators = []difficultySignalGenerator{
		signalTechnique,
		signalNumberOfSteps,
		signalTechniquePercentage,
		signalPercentageFilledSteps,
		signalNumberUnfilled,
		signalStepsUntilNonFill,
	}

	//TODO: set reasonable DifficultySignalWeights here after we have training data we feel confident in.
	worked := false
	difficultyFile := _DIFFICULTY_WEIGHT_FILENAME

	//For now, just search upwards until we find a difficulty CSV.
	for !worked {
		worked = loadDifficultyWeights(difficultyFile)
		if worked {
			break
		}
		difficultyFile = "../" + difficultyFile
		//We're just making an ever-longer ../../../ ... FILENAME, but if we absolutized it now, we couldn't easily continue
		//prepending ../ . So just test the absFile, but still operate on difficultyFile.
		absFile, _ := filepath.Abs(difficultyFile)
		if absFile == "/"+_DIFFICULTY_WEIGHT_FILENAME {
			//We're already at the top.
			log.Println("Couldn't find a difficulty weights file.")
			break
		}
	}
}

func loadDifficultyWeights(fileName string) bool {

	//TODO: test that this loading works.

	inputFile, err := os.Open(fileName)
	if err != nil {
		//This error will be common because we'll be calling into this repeatedly in init with filenames that we don't know are valid.
		return false
	}
	defer inputFile.Close()
	csvIn := csv.NewReader(inputFile)
	records, csvErr := csvIn.ReadAll()
	if csvErr != nil {
		log.Println("The provided CSV could not be parsed.")
		return false
	}

	//Load up the weights into a map.
	difficultySignalWeights = make(map[string]float64)
	for i, record := range records {
		if len(record) != 2 {
			log.Fatalln("Record in weights csv wasn't right size: ", i, record)
		}
		theFloat, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatalln("Record in weights had an invalid float: ", i, record, err)
		}
		difficultySignalWeights[record[0]] = theFloat
	}

	return true
}

//Stats returns a printout of interesting statistics about the SolveDirections, including number of steps,
//difficulty (based on this solve description alone), how unrelated the cells in subsequent steps are,
//and the values of all of the signals used to generate the difficulty.
func (self SolveDirections) Stats() []string {
	//TODO: test this.
	techniqueCount := make(map[string]int)
	var lastStep *SolveStep
	dissimilarityAccum := 0.0
	for _, step := range self {
		if lastStep != nil {
			dissimilarityAccum += step.TargetCells.chainDissimilarity(lastStep.TargetCells)
		}
		techniqueCount[step.Technique.Name()] += 1
		lastStep = step
	}
	dissimilarityAccum /= float64(len(self))

	var result []string

	//TODO: use a standard divider across the codebase
	divider := "-------------------------"

	result = append(result, divider)
	//TODO: we shouldn't even include this... it's not meaningful to report the difficulty of a single solve.
	result = append(result, fmt.Sprintf("Difficulty : %f", self.signals().Difficulty()))
	result = append(result, divider)
	result = append(result, fmt.Sprintf("Step count: %d", len(self)))
	result = append(result, divider)
	result = append(result, fmt.Sprintf("Avg Dissimilarity: %f", dissimilarityAccum))
	result = append(result, divider)

	//We want a stable ordering for technique counts.
	for _, technique := range AllTechniques {
		//TODO: pad the technique name with enough spaces so the colon lines up.
		result = append(result, fmt.Sprintf("%s : %d", technique.Name(), techniqueCount[technique.Name()]))
	}

	result = append(result, divider)

	return result
}

//Description returns a comprehensive prose description of the SolveDirections, including reasoning for each step, that
//if followed would lead to the grid being solved. Unlike Walkthrough, Description() does not include diagrams
//for each step.
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

//Walkthrough prints an exhaustive set of human-readable directions that includes diagrams at each
//step to make it easier to follow.
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

//Because of the contract of a DifficultySignalGenerator (that it always returns the same keys), as long as DifficultySignalGenerators stays constant
//it's reasonable for callers to assume that one call to Signals() will return all of the string keys you'll see any time you call Signals()
func (self SolveDirections) signals() difficultySignals {
	result := difficultySignals{}
	for _, generator := range difficultySignalGenerators {
		result.Add(generator(self))
	}
	return result
}

//This will overwrite colliding values
//TODO: this is confusingly named
func (self difficultySignals) Add(other difficultySignals) {
	for key, val := range other {
		self[key] = val
	}
}

//For keys in both, will sum them together.
//TODO: this is confusingly named (compared to Add)
// Do we really need both Sum and Add?
func (self difficultySignals) Sum(other difficultySignals) {
	for key, val := range other {
		self[key] += val
	}
}

func (self difficultySignals) Difficulty() float64 {
	accum := 0.0

	if constant, ok := difficultySignalWeights["Constant"]; ok {
		accum = constant
	} else {
		log.Println("Didn't have the constant term loaded.")
	}

	for signal, val := range self {
		//We can discard the OK because 0 is a reasonable thing to do with weights we aren't aware of.
		weight, _ := difficultySignalWeights[signal]

		accum += val * weight
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

//Rest of file is different Signals

//This technique returns a count of how many each type of technique is seen.
//Different techniques are different "difficulties" so seeing more of a hard technique will
//Lead to a higher overall difficulty.
func signalTechnique(directions SolveDirections) difficultySignals {
	//Our contract is to always return every signal name, even if it's 0.0.
	result := difficultySignals{}
	for _, technique := range AllTechniques {
		result[technique.Name()+" Count"] = 0.0
	}
	for _, step := range directions {
		result[step.Technique.Name()+" Count"]++
	}
	return result
}

//This signal is just number of steps. More steps is PROBABLY a harder puzzle.
func signalNumberOfSteps(directions SolveDirections) difficultySignals {
	return difficultySignals{
		"Number of Steps": float64(len(directions)),
	}
}

//This signal is like signalTechnique, except it returns the count divided by the TOTAL number of steps.
func signalTechniquePercentage(directions SolveDirections) difficultySignals {
	//Our contract is to always return every signal name, even if it's 0.0.
	result := difficultySignals{}
	for _, technique := range AllTechniques {
		result[technique.Name()+" Percentage"] = 0.0
	}

	count := len(directions)

	if count == 0 {
		return result
	}

	for _, step := range directions {
		result[step.Technique.Name()+" Percentage"]++
	}

	//Now normalize all of them
	for name, _ := range result {
		result[name] /= float64(count)
	}

	return result
}

//This signal is how many steps are filled out of all steps. Presumably harder puzzles will have more non-fill steps.
func signalPercentageFilledSteps(directions SolveDirections) difficultySignals {
	numerator := 0.0
	denominator := float64(len(directions))

	for _, step := range directions {
		if step.Technique.IsFill() {
			numerator += 1.0
		}
	}

	return difficultySignals{
		"Percentage Fill Steps": numerator / denominator,
	}
}

//This signal is how many cells are unfilled at the beginning. Presumably harder puzzles will have fewer cells filled (although obviously this isn't necessarily true)
func signalNumberUnfilled(directions SolveDirections) difficultySignals {

	//We don't have access to the underlying grid, so we'll just count how many fill steps (since each can only add one number, and no numbers are ever unfilled)

	count := 0.0
	for _, step := range directions {
		if step.Technique.IsFill() {
			count++
		}
	}

	return difficultySignals{
		"Number Unfilled Cells": count,
	}
}

//This signal is how many steps into the solve directions before you encounter your first non-fill step. Non-fill steps are harder, so this signal
//captures how easy the start of the puzzle is.
func signalStepsUntilNonFill(directions SolveDirections) difficultySignals {
	count := 0.0
	for _, step := range directions {
		if !step.Technique.IsFill() {
			break
		}
		count++
	}

	return difficultySignals{
		"Steps Until Nonfill": count,
	}

}
