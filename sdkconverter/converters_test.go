package sdkconverter

import (
	"dokugen"
	"io/ioutil"
	"testing"
)

func TestKomoConverterLoad(t *testing.T) {
	tests := [][2]string{
		{"puzzles/converter_one_komo.sdk", "puzzles/converter_one.sdk"},
		{"puzzles/converter_two_komo.sdk", "puzzles/converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, true, "komo", test[0], test[1])
	}
}

func TestKomoConverterDataString(t *testing.T) {
	tests := [][2]string{
		{"puzzles/converter_one_komo.sdk", "puzzles/converter_one.sdk"},
		{"puzzles/converter_two_komo.sdk", "puzzles/converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, false, "komo", test[0], test[1])
	}
}

func converterTesterHelper(t *testing.T, testLoad bool, format string, otherFile string, sdkFile string) {

	converter := Converters[format]

	if converter == nil {
		t.Fatal("Couldn't find converter of format", format)
	}

	var other string
	var sdk string

	if data, err := ioutil.ReadFile(otherFile); err == nil {
		other = string(data)
	}

	if data, err := ioutil.ReadFile(sdkFile); err == nil {
		sdk = string(data)
	}

	grid := sudoku.NewGrid()

	if testLoad {

		converter.Load(grid, other)

		if grid.DataString() != sdk {
			t.Error("Expected", sdk, "got", grid.DataString(), "for input", other)
		}
	} else {
		grid.Load(sdk)

		data := converter.DataString(grid)

		if data != other {
			t.Error("Expected", other, "got", data, "for input", sdk)
		}
	}
}
