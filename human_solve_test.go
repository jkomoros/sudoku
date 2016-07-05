package sudoku

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func BenchmarkHumanSolve(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grid := NewGrid()
		defer grid.Done()
		grid.LoadSDK(TEST_GRID)
		grid.HumanSolve(nil)
	}
}

func TestHumanSolveAlmostSolvedGrid(t *testing.T) {
	//Tests human solve on a grid with only one cell left to solve. This is an
	//interesting case in HumanSolve because it triggers ExitCondition #1.

	grid := NewGrid()
	grid.LoadSDK(SOLVED_TEST_GRID)

	cell := grid.Cell(0, 0)

	solvedNumber := cell.Number()

	//Unfill this cell only
	cell.SetNumber(0)

	directions := grid.HumanSolve(nil)

	if directions == nil {
		t.Error("Didn't get directions")
	}

	if !grid.Solved() {
		t.Error("HumanSolve didn't solve the grid")
	}

	if cell.Number() != solvedNumber {
		t.Error("Got wrong number in cell. Got", cell.Number(), "expected", solvedNumber)
	}
}

func TestCompoundSolveStep(t *testing.T) {

	nInRowTechnique := techniquesByName["Necessary In Row"]

	if nInRowTechnique == nil {
		t.Fatal("Couldn't find necessary in row technique")
	}

	simpleFillStep := &SolveStep{
		Technique: nInRowTechnique,
	}

	cullTechnique := techniquesByName["Hidden Quad Block"]

	if cullTechnique == nil {
		t.Fatal("Couldn't find hidden quad block technique")
	}

	cullStep := &SolveStep{
		Technique: cullTechnique,
	}

	compound := &CompoundSolveStep{
		PrecursorSteps: []*SolveStep{
			cullStep,
			cullStep,
		},
		FillStep: simpleFillStep,
	}

	if !compound.valid() {
		t.Error("A valid compound was not thought valid")
	}

	steps := compound.Steps()
	expected := []*SolveStep{
		cullStep,
		cullStep,
		simpleFillStep,
	}

	if !reflect.DeepEqual(steps, expected) {
		t.Error("compound.steps gave wrong result. Got", steps, "expected", expected)
	}

	compound.PrecursorSteps[0] = simpleFillStep

	if compound.valid() {
		t.Error("A compound tep with a fill precursor step was thought valid")
	}

	compound.PrecursorSteps = nil

	if !compound.valid() {
		t.Error("A compound step with no precursor steps was not thought valid")
	}

	compound.FillStep = nil

	if compound.valid() {
		t.Error("A compound step with no fill step was thought valid.")
	}

	createdCompound := newCompoundSolveStep([]*SolveStep{
		cullStep,
		cullStep,
		simpleFillStep,
	})

	if createdCompound == nil {
		t.Error("newCompoundSolveStep failed to create compound step")
	}

	if !createdCompound.valid() {
		t.Error("newCompoundSolveStep created invalid compound step")
	}
}

func TestNewCompoundSolveStep(t *testing.T) {

	fillTechnique := techniquesByName["Necessary In Row"]
	cullTechnique := techniquesByName["Pointing Pair Row"]

	if fillTechnique == nil || cullTechnique == nil {
		t.Fatal("couldn't find the fill or cull steps")
	}

	fillStep := &SolveStep{
		Technique: fillTechnique,
	}

	cullStep := &SolveStep{
		Technique: cullTechnique,
	}

	tests := []struct {
		steps       []*SolveStep
		expected    *CompoundSolveStep
		description string
	}{
		{
			[]*SolveStep{
				fillStep,
			},
			&CompoundSolveStep{
				FillStep: fillStep,
			},
			"Single fill step",
		},
		{
			[]*SolveStep{
				cullStep,
				fillStep,
			},
			&CompoundSolveStep{
				FillStep: fillStep,
				PrecursorSteps: []*SolveStep{
					cullStep,
				},
			},
			"Single cull then single fill",
		},
		{
			[]*SolveStep{
				cullStep,
				cullStep,
				fillStep,
			},
			&CompoundSolveStep{
				FillStep: fillStep,
				PrecursorSteps: []*SolveStep{
					cullStep,
					cullStep,
				},
			},
			"Double cull then single fill",
		},
		{
			[]*SolveStep{
				cullStep,
				cullStep,
			},
			nil,
			"Only cull steps",
		},
	}

	for i, test := range tests {
		result := newCompoundSolveStep(test.steps)
		if !reflect.DeepEqual(result, test.expected) {
			t.Error("Test", i, test.description, "Got", result, "Expected", test.expected)
		}
	}
}

