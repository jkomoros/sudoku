//auto-generated by difficulty-convert.py DO NOT EDIT

package sudoku

//DIFFICULTY_MODEL is a unique string representing the exact difficulty
//model in use.Every time a new model is trained, this value will change.
//Therefore, if the value is different than last time you checked, the model has changed.
//This is useful for throwing out caches that assume the same difficulty model is in use.
const DIFFICULTY_MODEL = "1373B1EA7DFB998D0DC409B83308BB"

func init() {
	difficultySignalWeights = map[string]float64{
		"Block Block Interactions Count":      -0.0202,
		"Block Block Interactions Percentage": 0.0068,
		"Constant":                            0.5446,
		"Forcing Chain (1 steps) Count":       0.0,
		"Forcing Chain (1 steps) Percentage":  0.0,
		"Forcing Chain (2 steps) Count":       0.0,
		"Forcing Chain (2 steps) Percentage":  0.0,
		"Forcing Chain (3 steps) Count":       0.0515,
		"Forcing Chain (3 steps) Percentage":  0.0035,
		"Forcing Chain (4 steps) Count":       -0.0108,
		"Forcing Chain (4 steps) Percentage":  0.0045,
		"Forcing Chain (5 steps) Count":       0.0122,
		"Forcing Chain (5 steps) Percentage":  0.0112,
		"Forcing Chain (6 steps) Count":       0.0282,
		"Forcing Chain (6 steps) Percentage":  0.0224,
		"Guess Count":                         0.0042,
		"Guess Percentage":                    0.0289,
		"Hidden Pair Block Count":             -0.0295,
		"Hidden Pair Block Percentage":        0.0136,
		"Hidden Pair Col Count":               0.0001,
		"Hidden Pair Col Percentage":          0.0128,
		"Hidden Pair Row Count":               -0.0151,
		"Hidden Pair Row Percentage":          0.0069,
		"Hidden Quad Block Count":             0.0816,
		"Hidden Quad Block Percentage":        0.0016,
		"Hidden Quad Col Count":               0.1234,
		"Hidden Quad Col Percentage":          0.0027,
		"Hidden Quad Row Count":               0.0468,
		"Hidden Quad Row Percentage":          0.0014,
		"Hidden Triple Block Count":           0.0553,
		"Hidden Triple Block Percentage":      0.0048,
		"Hidden Triple Col Count":             0.0171,
		"Hidden Triple Col Percentage":        0.0009,
		"Hidden Triple Row Count":             0.1262,
		"Hidden Triple Row Percentage":        0.0055,
		"Naked Pair Block Count":              -0.0242,
		"Naked Pair Block Percentage":         0.0572,
		"Naked Pair Col Count":                -0.0281,
		"Naked Pair Col Percentage":           0.0564,
		"Naked Pair Row Count":                -0.0269,
		"Naked Pair Row Percentage":           0.0079,
		"Naked Quad Block Count":              -0.0143,
		"Naked Quad Block Percentage":         0.0284,
		"Naked Quad Col Count":                0.0191,
		"Naked Quad Col Percentage":           0.0053,
		"Naked Quad Row Count":                -0.0274,
		"Naked Quad Row Percentage":           0.0185,
		"Naked Triple Block Count":            -0.0205,
		"Naked Triple Block Percentage":       0.026,
		"Naked Triple Col Count":              -0.0285,
		"Naked Triple Col Percentage":         0.0527,
		"Naked Triple Row Count":              -0.0223,
		"Naked Triple Row Percentage":         0.0244,
		"Necessary In Block Count":            -0.0298,
		"Necessary In Block Percentage":       0.3151,
		"Necessary In Col Count":              -0.0028,
		"Necessary In Col Percentage":         -0.2226,
		"Necessary In Row Count":              -0.0038,
		"Necessary In Row Percentage":         -0.3723,
		"Number Unfilled Cells":               -0.0013,
		"Number of Steps":                     0.0249,
		"Obvious In Block Count":              -0.0213,
		"Obvious In Block Percentage":         -0.0336,
		"Obvious In Col Count":                -0.0072,
		"Obvious In Col Percentage":           -0.0621,
		"Obvious In Row Count":                -0.0027,
		"Obvious In Row Percentage":           -0.2218,
		"Only Legal Number Count":             -0.0192,
		"Only Legal Number Percentage":        0.0393,
		"Percentage Fill Steps":               -0.4875,
		"Pointing Pair Col Count":             -0.0289,
		"Pointing Pair Col Percentage":        0.0525,
		"Pointing Pair Row Count":             -0.015,
		"Pointing Pair Row Percentage":        0.0585,
		"Steps Until Nonfill":                 -0.0076,
		"Swordfish Col Count":                 0.0559,
		"Swordfish Col Percentage":            0.003,
		"Swordfish Row Count":                 -0.1068,
		"Swordfish Row Percentage":            0.0002,
		"XWing Col Count":                     0.0023,
		"XWing Col Percentage":                0.0076,
		"XWing Row Count":                     -0.0209,
		"XWing Row Percentage":                0.0092,
		"XYWing (Same Block) Count":           -0.0286,
		"XYWing (Same Block) Percentage":      0.023,
		"XYWing Count":                        -0.0447,
		"XYWing Percentage":                   -0.0004,
	}
}
