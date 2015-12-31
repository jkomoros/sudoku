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

func expectUneventfulFixup(t *testing.T, options *appOptions) {
	errWriter := &bytes.Buffer{}

	options.fixUp(errWriter)

	errorReaderBytes, _ := ioutil.ReadAll(errWriter)

	errOutput := string(errorReaderBytes)

	if errOutput != "" {
		t.Error("The options fixup was expected to be uneventful but showed", errOutput)
	}
}

//TODO: test CSV output

//TODO: test inputting of a puzzle

func TestHelp(t *testing.T) {

	options := getDefaultOptions()

	options.HELP = true

	expectUneventfulFixup(t, options)

	output, errOutput := getOutput(options)

	//The output of -h is very finicky with tabs/spaces, and it's constnatly changing.
	//So our golden will just be a generated version of the help message.
	helpGoldenBuffer := &bytes.Buffer{}
	options.flagSet.SetOutput(helpGoldenBuffer)
	options.flagSet.PrintDefaults()

	helpGoldenBytes, _ := ioutil.ReadAll(helpGoldenBuffer)

	expectations := string(helpGoldenBytes)

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

	expectUneventfulFixup(t, options)

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

	expectUneventfulFixup(t, options)

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

	expectUneventfulFixup(t, options)

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

	expectUneventfulFixup(t, options)

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

func TestPuzzleFormat(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 1
	options.NO_PROGRESS = true
	options.FAKE_GENERATE = true
	options.PUZZLE_FORMAT = "komo"
	options.NO_CACHE = true

	expectUneventfulFixup(t, options)

	output, _ := getOutput(options)

	expectedKomoPuzzle := "6!,1!,2!,7,5,8,4!,9,3!;8,3!,5,4!,9!,6,1,7!,2!;9,4,7!,2,1,3,8,6!,5!;2,5,9,3,6!,1!,7,8!,4;1!,7,3!,8,4!,9,2!,5,6!;4,6!,8,5!,2!,7,3,1,9;3,9!,6,1,7,4,5!,2,8;7!,2!,4,9,8!,5!,6,3!,1;5!,8,1!,6,3,2,9!,4!,7!\n"

	if output != expectedKomoPuzzle {
		t.Error("Didn't get right output for komo format. Got*", output, "* expected *", expectedKomoPuzzle, "*")
	}

}

func TestInvalidPuzzleFormat(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 1
	options.NO_PROGRESS = true
	options.FAKE_GENERATE = true
	options.NO_CACHE = true
	options.PUZZLE_FORMAT = "foo"

	errWriter := &bytes.Buffer{}

	options.fixUp(errWriter)

	errorReaderBytes, _ := ioutil.ReadAll(errWriter)

	errOutput := string(errorReaderBytes)

	if !strings.Contains(errOutput, "Invalid format option: foo") {
		t.Error("Expected an error message about invalid format option. Wanted 'Invalid format option:foo', got", errOutput)
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