func TestHumanSolve(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)

	steps := grid.HumanSolution(nil)

	if steps == nil {
		t.Fatal("Human solution returned 0 techniques.")
	}

	if grid.Solved() {
		t.Log("Human Solutions mutated the grid.")
		t.Fail()
	}

	steps = grid.HumanSolve(nil)
	//TODO: test to make sure that we use a wealth of different techniques. This will require a cooked random for testing.
	if steps == nil {
		t.Log("Human solve returned 0 techniques")
		t.Fail()
	}
	if !grid.Solved() {
		t.Log("Human solve failed to solve the simple grid.")
		t.Fail()
	}
}

func TestHumanSolveOptionsNoGuess(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)

	options := DefaultHumanSolveOptions()
	options.TechniquesToUse = Techniques[0:3]
	options.NoGuess = true

	solution := grid.HumanSolution(options)

	if solution != nil && len(solution.CompoundSteps) != 0 {
		t.Error("A human solve with very limited techniques and no allowed guesses was still solved: ", solution)
	}
}

func TestShortTechniquesToUseHumanSolveOptions(t *testing.T) {

	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)

	shortTechniqueOptions := DefaultHumanSolveOptions()
	shortTechniqueOptions.TechniquesToUse = Techniques[0:2]

	steps := grid.HumanSolution(shortTechniqueOptions)

	if steps == nil {
		t.Fatal("Short technique Options returned nothing")
	}
}

func TestHumanSolveOptionsMethods(t *testing.T) {

	defaultOptions := &HumanSolveOptions{
		10,
		Techniques,
		false,
		nil,
	}

	options := DefaultHumanSolveOptions()

	if !reflect.DeepEqual(options, defaultOptions) {
		t.Error("defaultOptions came back incorrectly: ", options)
	}

	//Test the case where the user is deliberately trying to specify that no
	//normal techniques should use (and that they should implicitly guess
	//constantly)
	zeroLenTechniquesOptions := DefaultHumanSolveOptions()
	zeroLenTechniquesOptions.TechniquesToUse = []SolveTechnique{}

	zeroLenTechniquesOptions.validate()

	if len(zeroLenTechniquesOptions.TechniquesToUse) != 0 {
		t.Error("Validate treated a deliberate zero-len techniques to use as a nil to be replaced")
	}

	weirdOptions := &HumanSolveOptions{
		-3,
		nil,
		false,
		nil,
	}

	validatedOptions := &HumanSolveOptions{
		1,
		Techniques,
		false,
		nil,
	}

	weirdOptions.validate()

	if !reflect.DeepEqual(weirdOptions, validatedOptions) {
		t.Error("Weird options didn't validate:", weirdOptions, "wanted", validatedOptions)
	}

	guessOptions := DefaultHumanSolveOptions()
	guessOptions.TechniquesToUse = AllTechniques
	guessOptions.validate()

	for i, technique := range guessOptions.TechniquesToUse {
		if technique == GuessTechnique {
			t.Error("Validate didn't remove a guesstechnique (position", i, ")")
		}
	}

	//TODO: verify edge case of single GuessTechnique is fine.

}

func TestHint(t *testing.T) {

	//This is still flaky, but at least it's a little more likely to catch problems. :-/
	for i := 0; i < 10; i++ {
		hintTestHelper(t, nil, "base case"+strconv.Itoa(i))
	}

	options := DefaultHumanSolveOptions()
	options.TechniquesToUse = []SolveTechnique{}

	hintTestHelper(t, options, "guess")
}

func hintTestHelper(t *testing.T, options *HumanSolveOptions, description string) {
	grid := NewGrid()
	defer grid.Done()

	grid.LoadSDK(TEST_GRID)

	diagram := grid.Diagram(false)

	hint := grid.Hint(options, nil)

	if grid.Diagram(false) != diagram {
		t.Error("Hint mutated the grid but it wasn't supposed to.")
	}

	steps := hint.CompoundSteps

	if steps == nil || len(steps) == 0 {
		t.Error("No steps returned from Hint", description)
	}

	if len(steps) != 1 {
		t.Error("Hint was wrong length")
	}

	if !steps[0].valid() {
		t.Error("Hint compound step was invalid")
	}
}

