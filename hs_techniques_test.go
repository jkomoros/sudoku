package sudoku

import (
	"fmt"
	"github.com/kr/pretty"
	"log"
	"reflect"
	"testing"
)

func TestTechniquesSorted(t *testing.T) {
	lastLikelihood := 0.0
	for i, technique := range AllTechniques {
		if technique.humanLikelihood(nil) < lastLikelihood {
			t.Fatal("Technique named", technique.Name(), "with index", i, "has a likelihood lower than one of the earlier ones: ", technique.humanLikelihood(nil), lastLikelihood)
		}
		lastLikelihood = technique.humanLikelihood(nil)
	}
}

func TestAllVariantNames(t *testing.T) {
	expected := []string{
		"Obvious In Block",
		"Obvious In Row",
		"Obvious In Col",
		"Necessary In Block",
		"Necessary In Row",
		"Necessary In Col",
		"Only Legal Number",
		"Naked Pair Block",
		"Naked Pair Row",
		"Naked Pair Col",
		"Naked Triple Block",
		"Naked Triple Row",
		"Naked Triple Col",
		"Naked Quad Block",
		"Naked Quad Row",
		"Naked Quad Col",
		"Pointing Pair Row",
		"XWing Row",
		"Pointing Pair Col",
		"XWing Col",
		"Block Block Interactions",
		"XYWing",
		"XYWing (Same Block)",
		"Hidden Pair Block",
		"Hidden Pair Row",
		"Hidden Pair Col",
		"Swordfish Row",
		"Swordfish Col",
		"Hidden Triple Block",
		"Hidden Triple Row",
		"Hidden Triple Col",
		"Forcing Chain (1 steps)",
		"Forcing Chain (2 steps)",
		"Forcing Chain (3 steps)",
		"Forcing Chain (4 steps)",
		"Forcing Chain (5 steps)",
		"Forcing Chain (6 steps)",
		"Hidden Quad Block",
		"Hidden Quad Row",
		"Hidden Quad Col",
		"Guess",
	}
	if !reflect.DeepEqual(expected, AllTechniqueVariants) {
		t.Error("Got wrong technique variants. Expected", expected, "got", AllTechniqueVariants, "\nDifferences:\n", pretty.Diff(expected, AllTechniqueVariants))
	}
}

func TestSubsetIndexes(t *testing.T) {
	result := subsetIndexes(3, 1)
	expectedResult := [][]int{{0}, {1}, {2}}
	subsetIndexHelper(t, result, expectedResult)

	result = subsetIndexes(3, 2)
	expectedResult = [][]int{{0, 1}, {0, 2}, {1, 2}}
	subsetIndexHelper(t, result, expectedResult)

	result = subsetIndexes(5, 3)
	expectedResult = [][]int{{0, 1, 2}, {0, 1, 3}, {0, 1, 4}, {0, 2, 3}, {0, 2, 4}, {0, 3, 4}, {1, 2, 3}, {1, 2, 4}, {1, 3, 4}, {2, 3, 4}}
	subsetIndexHelper(t, result, expectedResult)

	if subsetIndexes(1, 2) != nil {
		t.Log("Subset indexes returned a subset where the length is greater than the len")
		t.Fail()
	}

}

func subsetIndexHelper(t *testing.T, result [][]int, expectedResult [][]int) {
	if len(result) != len(expectedResult) {
		t.Log("subset indexes returned wrong number of results for: ", result, " :", expectedResult)
		t.FailNow()
	}
	for i, item := range result {
		if len(item) != len(expectedResult[0]) {
			t.Log("subset indexes returned a result with wrong numbrer of items ", i, " : ", result, " : ", expectedResult)
			t.FailNow()
		}
		for j, value := range item {
			if value != expectedResult[i][j] {
				t.Log("Subset indexes had wrong number at ", i, ",", j, " : ", result, " : ", expectedResult)
				t.Fail()
			}
		}
	}
}

