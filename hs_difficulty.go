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

type DifficultySignals map[string]float64

//A difficulty signal generator can return more than one difficutly signal, so it doesn't just return float64
//Each signal generator should always return a map with the SAME keys--so if you've called it once you know what the
//next calls will have as keys.
type DifficultySignalGenerator func(directions SolveDirections) DifficultySignals

const DIFFICULTY_WEIGHT_FILENAME = "difficulties.csv"

var DifficultySignalGenerators []DifficultySignalGenerator
var DifficultySignalWeights map[string]float64

func init() {
	DifficultySignalGenerators = []DifficultySignalGenerator{
		signalTechnique,
		signalConstant,
		signalNumberOfSteps,
	}

	//TODO: set reasonable DifficultySignalWeights here after we have training data we feel confident in.
	worked := false
	difficultyFile := DIFFICULTY_WEIGHT_FILENAME

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
		if absFile == "/"+DIFFICULTY_WEIGHT_FILENAME {
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
	DifficultySignalWeights = make(map[string]float64)
	for i, record := range records {
		if len(record) != 2 {
			log.Fatalln("Record in weights csv wasn't right size: ", i)
		}
		theFloat, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatalln("Record in weights had an invalid float: ", i)
		}
		DifficultySignalWeights[record[0]] = theFloat
	}

	return true
}

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
		return 0.0
	}

	accum := 0.0

	signals := self.Signals()

	for signal, val := range signals {
		//We can discard the OK because 0 is a reasonable thing to do with weights we aren't aware of.
		weight, _ := DifficultySignalWeights[signal]

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

//Because of the contract of a DifficultySignalGenerator (that it always returns the same keys), as long as DifficultySignalGenerators stays constant
//it's reasonable for callers to assume that one call to Signals() will return all of the string keys you'll see any time you call Signals()
func (self SolveDirections) Signals() DifficultySignals {
	result := DifficultySignals{}
	for _, generator := range DifficultySignalGenerators {
		result.Add(generator(self))
	}
	return result
}

func (self DifficultySignals) Add(other DifficultySignals) {
	for key, val := range other {
		self[key] = val
	}
}

//Rest of file is different Signals

func signalTechnique(directions SolveDirections) DifficultySignals {
	result := DifficultySignals{}
	for _, step := range directions {
		result[step.Technique.Name()+" Count"]++
	}
	return result
}

func signalConstant(directions SolveDirections) DifficultySignals {
	//Just return 1.0 for everything
	return DifficultySignals{
		"Constant": 1.0,
	}
}

func signalNumberOfSteps(directions SolveDirections) DifficultySignals {
	return DifficultySignals{
		"Number of Steps": float64(len(directions)),
	}
}
