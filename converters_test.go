package sudoku

import (
	"io/ioutil"
	"testing"
)

func TestKomoConverter(t *testing.T) {
	//TODO: test a puzzle where everything is filled.
	tests := [][2]string{
		{"converter_one_komo.sdk", "converter_one.sdk"},
	}
	for _, test := range tests {
		converterDataStringTesterHelper(t, "komo", test[0], test[1])
	}
}

func converterDataStringTesterHelper(t *testing.T, format string, inputFile string, expectedSDKFile string) {
	converter := Converters[format]

	if converter == nil {
		t.Fatal("Couldn't find converter of format", format)
	}

	inputFile = puzzlePath(inputFile)
	expectedSDKFile = puzzlePath(expectedSDKFile)

	if inputFile == "" {
		t.Fatal("Couldn't find puzzle at", inputFile)
	}

	if expectedSDKFile == "" {
		t.Fatal("Couldn't find puzzle at", expectedSDKFile)
	}

	var input string
	var expectedSDK string

	if data, err := ioutil.ReadFile(inputFile); err == nil {
		input = string(data)
	}

	if data, err := ioutil.ReadFile(expectedSDKFile); err == nil {
		expectedSDK = string(data)
	}

	grid := NewGrid()

	converter.Load(grid, input)

	if grid.DataString() != expectedSDK {
		t.Error("Expected", expectedSDK, "got", grid.DataString(), "for input", input)
	}
}