type multipleValidStepLoopOptions struct {
	targetCells  CellRefSlice
	targetNums   IntSlice
	pointerCells CellRefSlice
	pointerNums  IntSlice
	description  string
	extra        interface{}
	variantName  string
}

//TODO: rename this to fit in with the other test helpers
func multipleValidStepsTestHelper(t *testing.T, puzzleName string, techniqueName string, tests []multipleValidStepLoopOptions) {
	options := solveTechniqueTestHelperOptions{
		checkAllSteps: true,
	}

	grid, solver, steps := humanSolveTechniqueTestHelperStepGenerator(t,
		puzzleName, techniqueName, options)

	options.stepsToCheck.grid = grid
	options.stepsToCheck.solver = solver
	options.stepsToCheck.steps = steps

	//OK, now we'll walk through all of the options in a loop and make sure they all show
	//up in the solve steps.

	for _, test := range tests {

		options.targetCells = test.targetCells
		options.targetNums = test.targetNums
		options.pointerCells = test.pointerCells
		options.pointerNums = test.pointerNums
		options.description = test.description
		options.extra = test.extra
		options.variantName = test.variantName

		humanSolveTechniqueTestHelper(t, puzzleName, techniqueName, options)
	}

	if len(tests) != len(steps) {
		t.Error("We didn't have enough tests for all of the steps that ", techniqueName, " returned. Got", len(tests), "expected", len(steps))
	}
}

func techniqueVariantsTestHelper(t *testing.T, techniqueName string, variantNames ...string) {

	technique, ok := techniquesByName[techniqueName]

	if !ok {
		t.Fatal("Couldn't find technqiue named", techniqueName)
	}

	if len(variantNames) == 0 {
		variantNames = []string{technique.Name()}
	}

	names := technique.Variants()
	if len(names) != len(variantNames) {
		t.Fatal("Didn't receive the right number of variants for", technique.Name(), "Got", len(names), "expected", len(variantNames))
	}
	for i, name := range names {
		goldenName := variantNames[i]
		if name != goldenName {
			t.Error(i, "th variant name for", technique.Name(), "wrong. Got", name, "expected", goldenName)
		}
	}
}

//multiTestWrapper wraps a testing.T and makes it possible to run loops
//where at least one run through the loop must not Error for the whole test
//to pass. Call t.Reset(), and at any time call Passed() to see if t.Error()
//has been called since last reset.
//Or, if looping is false, it's just a passthrough to t.Error.
type loopTest struct {
	t           *testing.T
	looping     bool
	lastMessage string
}

func (l *loopTest) Reset() {
	l.lastMessage = ""
}

func (l *loopTest) Passed() bool {
	return l.lastMessage == ""
}

func (l *loopTest) Error(args ...interface{}) {
	if l.looping == false {
		l.t.Error(args...)
	} else {
		l.lastMessage = fmt.Sprint(args...)
	}
}

type solveTechniqueMatchMode int

const (
	solveTechniqueMatchModeAll solveTechniqueMatchMode = iota
	solveTechniqueMatchModeAny
)

type solveTechniqueTestHelperOptions struct {
	transpose bool
	//Whether the descriptions of cells are a list of legal possible individual values, or must all match.
	matchMode    solveTechniqueMatchMode
	targetCells  CellRefSlice
	pointerCells CellRefSlice
	targetNums   IntSlice
	pointerNums  IntSlice
	targetSame   cellGroupType
	targetGroup  int
	variantName  string
	extra        interface{}
	//If true, will loop over all steps from the technique and see if ANY of them match.
	checkAllSteps bool
	//A way to skip the step generator by provding your own list of steps.
	//Useful if you're going to be do repeated calls to the test helper with the
	//same list of steps.
	stepsToCheck struct {
		grid   MutableGrid
		solver SolveTechnique
		steps  []*SolveStep
	}
	//If description provided, the description MUST match.
	description string
	//If descriptions provided, ONE of the descriptions must match.
	//generally used in conjunction with solveTechniqueMatchModeAny.
	descriptions []string
	debugPrint   bool
}

