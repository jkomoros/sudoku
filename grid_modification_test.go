package sudoku

import (
	"testing"
)

func TestCopyWithModifications(t *testing.T) {
	sourceGrid := NewGrid()
	sourceGrid.MutableCell(0, 0).SetMark(1, true)
	sourceGrid.MutableCell(0, 0).SetExcluded(1, true)

	gridFive := NewGrid()
	gridFive.MutableCell(0, 0).SetNumber(5)

	gridMarks := NewGrid()
	gridMarksCell := gridMarks.MutableCell(0, 0)
	gridMarksCell.SetMark(2, true)
	gridMarksCell.SetMark(1, false)

	gridExcludes := sourceGrid.MutableCopy()
	gridExcludesCell := gridExcludes.MutableCell(0, 0)
	gridExcludesCell.SetExcluded(2, true)
	gridExcludesCell.SetExcluded(1, false)

	tests := []struct {
		modifications GridModifcation
		expected      Grid
		description   string
	}{
		{
			GridModifcation{
				&CellModification{
					Cell:   sourceGrid.Cell(0, 0),
					Number: 5,
				},
			},
			gridFive,
			"Single valid number",
		},
		{
			GridModifcation{
				&CellModification{
					Cell:   sourceGrid.Cell(0, 0),
					Number: DIM,
				},
			},
			sourceGrid,
			"Single invalid number",
		},
		{
			GridModifcation{
				&CellModification{
					Cell:   sourceGrid.Cell(0, 0),
					Number: -1,
					MarksChanges: map[int]bool{
						1:       false,
						2:       true,
						DIM + 1: true,
					},
				},
			},
			gridMarks,
			"Marks",
		},
		{
			GridModifcation{
				&CellModification{
					Cell:   sourceGrid.Cell(0, 0),
					Number: -1,
					ExcludesChanges: map[int]bool{
						1:       false,
						2:       true,
						DIM + 1: true,
					},
				},
			},
			gridExcludes,
			"Excludes",
		},
	}

	for i, test := range tests {
		result := sourceGrid.CopyWithModifications(test.modifications)
		if result.Diagram(true) != test.expected.Diagram(true) {
			t.Error("Test", i, "failed", test.description, "Got", result.Diagram(true), "expected", test.expected.Diagram(true))
		}
	}

}
