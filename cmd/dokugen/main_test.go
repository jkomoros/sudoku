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
	getOutput(options)

	//TODO: compare this to expect output
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
