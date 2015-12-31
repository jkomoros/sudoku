package main

import (
	"bytes"
	"encoding/csv"
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
const FLOAT_RE = `(\d{1,5}\.\d{4,20}|0)`
const INT_RE = `\b\d{1,5}\b`

const KOMO_PUZZLE = "6!,1!,2!,7,5,8,4!,9,3!;8,3!,5,4!,9!,6,1,7!,2!;9,4,7!,2,1,3,8,6!,5!;2,5,9,3,6!,1!,7,8!,4;1!,7,3!,8,4!,9,2!,5,6!;4,6!,8,5!,2!,7,3,1,9;3,9!,6,1,7,4,5!,2,8;7!,2!,4,9,8!,5!,6,3!,1;5!,8,1!,6,3,2,9!,4!,7!"
const SOLVED_KOMO_PUZZLE = "6!,1!,2!,7!,5!,8!,4!,9!,3!;8!,3!,5!,4!,9!,6!,1!,7!,2!;9!,4!,7!,2!,1!,3!,8!,6!,5!;2!,5!,9!,3!,6!,1!,7!,8!,4!;1!,7!,3!,8!,4!,9!,2!,5!,6!;4!,6!,8!,5!,2!,7!,3!,1!,9!;3!,9!,6!,1!,7!,4!,5!,2!,8!;7!,2!,4!,9!,8!,5!,6!,3!,1!;5!,8!,1!,6!,3!,2!,9!,4!,7!"

const SOLVED_TEST_GRID = `6|1|2|7|5|8|4|9|3
8|3|5|4|9|6|1|7|2
9|4|7|2|1|3|8|6|5
2|5|9|3|6|1|7|8|4
1|7|3|8|4|9|2|5|6
4|6|8|5|2|7|3|1|9
3|9|6|1|7|4|5|2|8
7|2|4|9|8|5|6|3|1
5|8|1|6|3|2|9|4|7`

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

func TestCSVExportKomo(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 2
	options.FAKE_GENERATE = true
	options.NO_CACHE = true
	options.CSV = true
	options.PRINT_STATS = true
	options.PUZZLE_FORMAT = "komo"
	options.NO_PROGRESS = true

	expectUneventfulFixup(t, options)

	output, errOutput := getOutput(options)

	if output == "" {
		t.Fatal("Got no output from CSV export komo")
	}

	csvReader := csv.NewReader(strings.NewReader(output))

	recs, err := csvReader.ReadAll()

	if errOutput != "" {
		t.Error("For CSV generation expected no error output, got", errOutput)
	}

	if err != nil {
		t.Fatal("CSV export was not a valid CSV", err, output)
	}

	for i, rec := range recs {
		if rec[0] != KOMO_PUZZLE {
			t.Error("On line", i, "of the CSV col 1 expected", KOMO_PUZZLE, ", but got", rec[0])
		}
		if !regularExpressionMatch(FLOAT_RE, rec[1]) {
			t.Error("On line", i, "of the CSV col 2 expected a float, but got", rec[1])
		}
	}

}

func TestCSVExport(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 2
	options.FAKE_GENERATE = true
	options.NO_CACHE = true
	options.CSV = true
	options.PRINT_STATS = true
	options.NO_PROGRESS = true

	expectUneventfulFixup(t, options)

	output, errOutput := getOutput(options)

	if output == "" {
		t.Fatal("Got no output from CSV export")
	}

	csvReader := csv.NewReader(strings.NewReader(output))

	recs, err := csvReader.ReadAll()

	if errOutput != "" {
		t.Error("For CSV generation expected no error output, got", errOutput)
	}

	if err != nil {
		t.Fatal("CSV export was not a valid CSV", err, output)
	}

	for i, rec := range recs {
		if rec[0] != TEST_GRID {
			t.Error("On line", i, "of the CSV col 1 expected", TEST_GRID, ", but got", rec[0])
		}
		if !regularExpressionMatch(FLOAT_RE, rec[1]) {
			t.Error("On line", i, "of the CSV col 2 expected a float, but got", rec[1])
		}
	}

}

func TestCSVImport(t *testing.T) {

	//TODO: test importing files that are not in the expected format

	options := getDefaultOptions()

	options.CSV = true
	options.NO_PROGRESS = true
	options.PUZZLE_FORMAT = "komo"
	options.PUZZLE_TO_SOLVE = "tests/input.csv"

	expectUneventfulFixup(t, options)

	output, errOutput := getOutput(options)

	csvReader := csv.NewReader(strings.NewReader(output))

	recs, err := csvReader.ReadAll()

	if output == "" {
		t.Fatal("Got no output from CSV import")
	}

	if errOutput != "" {
		t.Error("For CSV import expected no error output, got", errOutput)
	}

	if err != nil {
		t.Fatal("CSV export was not a valid CSV", err, output)
	}

	for i, rec := range recs {
		if rec[0] != SOLVED_KOMO_PUZZLE {
			t.Error("On line", i, "of the CSV col 1 expected", SOLVED_KOMO_PUZZLE, ", but got", rec[0])
		}
	}
}

func TestPuzzleImportKomo(t *testing.T) {
	options := getDefaultOptions()

	options.PUZZLE_TO_SOLVE = "tests/puzzle_komo.sdk"
	options.PUZZLE_FORMAT = "komo"

	expectUneventfulFixup(t, options)

	output, errOutput := getOutput(options)

	if errOutput != "" {
		t.Error("For puzzle import expected no error output, got", errOutput)
	}

	if output != SOLVED_KOMO_PUZZLE+"\n" {
		t.Error("For puzzle import with komo format expected", SOLVED_KOMO_PUZZLE+"\n", "got", output)
	}
}

func TestPuzzleImport(t *testing.T) {
	options := getDefaultOptions()

	options.PUZZLE_TO_SOLVE = "tests/puzzle.sdk"

	expectUneventfulFixup(t, options)

	output, errOutput := getOutput(options)

	if errOutput != "" {
		t.Error("For puzzle import expected no error output, got", errOutput)
	}

	if output != SOLVED_TEST_GRID+"\n" {
		t.Error("For puzzle import expected", SOLVED_TEST_GRID+"\n", "got", output)
	}
}

//TODO: test walkthrough

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

	if output != KOMO_PUZZLE+"\n" {
		t.Error("Didn't get right output for komo format. Got*", output, "* expected *", KOMO_PUZZLE+"\n", "*")
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
