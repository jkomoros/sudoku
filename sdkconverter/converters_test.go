package sdkconverter

import (
	"github.com/jkomoros/sudoku"
	"io/ioutil"
	"testing"
)

func TestDokuConverterValid(t *testing.T) {
	//All sdk files are valid dokus
	validTestHelper(t, "doku", "converter_one.sdk", true)
	validTestHelper(t, "doku", "converter_two.sdk", true)
	validTestHelper(t, "doku", "sdk_no_sep.sdk", true)

	//Two SDKs it shouldn't match
	validTestHelper(t, "doku", "komo_with_marks.sdk", false)
	validTestHelper(t, "doku", "invalid_sdk_too_short.sdk", false)

	validTestHelper(t, "doku", "doku_complex.doku", true)

}

func TestDokuConverterLoad(t *testing.T) {
	tests := [][2]string{
		{"converter_one_doku.doku", "converter_one.sdk"},
		{"converter_two_doku.doku", "converter_two.sdk"},
	}
	for _, test := range tests {
		converterTesterHelper(t, true, "doku", test[0], test[1])
	}

	//Test that when we load an SDK we don't lock the filled cells.
	grid := Load(loadTestPuzzle("converter_one.sdk"))
	cell := grid.Cell(0, 1)
	if cell.Locked() {
		t.Error("Loading a doku that was a valid sdk erroneously locked a cell")
	}

	//Test locked numbers and marks, since the two tests above can't exercise
	//those since sdk doesn't have them
	grid = Load(loadTestPuzzle("doku_complex.doku"))
	cell = grid.Cell(0, 0)

	if !cell.Marks().SameContentAs(sudoku.IntSlice{2, 3, 4}) {
		t.Error("Doku importer didn't bring in marks. Got", cell.Marks())
	}

	cell = grid.Cell(0, 1)

	if !cell.Locked() {
		t.Error("Locked cell wasn't actually locked")
	}

}

func TestDokuConverterDataString(t *testing.T) {

	complexDoku := loadTestPuzzle("doku_complex.doku")
	complexDokuNormalized := loadTestPuzzle("doku_complex_normalized.doku")

	grid := Load(complexDoku)

	doku := Converters["doku"]

	dokuDataString := doku.DataString(grid)

	if dokuDataString != complexDokuNormalized {
		t.Error("Datastring for complex doku wrong. Got", dokuDataString, "expected", complexDokuNormalized)
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
	//This next puzzle is saved windows line encodings but should still validate.
	validTestHelper(t, "sdk", "nakedpair3.sdk", true)
}

func validTestHelper(t *testing.T, format Format, file string, expected bool) {
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

func TestDataString(t *testing.T) {

	dokuInput := loadTestPuzzle("doku_complex_normalized.doku")
	grid := Load(dokuInput)

	converted := DataString("doku", grid)

	if converted != dokuInput {
		t.Error("DataString for doku didn't work. Got", converted, "wanted", dokuInput)
	}

	converted = DataString("foo", grid)

	if converted != "" {
		t.Error("Unsuccessful data string thought it was OK")
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
	result := PuzzleFormat(loadTestPuzzle("converter_one.sdk"))
	//doku and sdk are both valid options
	if result != "sdk" && result != "doku" {
		t.Error("Format guessed wrong format:", result)
	}
	result = PuzzleFormat(loadTestPuzzle("converter_one_komo.sdk"))
	if result != "komo" {
		t.Error("Format guessed wrong format for komo puzzle: ", result)
	}
	result = PuzzleFormat(loadTestPuzzle("doku_complex.doku"))
	if result != "doku" {
		t.Error("Format guessed wrong format for doku puzzle: ", result)
	}
	result = PuzzleFormat(loadTestPuzzle("invalid_sdk_too_short.sdk"))
	if result != "" {
		t.Error("Format guessed wrong format for an unknown puzzle type", result)
	}
}

func converterTesterHelper(t *testing.T, testLoad bool, format Format, otherFile string, sdkFile string) {

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