func TestHumanSolveWithGuess(t *testing.T) {

	grid := NewGrid()
	defer grid.Done()

	if !grid.LoadSDKFromFile(puzzlePath("harddifficulty.sdk")) {
		t.Fatal("harddifficulty.sdk wasn't loaded")
	}

	solution := grid.HumanSolution(nil)
	steps := solution.Steps()

	if steps == nil {
		t.Fatal("Didn't find a solution to a grid that should have needed a guess")
	}

	foundGuess := false
	for i, step := range steps {
		if step.Technique.Name() == "Guess" {
			foundGuess = true
		}
		step.Apply(grid)
		if grid.Invalid() {
			t.Fatal("A solution with a guess in it got us into an invalid grid state. step", i)
		}
	}

	if !foundGuess {
		t.Error("Solution that should have used guess didn't have any guess.")
	}

	if !grid.Solved() {
		t.Error("A solution with a guess said it should solve the puzzle, but it didn't.")
	}

}

func TestStepsDescription(t *testing.T) {

	grid := NewGrid()
	defer grid.Done()

	//It's really brittle that we load techniques in this way... it changes every time we add a new early technique!
	steps := SolveDirections{
		grid,
		[]*CompoundSolveStep{
			{
				FillStep: &SolveStep{
					techniquesByName["Only Legal Number"],
					CellSlice{
						grid.Cell(0, 0),
					},
					IntSlice{1},
					nil,
					nil,
					nil,
				},
			},
			{
				PrecursorSteps: []*SolveStep{
					{
						techniquesByName["Pointing Pair Col"],
						CellSlice{
							grid.Cell(1, 0),
							grid.Cell(1, 1),
						},
						IntSlice{1, 2},
						CellSlice{
							grid.Cell(1, 3),
							grid.Cell(1, 4),
						},
						nil,
						nil,
					},
				},
				FillStep: &SolveStep{
					techniquesByName["Only Legal Number"],
					CellSlice{
						grid.Cell(2, 0),
					},
					IntSlice{2},
					nil,
					nil,
					nil,
				},
			},
		},
	}

	descriptions := steps.Description()

	GOLDEN_DESCRIPTIONS := []string{
		"First, based on the other numbers you've entered, (0,0) can only be a 1. How do we know that? We put 1 in cell (0,0) because 1 is the only remaining valid number for that cell.",
		"Finally, based on the other numbers you've entered, (2,0) can only be a 2. How do we know that? We can't fill any cells right away so first we need to cull some possibilities. First, we remove the possibilities 1 and 2 from cells (1,0) and (1,1) because 1 is only possible in column 0 of block 1, which means it can't be in any other cell in that column not in that block. Finally, we put 2 in cell (2,0) because 2 is the only remaining valid number for that cell.",
	}

	if len(descriptions) != len(GOLDEN_DESCRIPTIONS) {
		t.Fatal("Descriptions had too few items. Got\n", strings.Join(descriptions, "***"), "\nwanted\n", strings.Join(GOLDEN_DESCRIPTIONS, "***"))
	}

	for i := 0; i < len(GOLDEN_DESCRIPTIONS); i++ {
		if descriptions[i] != GOLDEN_DESCRIPTIONS[i] {
			t.Log("Got wrong human solve description: ", descriptions[i], "wanted", GOLDEN_DESCRIPTIONS[i])
			t.Fail()
		}
	}
}

//TODO: this is useful. Should we use this in other tests?
func cellRefsToCells(refs []cellRef, grid *Grid) CellSlice {
	var result CellSlice
	for _, ref := range refs {
		result = append(result, ref.Cell(grid))
	}
	return result
}

func TestPuzzleDifficulty(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)

	//We use the cheaper one for testing so it completes faster.
	difficulty := grid.calcluateDifficulty(false)

	if grid.Solved() {
		t.Log("Difficulty shouldn't have changed the underlying grid, but it did.")
		t.Fail()
	}

	if difficulty < 0.0 || difficulty > 1.0 {
		t.Log("The grid's difficulty was outside of allowed bounds.")
		t.Fail()
	}

	puzzleFilenames := []string{"harddifficulty.sdk", "harddifficulty2.sdk"}

	for _, filename := range puzzleFilenames {
		puzzleDifficultyHelper(filename, t)
	}
}

func puzzleDifficultyHelper(filename string, t *testing.T) {
	otherGrid := NewGrid()
	if !otherGrid.LoadSDKFromFile(puzzlePath(filename)) {
		t.Log("Whoops, couldn't load the file to test:", filename)
		t.Fail()
	}

	after := time.After(time.Second * 60)

	done := make(chan bool)

	go func() {
		//We use the cheaper one for testing so it completes faster
		_ = otherGrid.calcluateDifficulty(false)
		done <- true
	}()

	select {
	case <-done:
		//totally fine.
	case <-after:
		//Uh oh.
		t.Log("We never finished solving the hard difficulty puzzle: ", filename)
		t.Fail()
	}
}
