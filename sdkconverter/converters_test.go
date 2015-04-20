package sdkconverter

import (
	"github.com/jkomoros/sudoku"
	"io/ioutil"
	"testing"
)

func TestKomoConverterLoad(t *testing.T) {
	tests := [][2]string{
		{"converter_one_komo.sdk", "converter_one.sdk"},
		{"converter_two_komo.sdk", "converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, true, "komo", test[0], test[1])
	}
}

func TestKomoConverterDataString(t *testing.T) {
	tests := [][2]string{
		{"converter_one_komo.sdk", "converter_one.sdk"},
		{"converter_two_komo.sdk", "converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, false, "komo", test[0], test[1])
	}
}

func TestSDKConverterLoad(t *testing.T) {
	tests := [][2]string{
		{"converter_one.sdk", "converter_one.sdk"},
		{"converter_two.sdk", "converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, true, "sdk", test[0], test[1])
	}
}

func TestSDKConverterDataString(t *testing.T) {
	tests := [][2]string{
		{"converter_one.sdk", "converter_one.sdk"},
		{"converter_two.sdk", "converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, false, "sdk", test[0], test[1])
	}
}

func TestConvenienceFuncs(t *testing.T) {
	sdk := loadTestPuzzle("converter_one.sdk")
	other := loadTestPuzzle("converter_one_komo.sdk")

	result := ToSDK("komo", other)

	if result != sdk {
		t.Error("Testing ToSDK, expected", sdk, "got", result)
	}

	result = ToOther("komo", sdk)

	if result != other {
		t.Error("Testing ToOther, expected", other, "got", result)
	}
}

func converterTesterHelper(t *testing.T, testLoad bool, format string, otherFile string, sdkFile string) {

	converter := Converters[format]

	if converter == nil {
		t.Fatal("Couldn't find converter of format", format)
	}

	other := loadTestPuzzle(otherFile)
	sdk := loadTestPuzzle(sdkFile)

	if other == "" {
		t.Fatal("Couldn't load", otherFile)
	}

	if sdk == "" {
		t.Fatal("Couldn't load", sdkFile)
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

func loadTestPuzzle(puzzleName string) string {
	data, err := ioutil.ReadFile("puzzles/" + puzzleName)

	if err != nil {
		return ""
	}

	return string(data)
}
