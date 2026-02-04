package wekaparser

import (
	"os"
	"reflect"
	"testing"
)

func TestParseWeights(t *testing.T) {

	expectedWeights := map[string]float64{
		"Block Block Interactions Count":      -0.0182,
		"Block Block Interactions Percentage": 0.0034,
		"Constant":                            0.0577,
		"Forcing Chain (1 steps) Count":       0.0,
		"Forcing Chain (1 steps) Percentage":  0.0,
		"Forcing Chain (2 steps) Count":       0.0,
		"Forcing Chain (2 steps) Percentage":  0.0,
		"Forcing Chain (3 steps) Count":       0.0,
		"Forcing Chain (3 steps) Percentage":  0.0,
		"Forcing Chain (4 steps) Count":       0.0061,
		"Forcing Chain (4 steps) Percentage":  0.0011,
		"Forcing Chain (5 steps) Count":       0.011,
		"Forcing Chain (5 steps) Percentage":  0.0154,
		"Forcing Chain (6 steps) Count":       0.0286,
		"Forcing Chain (6 steps) Percentage":  0.0158,
		"Guess Count":                         0.027,
		"Guess Percentage":                    0.0266,
		"Hidden Pair Block Count":             0.1157,
		"Hidden Pair Block Percentage":        0.0035,
		"Hidden Pair Col Count":               -0.0276,
		"Hidden Pair Col Percentage":          -0.0006,
		"Hidden Pair Row Count":               -0.1,
		"Hidden Pair Row Percentage":          -0.0007,
		"Hidden Quad Block Count":             -0.1,
		"Hidden Quad Block Percentage":        -0.0016,
		"Hidden Quad Col Count":               0.0329,
		"Hidden Quad Col Percentage":          0.0004,
		"Hidden Quad Row Count":               0.0332,
		"Hidden Quad Row Percentage":          0.0005,
		"Hidden Triple Block Count":           0.0143,
		"Hidden Triple Block Percentage":      0.0003,
		"Hidden Triple Col Count":             0.1004,
		"Hidden Triple Col Percentage":        0.0015,
		"Hidden Triple Row Count":             0.0,
		"Hidden Triple Row Percentage":        0.0,
		"Naked Pair Block Count":              0.0098,
		"Naked Pair Block Percentage":         0.0213,
		"Naked Pair Col Count":                0.0068,
		"Naked Pair Col Percentage":           0.0071,
		"Naked Pair Row Count":                0.001,
		"Naked Pair Row Percentage":           0.0325,
		"Naked Quad Block Count":              -0.0167,
		"Naked Quad Block Percentage":         0.0073,
		"Naked Quad Col Count":                -0.0409,
		"Naked Quad Col Percentage":           -0.0001,
		"Naked Quad Row Count":                -0.0055,
		"Naked Quad Row Percentage":           0.0052,
		"Naked Triple Block Count":            0.0213,
		"Naked Triple Block Percentage":       0.0081,
		"Naked Triple Col Count":              -0.0265,
		"Naked Triple Col Percentage":         0.0114,
		"Naked Triple Row Count":              -0.01,
		"Naked Triple Row Percentage":         0.009,
		"Necessary In Block Count":            -0.0126,
		"Necessary In Block Percentage":       -0.0867,
		"Necessary In Col Count":              0.0027,
		"Necessary In Col Percentage":         -0.0146,
		"Necessary In Row Count":              0.009,
		"Necessary In Row Percentage":         -0.5822,
		"Number Unfilled Cells":               0.0237,
		"Number of Steps":                     -0.0037,
		"Obvious In Block Count":              -0.0188,
		"Obvious In Block Percentage":         0.2085,
		"Obvious In Col Count":                -0.0146,
		"Obvious In Col Percentage":           0.0286,
		"Obvious In Row Count":                -0.0109,
		"Obvious In Row Percentage":           0.2178,
		"Only Legal Number Count":             -0.0039,
		"Only Legal Number Percentage":        0.0317,
		"Percentage Fill Steps":               -0.1378,
		"Pointing Pair Col Count":             -0.0015,
		"Pointing Pair Col Percentage":        0.0096,
		"Pointing Pair Row Count":             -0.0074,
		"Pointing Pair Row Percentage":        0.0109,
		"Steps Until Nonfill":                 -0.0024,
		"Swordfish Col Count":                 -0.0342,
		"Swordfish Col Percentage":            0.0045,
		"Swordfish Row Count":                 -0.0475,
		"Swordfish Row Percentage":            0.0017,
		"XWing Col Count":                     0.0006,
		"XWing Col Percentage":                0.0014,
		"XWing Row Count":                     0.0271,
		"XWing Row Percentage":                0.0011,
		"XYWing (Same Block) Count":           0.0085,
		"XYWing (Same Block) Percentage":      0.0015,
		"XYWing Count":                        0.037,
		"XYWing Percentage":                   -0.0014,
	}

	data, err := os.ReadFile("test_input.txt")

	if err != nil {
		t.Error("Couldn't read file:", err)
	}

	weights, err := ParseWeights(string(data))

	if err != nil {
		t.Error("ParseWeights gave an error:", err)
	}

	if !reflect.DeepEqual(expectedWeights, weights) {
		t.Error("Didn't get the weights we expected. Got:", weights, "Expected:", expectedWeights)
	}
}

func TestParseR2(t *testing.T) {

	expectedR2 := 0.7681

	data, err := os.ReadFile("test_input.txt")

	if err != nil {
		t.Error("Couldn't read file:", err)
	}

	r2, err := ParseR2(string(data))

	if err != nil {
		t.Error("Got error from ParseR2", err)
	}

	if r2 != expectedR2 {
		t.Error("Got wrong R2. Got", r2, "expcted", expectedR2)
	}
}
