package sudoku

import (
	"math"
	"reflect"
	"strconv"
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

func TestHumanSolve(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)

	steps := grid.HumanSolution(nil)

	if steps == nil {
		t.Fatal("Human solution returned 0 techniques.")
	}

	if steps.IsHint {
		t.Error("Steps came back as a hint, not a full solution.")
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

	if len(solution.Steps) != 0 {
		t.Error("A human solve with very limited techniques and no allowed guesses was still solved: ", solution)
	}
}

func TestShortTechniquesToUseHumanSolveOptions(t *testing.T) {

	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)

	shortTechniqueOptions := DefaultHumanSolveOptions()
	shortTechniqueOptions.TechniquesToUse = Techniques[0:5]

	steps := grid.HumanSolution(shortTechniqueOptions)

	if steps == nil {
		t.Fatal("Short technique Options returned nothing")
	}
}

func TestHumanSolveOptionsMethods(t *testing.T) {

	defaultOptions := &HumanSolveOptions{
		15,
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

func TestTechniquesToUseAfterGuessHumanSolveOptions(t *testing.T) {
	grid := NewGrid()
	defer grid.Done()
	grid.LoadSDK(TEST_GRID)

	options := DefaultHumanSolveOptions()
	options.TechniquesToUse = []SolveTechnique{}
	options.techniquesToUseAfterGuess = Techniques[0:5]

	solution := grid.HumanSolution(options)

	steps := solution.Steps

	if len(steps) == 0 {
		t.Fatal("Options with techniques to use after guess returned nil")
	}

	if steps[0].Technique != GuessTechnique {
		t.Error("First technqiu with techniques to use after guess wasn't guess")
	}

	allowedTechniques := make(map[SolveTechnique]bool)

	for _, technique := range Techniques[0:5] {
		allowedTechniques[technique] = true
	}

	//Guess is also allowed to be used later, although we don't expect that.
	allowedTechniques[GuessTechnique] = true

	for i, step := range steps[1:len(steps)] {
		if _, ok := allowedTechniques[step.Technique]; !ok {
			t.Error("Step number", i, "was not in set of allowed techniques", step.Technique)
		}
	}

}

func TestHint(t *testing.T) {

	//This is still flaky, but at least it's a little more likely to catch problems. :-/
	for i := 0; i < 10; i++ {
		hintTestHelper(t, nil, "base case"+strconv.Itoa(i))
	}

	options := DefaultHumanSolveOptions()
	options.TechniquesToUse = []SolveTechnique{}
	options.techniquesToUseAfterGuess = Techniques

	hintTestHelper(t, options, "guess")
}

func hintTestHelper(t *testing.T, options *HumanSolveOptions, description string) {
	grid := NewGrid()
	defer grid.Done()

	grid.LoadSDK(TEST_GRID)

	diagram := grid.Diagram(false)

	hint := grid.Hint(options)

	if grid.Diagram(false) != diagram {
		t.Error("Hint mutated the grid but it wasn't supposed to.")
	}

	steps := hint.Steps

	if steps == nil || len(steps) == 0 {
		t.Error("No steps returned from Hint", description)
	}

	if !hint.IsHint {
		t.Error("Steps was not a hint, but a full solution.")
	}

	for count, step := range steps {
		if count == len(steps)-1 {
			//Last one
			if !step.Technique.IsFill() {
				t.Error("Non-fill step as last step in Hint: ", step.Technique.Name(), description)
			}
		} else {
			//Not last one
			if step.Technique.IsFill() {
				t.Error("Fill step as non-last step in Hint: ", count, step.Technique.Name(), description)
			}
		}
	}
}

func TestHumanSolveWithGuess(t *testing.T) {

	grid := NewGrid()
	defer grid.Done()

	if !grid.LoadSDKFromFile(puzzlePath("harddifficulty.sdk")) {
		t.Fatal("harddifficulty.sdk wasn't loaded")
	}

	solution := grid.HumanSolution(nil)
	steps := solution.Steps

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
		[]*SolveStep{
			&SolveStep{
				techniquesByName["Only Legal Number"],
				CellSlice{
					grid.Cell(0, 0),
				},
				IntSlice{1},
				nil,
				nil,
				nil,
			},
			&SolveStep{
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
			&SolveStep{
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
		false,
	}

	descriptions := steps.Description()

	GOLDEN_DESCRIPTIONS := []string{
		"First, we put 1 in cell (0,0) because 1 is the only remaining valid number for that cell.",
		"Next, we remove the possibilities 1 and 2 from cells (1,0) and (1,1) because 1 is only possible in column 0 of block 1, which means it can't be in any other cell in that column not in that block.",
		"Finally, we put 2 in cell (2,0) because 2 is the only remaining valid number for that cell.",
	}

	for i := 0; i < len(GOLDEN_DESCRIPTIONS); i++ {
		if descriptions[i] != GOLDEN_DESCRIPTIONS[i] {
			t.Log("Got wrong human solve description: ", descriptions[i])
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

func TestTweakChainedStepsWeights(t *testing.T) {

	//TODO: test other, harder cases as well.
	grid := NewGrid()
	lastStep := &SolveStep{
		nil,
		cellRefsToCells([]cellRef{
			{0, 0},
		}, grid),
		nil,
		nil,
		nil,
		nil,
	}
	possibilities := []*SolveStep{
		{
			nil,
			cellRefsToCells([]cellRef{
				{1, 0},
			}, grid),
			nil,
			nil,
			nil,
			nil,
		},
		{
			nil,
			cellRefsToCells([]cellRef{
				{2, 2},
			}, grid),
			nil,
			nil,
			nil,
			nil,
		},
		{
			nil,
			cellRefsToCells([]cellRef{
				{7, 7},
			}, grid),
			nil,
			nil,
			nil,
			nil,
		},
	}

	expected := []float64{
		3.727593720314952e+08,
		517947.4679231202,
		1.0,
	}

	results := tweakChainedStepsWeights(possibilities, lastStep.TargetCells)

	lastWeight := math.MaxFloat64
	for i, weight := range results {
		if weight >= lastWeight {
			t.Error("Tweak Chained Steps Weights didn't tweak things in the right direction: ", results, "at", i)
		}
		lastWeight = weight
	}

	for i, weight := range results {
		if math.Abs(expected[i]-weight) > 0.00001 {
			t.Error("Index", i, "was different than expected. Got", weight, "wanted", expected[i])
		}
	}
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
