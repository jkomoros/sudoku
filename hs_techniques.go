package sudoku

import (
	"log"
	"math"
)

/*
	This file is where the basic solve technique infrastructure is defined.

	The techniques that Human Solve uses are initalized and stored into human_solves.go Techniques slice here.

	Specific techniques are implemented in hst_*.go files (hst == human solve techniques), where there's
	a separate file for each class of technique.

*/

//You shouldn't just create a technique by hand (the interior basicSolveTechnique needs to be initalized with the right values)
//So in the rare cases where you want to grab a technique by name, grab it from here.
//TODO: it feels like a pattern smell that there's only a singleton for each technique that you can't cons up on demand.
var techniquesByName map[string]SolveTechnique

//SolveTechnique is a logical technique that, when applied to a grid, returns potential SolveSteps
//that will move the grid closer to being solved, and are based on sound logical reasoning. A stable
//of SolveTechniques (stored in Techniques) are repeatedly applied to the Grid in HumanSolve.
type SolveTechnique interface {
	//Name returns the human-readable shortname of the technique.
	Name() string
	//Description returns a human-readable phrase that describes the logical reasoning applied in the particular step; why
	//it is valid.
	Description(*SolveStep) string
	//IMPORTANT: a step should return a step IFF that step is valid AND the step would cause useful work to be done if applied.

	//Find returns as many steps as it can find in the grid for that technique, in a random order.
	//HumanSolve repeatedly applies technique.Find() to identify candidates for the next step in the solution.
	Find(*Grid) []*SolveStep
	//IsFill returns true if the techinque's action when applied to a grid is to fill a number (as opposed to culling possbilitie).
	IsFill() bool
	//HumanLikelihood is how likely a user would be to pick this technique when compared with other possible steps.
	//Generally inversely related to difficulty (but not perfectly).
	//This value will be used to pick which technique to apply when compared with other candidates.
	HumanLikelihood() float64
}

type cellGroupType int

const (
	_GROUP_NONE cellGroupType = iota
	_GROUP_ROW
	_GROUP_COL
	_GROUP_BLOCK
)

type basicSolveTechnique struct {
	name      string
	isFill    bool
	groupType cellGroupType
	//Size of set in technique, e.g. single = 1, pair = 2, triple = 3
	//Used for generating descriptions in some sub-structs.
	k int
}

func init() {

	//TODO: calculate more realistic weights.

	CheapTechniques = []SolveTechnique{
		&hiddenSingleTechnique{
			&basicSolveTechnique{
				//TODO: shouldn't this be "Hidden Single Row" (and likewise for others)
				"Necessary In Row",
				true,
				_GROUP_ROW,
				1,
			},
		},
		&hiddenSingleTechnique{
			&basicSolveTechnique{
				"Necessary In Col",
				true,
				_GROUP_COL,
				1,
			},
		},
		&hiddenSingleTechnique{
			&basicSolveTechnique{
				"Necessary In Block",
				true,
				_GROUP_BLOCK,
				1,
			},
		},
		&nakedSingleTechnique{
			&basicSolveTechnique{
				//TODO: shouldn't this name be Naked Single for consistency?
				"Only Legal Number",
				true,
				_GROUP_NONE,
				1,
			},
		},
		&obviousInCollectionTechnique{
			&basicSolveTechnique{
				"Obvious In Row",
				true,
				_GROUP_ROW,
				1,
			},
		},
		&obviousInCollectionTechnique{
			&basicSolveTechnique{
				"Obvious In Col",
				true,
				_GROUP_COL,
				1,
			},
		},
		&obviousInCollectionTechnique{
			&basicSolveTechnique{
				"Obvious In Block",
				true,
				_GROUP_BLOCK,
				1,
			},
		},
		&pointingPairTechnique{
			&basicSolveTechnique{
				"Pointing Pair Row",
				false,
				_GROUP_ROW,
				2,
			},
		},
		&pointingPairTechnique{
			&basicSolveTechnique{
				"Pointing Pair Col",
				false,
				_GROUP_COL,
				2,
			},
		},
		&blockBlockInteractionTechnique{
			&basicSolveTechnique{
				"Block Block Interactions",
				false,
				_GROUP_BLOCK,
				2,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Pair Col",
				false,
				_GROUP_COL,
				2,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Pair Row",
				false,
				_GROUP_ROW,
				2,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Pair Block",
				false,
				_GROUP_BLOCK,
				2,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Triple Col",
				false,
				_GROUP_COL,
				3,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Triple Row",
				false,
				_GROUP_ROW,
				3,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Triple Block",
				false,
				_GROUP_BLOCK,
				3,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Quad Col",
				false,
				_GROUP_COL,
				4,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Quad Row",
				false,
				_GROUP_ROW,
				4,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Quad Block",
				false,
				_GROUP_BLOCK,
				4,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Pair Row",
				false,
				_GROUP_ROW,
				2,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Pair Col",
				false,
				_GROUP_COL,
				2,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Pair Block",
				false,
				_GROUP_BLOCK,
				2,
			},
		},
		&xwingTechnique{
			&basicSolveTechnique{
				"XWing Row",
				false,
				_GROUP_ROW,
				2,
			},
		},
		&xwingTechnique{
			&basicSolveTechnique{
				"XWing Col",
				false,
				_GROUP_COL,
				2,
			},
		},
	}

	ExpensiveTechniques = []SolveTechnique{
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Triple Row",
				false,
				_GROUP_ROW,
				3,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Triple Col",
				false,
				_GROUP_COL,
				3,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Triple Block",
				false,
				_GROUP_BLOCK,
				3,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Quad Row",
				false,
				_GROUP_ROW,
				4,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Quad Col",
				false,
				_GROUP_COL,
				4,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Quad Block",
				false,
				_GROUP_BLOCK,
				4,
			},
		},
	}

	GuessTechnique = &guessTechnique{
		&basicSolveTechnique{
			"Guess",
			true,
			_GROUP_NONE,
			1,
		},
	}

	Techniques = append(CheapTechniques, ExpensiveTechniques...)

	AllTechniques = append(Techniques, GuessTechnique)

	techniquesByName = make(map[string]SolveTechnique)

	for _, technique := range AllTechniques {
		techniquesByName[technique.Name()] = technique
	}

}

