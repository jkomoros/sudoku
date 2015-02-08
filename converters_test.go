package sudoku

import (
	"io/ioutil"
	"testing"
)

func TestKomoConverter(t *testing.T) {
	tests := [][2]string{
		{"converter_one_komo.sdk", "converter_one.sdk"},
		{"converter_two_komo.sdk", "converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, true, "komo", test[0], test[1])
	}
}

func converterTesterHelper(t *testing.T, testLoad bool, format string, otherFile string, sdkFile string) {
	converter := Converters[format]

	if converter == nil {
		t.Fatal("Couldn't find converter of format", format)
	}

	otherFile = puzzlePath(otherFile)
	sdkFile = puzzlePath(sdkFile)

	if otherFile == "" {
		t.Fatal("Couldn't find puzzle at", otherFile)
	}

	if sdkFile == "" {
		t.Fatal("Couldn't find puzzle at", sdkFile)
	}

	var other string
	var sdk string

	if data, err := ioutil.ReadFile(otherFile); err == nil {
		other = string(data)
	}

	if data, err := ioutil.ReadFile(sdkFile); err == nil {
		sdk = string(data)
	}

	grid := NewGrid()

	if testLoad {

		converter.Load(grid, other)

		if grid.DataString() != sdk {
			t.Error("Expected", sdk, "got", grid.DataString(), "for input", other)
		}
	} else {
		t.Fatal("TesterHelper doesn't support testing loading right now.")
	}
}
