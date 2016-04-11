package sudoku

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

//DifficultySignals is a collection of names to float64 values, representing
//the various signals extracted from a SolveDirections, and used for the
//Difficulty calculation. Generally not useful to package users.
type DifficultySignals map[string]float64

//A difficulty signal generator can return more than one difficutly signal, so
//it doesn't just return float64 Each signal generator should always return a
//map with the SAME keys--so if you've called it once you know what the next
//calls will have as keys.
type difficultySignalGenerator func(directions SolveDirections) DifficultySignals

const _DIFFICULTY_WEIGHT_FILENAME = "difficulties.csv"

var difficultySignalGenerators []difficultySignalGenerator

//These are the weights that will be used to turn a list of signals into a
//difficulty. starting weights are set in hs_difficulty_weights.go, which is
//auto-generated. Generate those now:
//go:generate cmd/dokugen-analysis/internal/gendifficulties/gendifficulties
var difficultySignalWeights map[string]float64

//difficultyModelHashValue stashes the value of the hash of the model, so we
//don't have to calculate it too often.
var difficultyModelHashValue string

func init() {
	difficultySignalGenerators = []difficultySignalGenerator{
		signalTechnique,
		signalNumberOfSteps,
		signalTechniquePercentage,
		signalPercentageFilledSteps,
		signalNumberUnfilled,
		signalStepsUntilNonFill,
	}
}

//LoadDifficultyModel loads in a new difficulty model to use to score puzzles'
//difficulties. This will automatically change what DifficultyModelHash will
//return.
func LoadDifficultyModel(model map[string]float64) {
	difficultySignalWeights = model
	//Reset the stored model hash value so the next call to
	//DifficultyModelHash will recacluate it.
	difficultyModelHashValue = ""
}

