package sdkconverter

import (
	"github.com/jkomoros/sudoku"
	"io/ioutil"
	"testing"
)

func TestDokuConverter(t *testing.T) {
	c := Converters["doku"]
	if c == nil {
		t.Error("Couldn't find the doku converter")
	}
}

func TestKomoConverterLoad(t *testing.T) {
	tests := [][2]string{
		{"converter_one_komo.sdk", "converter_one.sdk"},
		{"converter_two_komo.sdk", "converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, true, "komo", test[0], test[1])
	}

	//Test marks come in
	grid := Load(loadTestPuzzle("komo_with_marks.sdk"))
	cell := grid.Cell(0, 0)

	if !cell.Marks().SameContentAs(sudoku.IntSlice{2, 3, 4}) {
		t.Error("Komo importer didn't bring in marks. Got", cell.Marks())
	}
	if cell.Number() != 9 {
		t.Error("Komo importer didn't bring in user-filled number")
	}
}

func TestKomoRoundTrip(t *testing.T) {

	files := []string{"converter_one_komo.sdk",
		"converter_two_komo.sdk",
		"komo_with_marks.sdk",
	}

	for _, test := range files {

		puzzle := loadTestPuzzle(test)

		converter := Converters["komo"]
		grid := sudoku.NewGrid()
		converter.Load(grid, puzzle)
		dataString := converter.DataString(grid)

		if dataString != puzzle {
			t.Error("Failed round trip with file", test, "Got", dataString, "\nExpected", puzzle)
		}

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

func TestKomoConverterValid(t *testing.T) {
	validTestHelper(t, "komo", "converter_one_komo.sdk", true)
	validTestHelper(t, "komo", "converter_two_komo.sdk", true)
	validTestHelper(t, "komo", "komo_with_marks.sdk", true)
	validTestHelper(t, "komo", "invalid_komo_too_short.sdk", false)
	validTestHelper(t, "komo", "converter_one.sdk", false)
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

func TestSDKConverterValid(t *testing.T) {
	validTestHelper(t, "sdk", "converter_one.sdk", true)
	validTestHelper(t, "sdk", "converter_two.sdk", true)
	validTestHelper(t, "sdk", "sdk_no_sep.sdk", true)
	validTestHelper(t, "sdk", "invalid_sdk_invalid_char.sdk", false)
	validTestHelper(t, "sdk", "invalid_sdk_too_short.sdk", false)
	validTestHelper(t, "sdk", "converter_one_komo.sdk", false)
}

func validTestHelper(t *testing.T, format string, file string, expected bool) {
	converter := Converters[format]

	if converter == nil {
		t.Fatal("Couldn't find converter of format", format)
	}

	contents := loadTestPuzzle(file)

	if contents == "" {
		t.Fatal("Couldn't load", file)
	}

	result := converter.Valid(contents)

	if result != expected {
		t.Error("Got wrong result for file", contents, "got", result, "expected", expected)
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

func TestLoadInto(t *testing.T) {
	grid := Load(loadTestPuzzle("converter_one_komo.sdk"))

	expected := loadTestPuzzle("converter_one.sdk")
	if grid.DataString() != expected {
		t.Error("Loading komo puzzle 1 loaded wrong; got", grid.DataString(), "wanted ", expected)
	}

	LoadInto(grid, loadTestPuzzle("converter_two_komo.sdk"))

	expected = loadTestPuzzle("converter_two.sdk")

	if grid.DataString() != expected {
		t.Error("Loading komo puzzle 2 loaded wrong; got", grid.DataString(), "wanted", expected)
	}

}

func TestFormat(t *testing.T) {
	result := Format(loadTestPuzzle("converter_one.sdk"))
	if result != "sdk" {
		t.Error("Format guessed wrong format:", result)
	}
	result = Format(loadTestPuzzle("converter_one_komo.sdk"))
	if result != "komo" {
		t.Error("Format guessed wrong format for komo puzzle: ", result)
	}
	result = Format(loadTestPuzzle("invalid_sdk_too_short.sdk"))
	if result != "" {
		t.Error("Format guessed wrong format for an unknown puzzle type", result)
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
			t.Error("Loading", otherFile, sdkFile, format, "Expected", sdk, "got", grid.DataString(), "for input", other)
		}
	} else {
		grid.LoadSDK(sdk)

		data := converter.DataString(grid)

		if data != other {
			t.Error("DataString", otherFile, sdkFile, format, "Expected", other, "got", data, "for input", sdk)
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
