package sudoku

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
	//How difficult a real human would say this technique is. Generally inversely related to how often a real person would pick it. 0.0 to 1.0.
	Difficulty() float64
}

type basicSolveTechnique struct {
	name       string
	isFill     bool
	difficulty float64
}

func init() {

	//TODO: calculate more realistic weights.

	Techniques = []SolveTechnique{
		hiddenSingleInRow{
			basicSolveTechnique{
				//TODO: shouldn't this be "Hidden Single Row" (and likewise for others)
				"Necessary In Row",
				true,
				0.0,
			},
		},
		hiddenSingleInCol{
			basicSolveTechnique{
				"Necessary In Col",
				true,
				0.0,
			},
		},
		hiddenSingleInBlock{
			basicSolveTechnique{
				"Necessary In Block",
				true,
				0.0,
			},
		},
		nakedSingleTechnique{
			basicSolveTechnique{
				//TODO: shouldn't this name be Naked Single for consistency?
				"Only Legal Number",
				true,
				5.0,
			},
		},
		pointingPairRow{
			basicSolveTechnique{
				"Pointing Pair Row",
				false,
				25.0,
			},
		},
		pointingPairCol{
			basicSolveTechnique{
				"Pointing Pair Col",
				false,
				25.0,
			},
		},
		nakedPairCol{
			basicSolveTechnique{
				"Naked Pair Col",
				false,
				75.0,
			},
		},
		nakedPairRow{
			basicSolveTechnique{
				"Naked Pair Row",
				false,
				75.0,
			},
		},
		nakedPairBlock{
			basicSolveTechnique{
				"Naked Pair Block",
				false,
				85.0,
			},
		},
		nakedTripleCol{
			basicSolveTechnique{
				"Naked Triple Col",
				false,
				125.0,
			},
		},
		nakedTripleRow{
			basicSolveTechnique{
				"Naked Triple Row",
				false,
				125.0,
			},
		},
		nakedTripleBlock{
			basicSolveTechnique{
				"Naked Triple Block",
				false,
				140.0,
			},
		},
		hiddenPairRow{
			basicSolveTechnique{
				"Hidden Pair Row",
				false,
				300.0,
			},
		},
		hiddenPairCol{
			basicSolveTechnique{
				"Hidden Pair Col",
				false,
				300.0,
			},
		},
		hiddenPairBlock{
			basicSolveTechnique{
				"Hidden Pair Block",
				false,
				250.0,
			},
		},
	}

	techniquesByName = make(map[string]SolveTechnique)

	for _, technique := range Techniques {
		techniquesByName[technique.Name()] = technique
	}

}

func (self basicSolveTechnique) Name() string {
	return self.name
}

func (self basicSolveTechnique) IsFill() bool {
	return self.isFill
}

func (self basicSolveTechnique) Difficulty() float64 {
	return self.difficulty
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
