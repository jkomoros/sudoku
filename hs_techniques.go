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

type SolveTechnique interface {
	Name() string
	Description(*SolveStep) string
	//IMPORTANT: a step should return a step IFF that step is valid AND the step would cause useful work to be done if applied.

	//Find returns as many steps as it can find in the grid for that technique. This helps ensure that when we pick a step,
	//it's more likely to be an "easy" step because there will be more of them at any time.
	Find(*Grid) []*SolveStep
	IsFill() bool
	//How likely a user would be to pick this technique. Generally inversely related to difficulty (but not perfectly).
	//This value will be used to pick which technique to apply.
	//The value is inversely related to how often it will be picked.
	HumanLikelihood() float64
}

type cellGroupType int

const (
	GROUP_NONE = iota
	GROUP_ROW
	GROUP_COL
	GROUP_BLOCK
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
				GROUP_ROW,
				1,
			},
		},
		&hiddenSingleTechnique{
			&basicSolveTechnique{
				"Necessary In Col",
				true,
				GROUP_COL,
				1,
			},
		},
		&hiddenSingleTechnique{
			&basicSolveTechnique{
				"Necessary In Block",
				true,
				GROUP_BLOCK,
				1,
			},
		},
		&nakedSingleTechnique{
			&basicSolveTechnique{
				//TODO: shouldn't this name be Naked Single for consistency?
				"Only Legal Number",
				true,
				GROUP_NONE,
				1,
			},
		},
		&obviousInCollectionTechnique{
			&basicSolveTechnique{
				"Obvious In Row",
				true,
				GROUP_ROW,
				1,
			},
		},
		&obviousInCollectionTechnique{
			&basicSolveTechnique{
				"Obvious In Col",
				true,
				GROUP_COL,
				1,
			},
		},
		&obviousInCollectionTechnique{
			&basicSolveTechnique{
				"Obvious In Block",
				true,
				GROUP_BLOCK,
				1,
			},
		},
		&pointingPairTechnique{
			&basicSolveTechnique{
				"Pointing Pair Row",
				false,
				GROUP_ROW,
				2,
			},
		},
		&pointingPairTechnique{
			&basicSolveTechnique{
				"Pointing Pair Col",
				false,
				GROUP_COL,
				2,
			},
		},
		&blockBlockInteractionTechnique{
			&basicSolveTechnique{
				"Block Block Interactions",
				false,
				GROUP_BLOCK,
				2,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Pair Col",
				false,
				GROUP_COL,
				2,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Pair Row",
				false,
				GROUP_ROW,
				2,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Pair Block",
				false,
				GROUP_BLOCK,
				2,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Triple Col",
				false,
				GROUP_COL,
				3,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Triple Row",
				false,
				GROUP_ROW,
				3,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Triple Block",
				false,
				GROUP_BLOCK,
				3,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Quad Col",
				false,
				GROUP_COL,
				4,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Quad Row",
				false,
				GROUP_ROW,
				4,
			},
		},
		&nakedSubsetTechnique{
			&basicSolveTechnique{
				"Naked Quad Block",
				false,
				GROUP_BLOCK,
				4,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Pair Row",
				false,
				GROUP_ROW,
				2,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Pair Col",
				false,
				GROUP_COL,
				2,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Pair Block",
				false,
				GROUP_BLOCK,
				2,
			},
		},
		&xwingTechnique{
			&basicSolveTechnique{
				"XWing Row",
				false,
				GROUP_ROW,
				2,
			},
		},
		&xwingTechnique{
			&basicSolveTechnique{
				"XWing Col",
				false,
				GROUP_COL,
				2,
			},
		},
	}

	ExpensiveTechniques = []SolveTechnique{
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Triple Row",
				false,
				GROUP_ROW,
				3,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Triple Col",
				false,
				GROUP_COL,
				3,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Triple Block",
				false,
				GROUP_BLOCK,
				3,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Quad Row",
				false,
				GROUP_ROW,
				4,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Quad Col",
				false,
				GROUP_COL,
				4,
			},
		},
		&hiddenSubsetTechnique{
			&basicSolveTechnique{
				"Hidden Quad Block",
				false,
				GROUP_BLOCK,
				4,
			},
		},
	}

	GuessTechnique = &guessTechnique{
		&basicSolveTechnique{
			"Guess",
			true,
			GROUP_NONE,
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
	case GROUP_ROW:
		//Rows are easier to scan because most humans are used to reading LTR
		groupMultiplier = 1.0
	case GROUP_COL:
		//Cols are easy to scan because the eye can move in one line
		groupMultiplier = 1.05
	case GROUP_BLOCK:
		//Blocks are harder to notice because the eye has to zag at least twice.
		groupMultiplier = 1.5
	}

	return groupMultiplier * math.Pow(baseDifficulty, float64(self.k))
}

func (self *basicSolveTechnique) getter(grid *Grid) func(int) CellList {
	switch self.groupType {
	case GROUP_ROW:
		return func(i int) CellList {
			return grid.Row(i)
		}
	case GROUP_COL:
		return func(i int) CellList {
			return grid.Col(i)
		}
	case GROUP_BLOCK:
		return func(i int) CellList {
			return grid.Block(i)
		}
	default:
		//This should never happen in normal execution--the rare techniques where it doesn't work should never call getter.
		log.Println("Asked for a getter for a function with GROUP_NONE")
		//Return a shell of  a function just to not trip up things downstream.
		return func(i int) CellList {
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