//TODO: 97473c18633203a6eaa075d968ba77d85ba28390 introduced an error here where we don't return all techniques,
//at least for forcing chains technique.
func getStepsForTechnique(technique SolveTechnique, grid Grid, fetchAll bool) []*SolveStep {

	maxResults := 0
	if !fetchAll {
		maxResults = 1
	}

	return technique.Candidates(grid, maxResults)

}

func humanSolveTechniqueTestHelperStepGenerator(t *testing.T, puzzleName string, techniqueName string, options solveTechniqueTestHelperOptions) (MutableGrid, SolveTechnique, []*SolveStep) {

	var grid MutableGrid

	if options.stepsToCheck.grid != nil {
		grid = options.stepsToCheck.grid
	} else {
		tempGrid, err := MutableLoadSDKFromFile(puzzlePath(puzzleName))
		if err != nil {
			t.Fatal("Couldn't load puzzle ", puzzleName)
		}
		grid = tempGrid
	}

	if options.transpose {
		newGrid := grid.(*mutableGridImpl).transpose()
		grid = newGrid
	}

	solver := techniquesByName[techniqueName]

	if solver == nil {
		t.Fatal("Couldn't find technique object: ", techniqueName)
	}

	steps := getStepsForTechnique(solver, grid, options.checkAllSteps)

	return grid, solver, steps
}

