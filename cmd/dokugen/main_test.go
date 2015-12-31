package main

import (
	"bytes"
	"flag"
	"github.com/jkomoros/sudoku"
	"io/ioutil"
	"regexp"
	"testing"
)

const GRID_RE = `((\d\||\.\|){8}(\d|\.)\n){8}((\d\||\.\|){8}(\d|\.))`

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
	options.NO_CACHE = true

	options.fixUp()

	output, errOutput := getOutput(options)

	if errOutput != "" {
		t.Error("Generating a puzzle expected empty stderr, but got", errOutput)
	}

	if !regularExpressionMatch(GRID_RE, output) {
		t.Error("Output didn't match the expected RE for a grid", output)
	}

	grid := sudoku.NewGrid()
	grid.Load(output)

	if grid.Invalid() || grid.Empty() {
		t.Error("Output for single generate was not a valid puzzle", output)
	}

}

func TestNoProgress(t *testing.T) {
	options := getDefaultOptions()

	options.GENERATE = true
	options.NUM = 2
	options.NO_PROGRESS = true
	options.NO_CACHE = true

	options.fixUp()

	_, errOutput := getOutput(options)

	if errOutput != "" {
		t.Error("Generating multiple puzzles with -no-progress expected empty stderr, but got", errOutput)
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
