//wekaparser takes the output of Weka's training model and parses it into a
//map of weights and r2.
package wekaparser

import (
	"errors"
	"strconv"
	"strings"
)

//TODO: create a tool that takes in the input and outputs hs_difficulties.go

//ParseWeights takes the output of the weka-trainer and returns a map of the weights.
func ParseWeights(input string) (weights map[string]float64, err error) {

	weights = make(map[string]float64)

	for i, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0] != '+' && line[0] != '-' {
			//All lines we're interested in start with either + or -
			continue
		}
		negative := line[0] == '-'
		line = strings.TrimSpace(line[1:])
		parts := strings.Split(line, " * ")
		if len(parts) > 2 {
			return nil, errors.New("Skipped line " + strconv.Itoa(i) + " because it was not shaped the way we expected.")
		}
		var name string
		if len(parts) == 1 {
			name = "Constant"
		} else {
			name = strings.TrimSpace(parts[1])
		}

		if negative {
			parts[0] = "-" + parts[0]
		}

		parts[0] = strings.TrimSpace(parts[0])

		flt, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return nil, errors.New("Skipping line " + strconv.Itoa(i) + " because float could not be parsed:" + err.Error())
		}

		weights[name] = flt
	}

	return weights, nil

}

//ParseR2 takes the output of weka-trainer and returns the R2
func ParseR2(input string) (r2 float64, err error) {

	inCrossValidationSection := false

	for _, line := range strings.Split(input, "\n") {
		if strings.Contains(line, "Cross-validation") {
			inCrossValidationSection = true
			continue
		}
		if !inCrossValidationSection {
			continue
		}
		if !strings.HasPrefix(line, "Correlation coefficient") {
			continue
		}

		line = strings.Replace(line, "Correlation coefficient", "", -1)
		line = strings.TrimSpace(line)

		flt, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return 0.0, errors.New("Couldn't parse r2 when found:" + err.Error())
		}
		return flt, nil
	}
	return 0.0, errors.New("Couldn't find r2 in input")

}