func humanSolveTechniqueTestHelper(t *testing.T, puzzleName string, techniqueName string, options solveTechniqueTestHelperOptions) {
	//TODO: it's weird that you have to pass in puzzleName a second time if you're also passing in options.

	//TODO: test for col and block as well

	var grid Grid
	var solver SolveTechnique
	var steps []*SolveStep

	if options.stepsToCheck.grid != nil {
		grid = options.stepsToCheck.grid
		solver = options.stepsToCheck.solver
		steps = options.stepsToCheck.steps
	} else {
		grid, solver, steps = humanSolveTechniqueTestHelperStepGenerator(t, puzzleName, techniqueName, options)
	}

	//This is so weird... if I don't have this no-op here the compiler warns
	//that grid is declared and not used... despite the fact that it OBVIOUSLY
	//is.
	grid.Cell(0, 0)

	//Check if solveStep is nil here
	if len(steps) == 0 {
		t.Fatal(techniqueName, " didn't find a cell it should have.")
	}

	//Instead of calling error on t, we'll call it on l. If we're not in checkAllSteps mode,
	//l.Error() will be  pass through; otherwise we can interrogate it at any point in the loop.
	l := &loopTest{t: t, looping: options.checkAllSteps}

	for _, step := range steps {

		l.Reset()

		if options.debugPrint {
			log.Println(step)
		}

		variantName := options.variantName

		if options.variantName == "" {
			variantName = techniqueName
		}

		if step.TechniqueVariant() != variantName {
			l.Error("TechniqueVariant name was not what was expected. Got", step.TechniqueVariant(), "expected", variantName)
			continue
		}

		foundVariantNameMatch := false
		for _, variant := range AllTechniqueVariants {
			if variant == step.TechniqueVariant() {
				foundVariantNameMatch = true
				break
			}
		}

		if !foundVariantNameMatch {
			//This is a t.error, because every step should be valid in this way.
			t.Error("Found a variant name that's not in the set: ", step.TechniqueVariant())
		}

		if !reflect.DeepEqual(step.extra, options.extra) {
			l.Error("Extra did not match. Got", step.extra, "expected", options.extra)
			continue
		}

		if options.matchMode == solveTechniqueMatchModeAll {

			//All must match

			if options.targetCells != nil {
				if !step.TargetCells.sameAs(options.targetCells) {
					l.Error(techniqueName, " had the wrong target cells: ", step.TargetCells)
					continue
				}
			}
			if options.pointerCells != nil {
				if !step.PointerCells.sameAs(options.pointerCells) {
					l.Error(techniqueName, " had the wrong pointer cells: ", step.PointerCells)
					continue
				}
			}

			switch options.targetSame {
			case _GROUP_ROW:
				if !step.TargetCells.SameRow() || step.TargetCells.Row() != options.targetGroup {
					l.Error("The target cells in the ", techniqueName, " were wrong row :", step.TargetCells.Row())
					continue
				}
			case _GROUP_BLOCK:
				if !step.TargetCells.SameBlock() || step.TargetCells.Block() != options.targetGroup {
					l.Error("The target cells in the ", techniqueName, " were wrong block :", step.TargetCells.Block())
					continue
				}
			case _GROUP_COL:
				if !step.TargetCells.SameCol() || step.TargetCells.Col() != options.targetGroup {
					l.Error("The target cells in the ", techniqueName, " were wrong col :", step.TargetCells.Col())
					continue
				}
			case _GROUP_NONE:
				//Do nothing
			default:
				l.Error("human solve technique helper error: unsupported group type: ", options.targetSame)
				continue
			}

			if options.targetNums != nil {
				if !step.TargetNums.SameContentAs(options.targetNums) {
					l.Error(techniqueName, " found the wrong numbers: ", step.TargetNums)
					continue
				}
			}

			if options.pointerNums != nil {
				if !step.PointerNums.SameContentAs(options.pointerNums) {
					l.Error(techniqueName, "found the wrong numbers:", step.PointerNums)
					continue
				}
			}
		} else if options.matchMode == solveTechniqueMatchModeAny {

			foundMatch := false

			if !reflect.DeepEqual(step.extra, options.extra) {
				l.Error("Extra did not match. Got", step.extra, "expected", options.extra)
				continue
			}

			if options.targetCells != nil {
				foundMatch = false
				for _, ref := range options.targetCells {
					for _, cell := range step.TargetCells {
						if ref == cell {
							//TODO: break out early
							foundMatch = true
						}
					}
				}
				if !foundMatch {
					l.Error(techniqueName, " had the wrong target cells: ", step.TargetCells)
					continue
				}
			}
			if options.pointerCells != nil {
				l.Error("Pointer cells in match mode any not yet supported.")
				continue
			}

			if options.targetSame != _GROUP_NONE {
				l.Error("Target Same in match mode any not yet supported.")
				continue
			}

			if options.targetNums != nil {
				foundMatch = false
				for _, targetNum := range options.targetNums {
					for _, num := range step.TargetNums {
						if targetNum == num {
							foundMatch = true
							//TODO: break early here.
						}
					}
				}
				if !foundMatch {
					l.Error(techniqueName, " had the wrong target nums: ", step.TargetNums)
					continue
				}
			}

			if options.pointerNums != nil {
				foundMatch = false
				for _, pointerNum := range options.pointerNums {
					for _, num := range step.PointerNums {
						if pointerNum == num {
							foundMatch = true
							//TODO: break early here
						}
					}
				}
				if !foundMatch {
					l.Error(techniqueName, " had the wrong pointer nums: ", step.PointerNums)
					continue
				}
			}
		}

		if options.description != "" {
			//Normalize the step so that the description will be stable for the test.
			step.normalize()
			description := solver.Description(step)
			if description != options.description {
				l.Error("Wrong description for ", techniqueName, ". Got:*", description, "* expected: *", options.description, "*")
				continue
			}
		} else if options.descriptions != nil {
			foundMatch := false
			step.normalize()
			description := solver.Description(step)
			for _, targetDescription := range options.descriptions {
				if description == targetDescription {
					foundMatch = true
				}
			}
			if !foundMatch {
				l.Error("No descriptions matched for ", techniqueName, ". Got:*", description)
				continue
			}
		}

		if options.checkAllSteps && l.Passed() {
			break
		}
	}

	if !l.Passed() {
		t.Error("No cells matched any of the options: ", options)
	}

	//TODO: we should do exhaustive testing of SolveStep application. We used to test it here, but as long as targetCells and targetNums are correct it should be fine.
}
