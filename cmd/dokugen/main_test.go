package main

import (
	"bytes"
	"flag"
	"github.com/jkomoros/sudoku"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

//regexp101.com and regexr.com are good tools for creating the regular expressions.
const GRID_RE = `((\d\||\.\|){8}(\d|\.)\n){8}((\d\||\.\|){8}(\d|\.))\n?`
const OUTPUT_DIVIDER_RE = `-{25}\n?`
const FLOAT_RE = `\d{1,5}\.\d{4,20}`
const INT_RE = `\b\d{1,5}\b`

var VARIANT_RE string

func init() {
	variantsPortion := strings.Join(sudoku.AllTechniqueVariants, "|")
	variantsPortion = strings.Replace(variantsPortion, "(", "\\(", -1)
	variantsPortion = strings.Replace(variantsPortion, ")", "\\)", -1)
	VARIANT_RE = "(" + variantsPortion + ")"
}

func numLineRE(word string, isFloat bool) string {
	numPortion := INT_RE
	if isFloat {
		numPortion = FLOAT_RE
	}
	return word + `:\s` + numPortion + `\n?`

}

func TestHelp(t *testing.T) {

	options := getDefaultOptions()

	options.HELP = true

	options.fixUp()
	output, errOutput := getOutput(options)
	expectations := getExpectations("help")

	if output != "" {
		t.Error("For help message, expected empty stdout, got", output)
	}

	if errOutput != expectations {
		t.Error("For help message, got\n", errOutput, "\nwanted\n", expectations)
	}
}

func regularExpressionMatch(reText string, input string) bool {
	re := regexp.MustCompile(reText)
	return re.MatchString(input)
}

func TestSingleGenerate(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 1
	//We don't do FAKE_GENERATE here because we want to make sure at least one comes back legit.
	options.NO_CACHE = true

	options.fixUp()

	output, errOutput := getOutput(options)

	if errOutput != "" {
		t.Error("Generating a puzzle expected empty stderr, but got", errOutput)
	}

	grid := sudoku.NewGrid()
	grid.Load(output)

	if grid.Invalid() || grid.Empty() {
		t.Error("Output for single generate was not a valid puzzle", output)
	}

}

func TestMultiGenerate(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 3
	//TestSingleGenerate already validated that generation worked; so now we can cut corners to save time.
	options.FAKE_GENERATE = true
	options.NO_CACHE = true

	options.fixUp()

	output, _ := getOutput(options)

	if !regularExpressionMatch(GRID_RE+GRID_RE+GRID_RE, output) {
		t.Error("Output didn't match the expected RE for a grid", output)
	}
}

func TestNoProgress(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 2
	options.NO_PROGRESS = true
	options.FAKE_GENERATE = true
	options.NO_CACHE = true

	options.fixUp()

	_, errOutput := getOutput(options)

	if errOutput != "" {
		t.Error("Generating multiple puzzles with -no-progress expected empty stderr, but got", errOutput)
	}
}

func TestPrintStats(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 1
	options.PRINT_STATS = true
	options.NO_PROGRESS = true
	options.FAKE_GENERATE = true
	options.NO_CACHE = true

	options.fixUp()

	output, _ := getOutput(options)

	re := GRID_RE +
		FLOAT_RE + `\n` +
		OUTPUT_DIVIDER_RE +
		numLineRE("Difficulty", true) +
		OUTPUT_DIVIDER_RE +
		numLineRE("Step count", false) +
		OUTPUT_DIVIDER_RE +
		numLineRE("Avg Dissimilarity", true) +
		OUTPUT_DIVIDER_RE +
		"(" + numLineRE(VARIANT_RE, false) + "){" + strconv.Itoa(len(sudoku.AllTechniqueVariants)) + "}" +
		OUTPUT_DIVIDER_RE

	if !regularExpressionMatch(re, output) {
		t.Error("Output didn't match the expected RE for the output", output)
	}

}

//Callers should call fixUpOptions after receiving this.
func getDefaultOptions() *appOptions {
	options := &appOptions{
		flagSet: flag.NewFlagSet("main", flag.ExitOnError),
	}
	defineFlags(options)
	options.flagSet.Parse([]string{})
	return options
}

func getOutput(options *appOptions) (outputResult string, errorResult string) {

	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}

	process(options, output, errOutput)

	outputReaderBytes, _ := ioutil.ReadAll(output)
	errorReaderBytes, _ := ioutil.ReadAll(errOutput)

	return string(outputReaderBytes), string(errorReaderBytes)
}

func getExpectations(name string) string {
	bytes, _ := ioutil.ReadFile("test_expectations/" + name + ".txt")
	return string(bytes)
}
