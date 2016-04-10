//wekaparser takes the output of Weka's training model and parses it into a
//map of weights and r2.
package wekaparser

import (
	"errors"
	"strconv"
	"strings"
)

//TODO: implement ParseR2
//TODO: use ParseR2 in analysis-pipeline
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
