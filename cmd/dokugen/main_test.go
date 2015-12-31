package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"testing"
)

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