func (self *basicSolveTechnique) Name() string {
	return self.name
}

func (self *basicSolveTechnique) IsFill() bool {
	return self.isFill
}

//TOOD: this is now named incorrectly. (It should be likelihoodHelper)
func (self *basicSolveTechnique) difficultyHelper(baseDifficulty float64) float64 {
	//Embedding structs should call into this to provide their own Difficulty

	//TODO: the default difficulties, as configured, will mean that SolveDirection's Difficulty() will almost always clamp to 1.0.
	//They're only useful in terms of a reasonable picking of techniques when multiple apply.

	groupMultiplier := 1.0

	switch self.groupType {
	case _GROUP_BLOCK:
		//Blocks are the easiest to notice; although they require zig-zag scanning, the eye doesn't have to move far.
		groupMultiplier = 1.0
	case _GROUP_ROW:
		//Rows are easier to scan than columns because most humans are used to reading LTR
		groupMultiplier = 1.25
	case _GROUP_COL:
		//Cols are easy to scan because the eye can move in one line, but they have to move a long way in an unnatural direction
		groupMultiplier = 1.3
	}

	//TODO: Arguably, the "fill-ness" of a technique should be encoded in the baseDifficulty, and this is a hack to quickly change it for all fill techniques.
	fillMultiplier := 1.0

	if !self.IsFill() {
		fillMultiplier = 5.0
	}

	return groupMultiplier * fillMultiplier * math.Pow(baseDifficulty, float64(self.k))
}

func (self *basicSolveTechnique) getter(grid *Grid) func(int) CellSlice {
	switch self.groupType {
	case _GROUP_ROW:
		return func(i int) CellSlice {
			return grid.Row(i)
		}
	case _GROUP_COL:
		return func(i int) CellSlice {
			return grid.Col(i)
		}
	case _GROUP_BLOCK:
		return func(i int) CellSlice {
			return grid.Block(i)
		}
	default:
		//This should never happen in normal execution--the rare techniques where it doesn't work should never call getter.
		log.Println("Asked for a getter for a function with GROUP_NONE")
		//Return a shell of  a function just to not trip up things downstream.
		return func(i int) CellSlice {
			return nil
		}
	}
}

//This is useful both for hidden and naked subset techniques
func subsetIndexes(len int, size int) [][]int {
	//Given size of array to generate subset for, and size of desired subset, returns an array of all subset-indexes to try.
	//Sanity check
	if size > len {
		return nil
	}

	//returns an array of slices of size size that give you all of the subsets of a list of length len
	result := make([][]int, 0)
	counters := make([]int, size)
	for i, _ := range counters {
		counters[i] = i
	}
	for {
		innerResult := make([]int, size)
		for i, counter := range counters {
			innerResult[i] = counter
		}
		result = append(result, innerResult)
		//Now, increment.
		//Start at the end and try to increment each counter one.
		incremented := false
		for i := size - 1; i >= 0; i-- {

			counter := counters[i]
			if counter < len-(size-i) {
				//Found one!
				counters[i]++
				incremented = true
				if i < size-1 {
					//It was an inner counter; need to set all of the higher counters to one above the one to the left.
					base := counters[i] + 1
					for j := i + 1; j < size; j++ {
						counters[j] = base
						base++
					}
				}
				break
			}
		}
		//If we couldn't increment any, there's nothing to do.
		if !incremented {
			break
		}
	}
	return result
}