//DifficultyModelHash is a unique string representing the exact difficulty
//model in use. Every time a new model is trained or loaded, this value will change.
//Therefore, if the value is different than last time you checked, the model has changed.
//This is useful for throwing out caches that assume the same difficulty model is in use.
func DifficultyModelHash() string {

	if difficultyModelHashValue == "" {

		//Generate a string with keys, vals in a known sequence, then hash.

		//first, generate a list of all keys
		var keys []string
		for k := range difficultySignalWeights {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		hash := sha1.New()

		for _, k := range keys {
			hash.Write([]byte(k + ":" + strconv.FormatFloat(difficultySignalWeights[k], 'f', -1, 64) + "\n"))
		}

		hashBytes := hash.Sum(nil)

		base32str := strings.ToUpper(hex.EncodeToString(hashBytes))

		difficultyModelHashValue = base32str
	}
	return difficultyModelHashValue
}

//Stats returns a printout of interesting statistics about the
//SolveDirections, including number of steps, difficulty (based on this solve
//description alone), how unrelated the cells in subsequent steps are, and the
//values of all of the signals used to generate the difficulty.
func (self SolveDirections) Stats() []string {
	//TODO: test this. dokugen has a method that effectively tests this; just use that.
	techniqueCount := make(map[string]int)
	var lastStep *SolveStep
	similarityAccum := 0.0
	for _, step := range self.Steps {
		if lastStep != nil {
			similarityAccum += step.TargetCells.chainSimilarity(lastStep.TargetCells)
		}
		techniqueCount[step.TechniqueVariant()] += 1
		lastStep = step
	}
	similarityAccum /= float64(len(self.Steps))

	var result []string

	//TODO: use a standard divider across the codebase
	divider := "-------------------------"

	result = append(result, divider)
	//TODO: we shouldn't even include this... it's not meaningful to report the difficulty of a single solve.
	result = append(result, fmt.Sprintf("Difficulty: %f", self.Signals().difficulty()))
	result = append(result, divider)
	result = append(result, fmt.Sprintf("Step count: %d", len(self.Steps)))
	result = append(result, divider)
	result = append(result, fmt.Sprintf("Avg Similarity: %f", similarityAccum))
	result = append(result, divider)

	//We want a stable ordering for technique counts.
	for _, technique := range AllTechniqueVariants {
		//TODO: pad the technique name with enough spaces so the colon lines up.
		result = append(result, fmt.Sprintf("%s: %d", technique, techniqueCount[technique]))
	}

	result = append(result, divider)

	return result
}

//Description returns a comprehensive prose description of the
//SolveDirections, including reasoning for each step, that if followed would
//lead to the grid being solved. Unlike Walkthrough, Description() does not
//include diagrams for each step.
func (self SolveDirections) Description() []string {

	//TODO: the IsHint directions don't sound as good as the old ones in OnlineSudoku did.

	if len(self.Steps) == 0 {
		return []string{"The puzzle is already solved."}
	}

	var descriptions []string

	//Hints have a special preamble.
	if self.IsHint {
		lastStep := self.Steps[len(self.Steps)-1]
		//TODO: this terminology is too tuned for the Online Sudoku use case.
		//it practice it should probably name the cell in text.
		descriptions = append(descriptions, "Based on the other numbers you've entered, "+lastStep.TargetCells[0].ref().String()+" can only be a "+strconv.Itoa(lastStep.TargetNums[0])+".")
		descriptions = append(descriptions, "How do we know that?")
		if len(self.Steps) > 1 {
			descriptions = append(descriptions, "We can't fill any cells right away so first we need to cull some possibilities.")
		}
	}

	for i, step := range self.Steps {
		intro := ""
		description := step.Description()
		if len(self.Steps) > 1 {
			description = strings.ToLower(description)
			switch i {
			case 0:
				intro = "First, "
			case len(self.Steps) - 1:
				intro = "Finally, "
			default:
				//TODO: switch between "then" and "next" randomly.
				intro = "Next, "
			}
		}

		descriptions = append(descriptions, intro+description)

	}
	return descriptions
}

//Walkthrough prints an exhaustive set of human-readable directions that
//includes diagrams at each step to make it easier to follow.
func (self SolveDirections) Walkthrough() string {

	//TODO: test this.

	if len(self.Steps) == 0 {
		return "The puzzle could not be solved with any of the techniques we're aware of."
	}

	clone := self.Grid()
	defer clone.Done()

	DIVIDER := "\n\n--------------------------------------------\n\n"

	intro := fmt.Sprintf("This will take %d steps to solve.", len(self.Steps))

	intro += "\nWhen you start, your grid looks like this:\n"

	intro += clone.Diagram(false)

	intro += "\n"

	intro += DIVIDER

	descriptions := self.Description()

	results := make([]string, len(self.Steps))

	for i, description := range descriptions {

		result := description + "\n"
		result += "After doing that, your grid will look like: \n\n"

		self.Steps[i].Apply(clone)

		result += clone.Diagram(false)

		results[i] = result
	}

	return intro + strings.Join(results, DIVIDER) + DIVIDER + "Now the puzzle is solved."
}

//Signals returns the DifficultySignals for this set of SolveDirections.
func (self SolveDirections) Signals() DifficultySignals {
	//Because of the contract of a DifficultySignalGenerator (that it always
	//returns the same keys), as long as DifficultySignalGenerators stays
	//constant it's reasonable for callers to assume that one call to
	//Signals() will return all of the string keys you'll see any time you
	//call Signals()
	result := DifficultySignals{}
	for _, generator := range difficultySignalGenerators {
		result.add(generator(self))
	}
	return result
}

//This will overwrite colliding values
//TODO: this is confusingly named
func (self DifficultySignals) add(other DifficultySignals) {
	for key, val := range other {
		self[key] = val
	}
}

//For keys in both, will sum them together.
//TODO: this is confusingly named (compared to Add)  Do we really need both Sum and Add?
func (self DifficultySignals) sum(other DifficultySignals) {
	for key, val := range other {
		self[key] += val
	}
}

func (self DifficultySignals) difficulty() float64 {
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

//TODO: now that SolveDirections includes gridSnapshot, think if there are any
//additional Signals we can generate.

//This technique returns a count of how many each type of technique is seen.
//Different techniques are different "difficulties" so seeing more of a hard
//technique will Lead to a higher overall difficulty.
func signalTechnique(directions SolveDirections) DifficultySignals {
	//Our contract is to always return every signal name, even if it's 0.0.
	result := DifficultySignals{}
	for _, techniqueName := range AllTechniqueVariants {
		result[techniqueName+" Count"] = 0.0
	}
	for _, step := range directions.Steps {
		result[step.TechniqueVariant()+" Count"]++
	}
	return result
}

//This signal is just number of steps. More steps is PROBABLY a harder puzzle.
func signalNumberOfSteps(directions SolveDirections) DifficultySignals {
	return DifficultySignals{
		"Number of Steps": float64(len(directions.Steps)),
	}
}

//This signal is like signalTechnique, except it returns the count divided by
//the TOTAL number of steps.
func signalTechniquePercentage(directions SolveDirections) DifficultySignals {
	//Our contract is to always return every signal name, even if it's 0.0.
	result := DifficultySignals{}
	for _, techniqueName := range AllTechniqueVariants {
		result[techniqueName+" Percentage"] = 0.0
	}

	count := len(directions.Steps)

	if count == 0 {
		return result
	}

	for _, step := range directions.Steps {
		result[step.TechniqueVariant()+" Percentage"]++
	}

	//Now normalize all of them
	for name := range result {
		result[name] /= float64(count)
	}

	return result
}

//This signal is how many steps are filled out of all steps. Presumably harder
//puzzles will have more non-fill steps.
func signalPercentageFilledSteps(directions SolveDirections) DifficultySignals {
	numerator := 0.0
	denominator := float64(len(directions.Steps))

	for _, step := range directions.Steps {
		if step.Technique.IsFill() {
			numerator += 1.0
		}
	}

	return DifficultySignals{
		"Percentage Fill Steps": numerator / denominator,
	}
}

//This signal is how many cells are unfilled at the beginning. Presumably
//harder puzzles will have fewer cells filled (although obviously this isn't
//necessarily true)
func signalNumberUnfilled(directions SolveDirections) DifficultySignals {

	//We don't have access to the underlying grid, so we'll just count how
	//many fill steps (since each can only add one number, and no numbers are
	//ever unfilled)

	count := 0.0
	for _, step := range directions.Steps {
		if step.Technique.IsFill() {
			count++
		}
	}

	return DifficultySignals{
		"Number Unfilled Cells": count,
	}
}

//This signal is how many steps into the solve directions before you encounter
//your first non-fill step. Non-fill steps are harder, so this signal captures
//how easy the start of the puzzle is.
func signalStepsUntilNonFill(directions SolveDirections) DifficultySignals {
	count := 0.0
	for _, step := range directions.Steps {
		if !step.Technique.IsFill() {
			break
		}
		count++
	}

	return DifficultySignals{
		"Steps Until Nonfill": count,
	}

}
