//gendifficulties is used to take weka output and create the
//hs_difficulty_weights file in the main package with go generate.
package main

import (
	"github.com/jkomoros/sudoku/cmd/dokugen-analysis/internal/wekaparser"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
)

var BASE_SUDOKU_DIR string
var BASE_DIR string
var INPUT_FILE_NAME string
var INPUT_SAMPLE_FILE_NAME string
var WEIGHTS_FILE_NAME string
var OUTPUT_FILE_NAME string

func init() {
	BASE_SUDOKU_DIR = os.ExpandEnv("$GOPATH/src/github.com/jkomoros/sudoku/")
	BASE_DIR = BASE_SUDOKU_DIR + "cmd/dokugen-analysis/internal/gendifficulties/"
	INPUT_FILE_NAME = BASE_DIR + "input.txt"
	INPUT_SAMPLE_FILE_NAME = BASE_DIR + "input.SAMPLE.txt"
	WEIGHTS_FILE_NAME = "hs_difficulty_weights.go"
	OUTPUT_FILE_NAME = BASE_SUDOKU_DIR + WEIGHTS_FILE_NAME
}

func main() {

	input, err := ioutil.ReadFile(INPUT_FILE_NAME)

	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalln("Got err trying to open first file:", err)
		}
		input, err = ioutil.ReadFile(INPUT_SAMPLE_FILE_NAME)
		if err != nil {
			log.Fatalln("Couldn't load either of the input files.")
		}
	}

	weights, err := wekaparser.ParseWeights(string(input))

	if err != nil {
		log.Fatalln("Couldn't parse weights:", err)
	}

	r2, err := wekaparser.ParseR2(string(input))

	if err != nil {
		log.Println("Couldn't extract r2:", err)
	}

	var output string

	output += "package sudoku\n\n"
	output += "//auto-generated by difficulty-convert.py DO NOT EDIT\n\n"

	output += "func init() {\n"

	output += "\t//Model with R2 = " + strconv.FormatFloat(r2, 'f', -1, 64) + "\n"

	output += "\tLoadDifficultyModel(map[string]float64{\n"

	var keys []string
	for key, _ := range weights {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		output += "\t\t\"" + key + "\" : " + strconv.FormatFloat(weights[key], 'f', -1, 64) + ",\n"
	}

	output += "\t})\n"
	output += "}\n"

	err = ioutil.WriteFile(OUTPUT_FILE_NAME, []byte(output), 0644)

	if err != nil {
		log.Fatalln("Writing file didn't work:", err)
	}

	//Gofmt the new output

	cmd := exec.Command("go", "fmt", OUTPUT_FILE_NAME)

	goFmtOutput, err := cmd.Output()

	if err != nil {
		log.Println("Running go format failed. You should run go fmt yourself on", WEIGHTS_FILE_NAME)
		return
	}

	if string(goFmtOutput) != WEIGHTS_FILE_NAME+"\n" {
		log.Println("Go fmt did not give expected output:", string(goFmtOutput))
	}
}
