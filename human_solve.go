package sudoku

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

//The number of solves we should average the signals together for before asking them for their difficulty
//Note: this should be set to the num-solves parameter used to train the currently configured weights.
const _NUM_SOLVES_FOR_DIFFICULTY = 10

//The list of techniques that HumanSolve will use to try to solve the puzzle, with the oddball Guess split out.
var (
	//All of the 'normal' Techniques that will be used to solve the puzzle
	Techniques []SolveTechnique
	//The special GuessTechnique that is used only if no other techniques find options.
	GuessTechnique SolveTechnique
	//Every technique that HumanSolve could ever use, including the oddball Guess technique.
	AllTechniques []SolveTechnique
	//Every variant name for every TechniqueVariant that HumanSolve could ever use.
	AllTechniqueVariants []string
)

//The actual techniques are intialized in hs_techniques.go, and actually defined in hst_*.go files.

//Worst case scenario, how many times we'd call HumanSolve to get a difficulty.
const _MAX_DIFFICULTY_ITERATIONS = 50

//TODO: consider relaxing this even more.
//How close we have to get to the average to feel comfortable our difficulty is converging.
const _DIFFICULTY_CONVERGENCE = 0.005

//SolveDirections is a list of CompoundSolveSteps that, when applied in order
//to its Grid, would cause it to be solved (or, for a hint, would cause it to
//have precisely one more fill step filled).
type SolveDirections struct {
	//A copy of the Grid when the SolveDirections was generated. Grab a
	//reference from SolveDirections.Grid().
	gridSnapshot Grid
	//The list of CompoundSolveSteps that, when applied in order, would cause
	//the SolveDirection's Grid() to be solved.
	CompoundSteps []*CompoundSolveStep
}

//SolveStep is a step to fill in a number in a cell or narrow down the
//possibilities in a cell to get it closer to being solved. SolveSteps model
//techniques that humans would use to solve a puzzle. Most HumanSolve related
//methods return CompoundSolveSteps, which are higher-level collections of the
//base SolveSteps.
type SolveStep struct {
	//The technique that was used to identify that this step is logically valid at this point in the solution.
	Technique SolveTechnique
	//The cells that will be affected by the techinque (either the number to fill in or possibilities to exclude).
	TargetCells CellSlice
	//The numbers we will remove (or, in the case of Fill, add) to the TargetCells.
	TargetNums IntSlice
	//The cells that together lead the techinque to logically apply in this case; the cells behind the reasoning
	//why the TargetCells will be mutated in the way specified by this SolveStep.
	PointerCells CellSlice
	//The specific numbers in PointerCells that lead us to remove TargetNums from TargetCells.
	//This is only very rarely needed (at this time only for hiddenSubset techniques)
	PointerNums IntSlice
	//extra is a private place that information relevant to only specific techniques
	//can be stashed.
	extra interface{}
}

//CompoundSolveStep is a special type of meta SolveStep that has 0 to n
//precursor steps that cull possibilities (instead of filling in a number),
//followed by precisely one fill step. It reflects the notion that logically
//only fill steps actually advance the grid towards being solved, and all cull
//steps are in service of getting the grid to a state where a Fill step can be
//found. CompoundSolveSteps are the primary units returned from
//HumanSolutions.
type CompoundSolveStep struct {
	PrecursorSteps []*SolveStep
	FillStep       *SolveStep
	//explanation is an optional string describing extra stuff about the reasoning.
	explanation []string
}

//TODO: consider passing a non-pointer humanSolveOptions so that mutations
//deeper  in the solve stack don' tmatter.

//HumanSolveOptions configures how precisely the human solver should operate.
//Passing nil where a HumanSolveOptions is expected will use reasonable
//defaults. Note that the various human solve methods may mutate your options
//object.
type HumanSolveOptions struct {
	//At each step in solving the puzzle, how many candidate SolveSteps should
	//we generate before stopping the search for more? Higher values will give
	//more 'realistic' solves, but at the cost of *much* higher performance
	//costs. Also note that the difficulty may be wrong if the difficulty
	//model in use was trained on a different NumOptionsToCalculate.
	NumOptionsToCalculate int
	//Which techniques to try at each step of the puzzle, sorted in the order
	//to try them out (generally from cheapest to most expensive). A value of
	//nil will use Techniques (the default). Any GuessTechniques will be
	//ignored.
	TechniquesToUse []SolveTechnique
	//NoGuess specifies that even if no other techniques work, the HumanSolve
	//should not fall back on guessing, and instead just return failure.
	NoGuess bool

	//TODO: figure out how to test that we do indeed use different values of
	//numOptionsToCalculate.
	//TODO: add a TwiddleChainDissimilarity bool.
	cachedEffectiveTechniques []SolveTechnique
}

//DefaultHumanSolveOptions returns a HumanSolveOptions object configured to
//have reasonable defaults.
func DefaultHumanSolveOptions() *HumanSolveOptions {
	result := &HumanSolveOptions{}

	result.NumOptionsToCalculate = 10
	result.TechniquesToUse = Techniques
	result.NoGuess = false

	//Have to set even zero valued properties, because the Options isn't
	//necessarily default initalized.

	return result

}

//Grid returns a snapshot of the grid at the time this SolveDirections was
//generated.
func (self SolveDirections) Grid() Grid {
	//TODO: this is the only pointer receiver method on SolveDirections.
	return self.gridSnapshot
}

//Steps returns the list of all CompoundSolveSteps flattened into one stream
//of SolveSteps.
func (s SolveDirections) Steps() []*SolveStep {
	var result []*SolveStep
	for _, compound := range s.CompoundSteps {
		result = append(result, compound.Steps()...)
	}
	return result
}

//Modifies the options object to make sure all of the options are set
//in a legal way. Returns itself for convenience.
func (self *HumanSolveOptions) validate() *HumanSolveOptions {

	if self.TechniquesToUse == nil {
		self.TechniquesToUse = Techniques
	}

	if self.NumOptionsToCalculate < 1 {
		self.NumOptionsToCalculate = 1
	}

	//Remove any GuessTechniques that might be in there because
	//the are invalid.
	var techniques []SolveTechnique

	for _, technique := range self.TechniquesToUse {
		if technique == GuessTechnique {
			continue
		}
		techniques = append(techniques, technique)
	}

	self.TechniquesToUse = techniques

	return self

}

//effectiveTechniquesToUse returns the effective list of techniques to use.
//Basically just o.TechniquesToUse + Guess if NoGuess is not provided.
func (o *HumanSolveOptions) effectiveTechniquesToUse() []SolveTechnique {
	if o.cachedEffectiveTechniques == nil {
		if o.NoGuess {
			return o.TechniquesToUse
		}
		o.cachedEffectiveTechniques = append(o.TechniquesToUse, GuessTechnique)
	}
	return o.cachedEffectiveTechniques
}

//IsUseful returns true if this SolveStep, when applied to the given grid, would do useful work--that is, it would
//either fill a previously unfilled number, or cull previously un-culled possibilities. This is useful to ensure
//HumanSolve doesn't get in a loop of applying the same useless steps.
func (self *SolveStep) IsUseful(grid Grid) bool {
	//Returns true IFF calling Apply with this step and the given grid would result in some useful work. Does not modify the gri.d

	//All of this logic is substantially recreated in Apply.

	if self.Technique == nil {
		return false
	}

	//TODO: test this.
	if self.Technique.IsFill() {
		if len(self.TargetCells) == 0 || len(self.TargetNums) == 0 {
			return false
		}
		cell := self.TargetCells[0].InGrid(grid)
		return self.TargetNums[0] != cell.Number()
	} else {
		useful := false
		for _, cell := range self.TargetCells {
			gridCell := cell.InGrid(grid)
			for _, exclude := range self.TargetNums {
				//It's right to use Possible because it includes the logic of "it's not possible if there's a number in there already"
				//TODO: ensure the comment above is correct logically.
				if gridCell.Possible(exclude) {
					useful = true
				}
			}
		}
		return useful
	}
}

//Apply does the solve operation to the Grid that is defined by the configuration of the SolveStep, mutating the
//grid and bringing it one step closer to being solved.
func (self *SolveStep) Apply(grid MutableGrid) {
	//All of this logic is substantially recreated in IsUseful.
	if self.Technique.IsFill() {
		if len(self.TargetCells) == 0 || len(self.TargetNums) == 0 {
			return
		}
		cell := self.TargetCells[0].MutableInGrid(grid)
		cell.SetNumber(self.TargetNums[0])
	} else {
		for _, cell := range self.TargetCells {
			gridCell := cell.MutableInGrid(grid)
			for _, exclude := range self.TargetNums {
				gridCell.SetExcluded(exclude, true)
			}
		}
	}
}

//Modifications returns the GridModifications repesenting how this SolveStep
//would mutate the grid.
func (self *SolveStep) Modifications() GridModification {
	var result GridModification

	for _, cell := range self.TargetCells {
		modification := newCellModification(cell)
		if self.Technique.IsFill() {
			if len(self.TargetNums) != 1 {
				//Sanity check
				continue
			}
			modification.Number = self.TargetNums[0]
		} else {
			for _, num := range self.TargetNums {
				modification.ExcludesChanges[num] = true
			}
		}
		result = append(result, modification)
	}
	return result
}

//Description returns a human-readable sentence describing what the SolveStep
//instructs the user to do, and what reasoning it used to decide that this
//step was logically valid to apply.
func (self *SolveStep) Description() string {
	result := ""
	if self.Technique.IsFill() {
		result += fmt.Sprintf("We put %s in cell %s ", self.TargetNums.Description(), self.TargetCells.Description())
	} else {
		//TODO: pluralize based on length of lists.
		result += fmt.Sprintf("We remove the possibilities %s from cells %s ", self.TargetNums.Description(), self.TargetCells.Description())
	}
	result += "because " + self.Technique.Description(self) + "."
	return result
}

//HumanLikelihood is how likely a user would be to pick this step when compared with other possible steps.
//Generally inversely related to difficulty (but not perfectly).
//This value will be used to pick which technique to apply when compared with other candidates.
//Based on the technique's HumanLikelihood, possibly attenuated by this particular step's variant
//or specifics.
func (self *SolveStep) HumanLikelihood() float64 {
	//TODO: attenuate by variant
	return self.Technique.humanLikelihood(self)
}

//TechniqueVariant returns the name of the precise variant of the Technique
//that this step represents. This information is useful for figuring out
//which weight to apply when calculating overall difficulty. A Technique would have
//variants (as opposed to simply other Techniques) when the work to calculate all
//variants is the same, but the difficulty of produced steps may vary due to some
//property of the technique. Forcing Chains is the canonical example.
func (self *SolveStep) TechniqueVariant() string {
	//Defer to the Technique.variant implementation entirely.
	//This allows us to most easily share code for the simple case.
	return self.Technique.variant(self)
}

//normalize puts the step in a known, deterministic state, which eases testing.
func (self *SolveStep) normalize() {
	//Different techniques will want to normalize steps in different ways.
	self.Technique.normalizeStep(self)
}

//newCompoundSolveStep will create a CompoundSolveStep from a series of
//SolveSteps, along as that series is a valid CompoundSolveStep.
func newCompoundSolveStep(steps []*SolveStep) *CompoundSolveStep {
	var result *CompoundSolveStep

	if len(steps) < 1 {
		return nil
	} else if len(steps) == 1 {
		result = &CompoundSolveStep{
			FillStep: steps[0],
		}
	} else {
		result = &CompoundSolveStep{
			PrecursorSteps: steps[0 : len(steps)-1],
			FillStep:       steps[len(steps)-1],
		}
	}
	if result.valid() {
		return result
	}
	return nil
}

//valid returns true iff there are 0 or more cull-steps in PrecursorSteps and
//a non-nill Fill step.
func (c *CompoundSolveStep) valid() bool {
	if c.FillStep == nil {
		return false
	}
	if !c.FillStep.Technique.IsFill() {
		return false
	}
	for _, step := range c.PrecursorSteps {
		if step.Technique.IsFill() {
			return false
		}
	}
	return true
}

//Apply applies all of the steps in the CompoundSolveStep to the grid in
//order: first each of the PrecursorSteps in order, then the fill step. It is
//equivalent to calling Apply() on every step returned by Steps().
func (c *CompoundSolveStep) Apply(grid MutableGrid) {
	//TODO: test this
	if !c.valid() {
		return
	}
	for _, step := range c.PrecursorSteps {
		step.Apply(grid)
	}
	c.FillStep.Apply(grid)
}

//Modifications returns the set of modifications that this CompoundSolveStep
//would make to a Grid if Apply were called.
func (c *CompoundSolveStep) Modifications() GridModification {
	var result GridModification
	for _, step := range c.PrecursorSteps {
		result = append(result, step.Modifications()...)
	}
	result = append(result, c.FillStep.Modifications()...)
	return result
}

//Description returns a human-readable sentence describing what the CompoundSolveStep
//instructs the user to do, and what reasoning it used to decide that this
//step was logically valid to apply.
func (c *CompoundSolveStep) Description() string {
	//TODO: this terminology is too tuned for the Online Sudoku use case.
	//it practice it should probably name the cell in text.

	if c.FillStep == nil || c.FillStep.TargetCells == nil {
		return ""
	}

	var result []string
	result = append(result, "Based on the other numbers you've entered, "+c.FillStep.TargetCells[0].ref().String()+" can only be a "+strconv.Itoa(c.FillStep.TargetNums[0])+".")
	result = append(result, "How do we know that?")
	if len(c.PrecursorSteps) > 0 {
		result = append(result, "We can't fill any cells right away so first we need to cull some possibilities.")
	}
	steps := c.Steps()
	for i, step := range steps {
		intro := ""
		description := step.Description()
		if len(steps) > 1 {
			description = strings.ToLower(description)
			switch i {
			case 0:
				intro = "First, "
			case len(steps) - 1:
				intro = "Finally, "
			default:
				//TODO: switch between "then" and "next" randomly.
				intro = "Next, "
			}
		}
		result = append(result, intro+description)
	}
	return strings.Join(result, " ")
}

//ScoreExplanation returns a in-depth descriptive string about why this
//compound step was scored the way it was. Intended for debugging purposes;
//its primary use is in i-sudoku.
func (c *CompoundSolveStep) ScoreExplanation() []string {
	return c.explanation
}

//Steps returns the simple list of SolveSteps that this CompoundSolveStep represents.
func (c *CompoundSolveStep) Steps() []*SolveStep {
	return append(c.PrecursorSteps, c.FillStep)
}

func (self *gridImpl) HumanSolution(options *HumanSolveOptions) *SolveDirections {
	return humanSolveHelper(self, options, nil, true)

}

func (self *mutableGridImpl) HumanSolution(options *HumanSolveOptions) *SolveDirections {
	clone := self.MutableCopy()
	return clone.HumanSolve(options)
}

func (self *mutableGridImpl) HumanSolve(options *HumanSolveOptions) *SolveDirections {
	return humanSolveHelper(self, options, nil, true)
}

func (self *gridImpl) Hint(options *HumanSolveOptions, optionalPreviousSteps []*CompoundSolveStep) *SolveDirections {

	return humanSolveHelper(self, options, optionalPreviousSteps, false)
}

func (self *mutableGridImpl) Hint(options *HumanSolveOptions, optionalPreviousSteps []*CompoundSolveStep) *SolveDirections {

	//TODO: test that non-fill steps before the last one are necessary to unlock
	//the fill step at the end (cull them if not), and test that.

	clone := self.MutableCopy()

	result := humanSolveHelper(clone, options, optionalPreviousSteps, false)

	return result

}

func (self *gridImpl) Difficulty() float64 {
	return calcluateGridDifficulty(self, true)
}

func (self *mutableGridImpl) Difficulty() float64 {

	//TODO: test that the memoization works (that is, the cached value is thrown out if the grid is modified)
	//It's hard to test because self.calculateDifficulty(true) is so expensive to run.

	//This is so expensive and during testing we don't care if converges.
	//So we split out the meat of the method separately.

	if self == nil {
		return 0.0
	}

	//Yes, this memoization will fail in the (rare!) cases where a grid's actual difficulty is 0.0, but
	//the worst case scenario is that we just return the same value.
	if self.cachedDifficulty == 0.0 {
		self.cachedDifficulty = calcluateGridDifficulty(self, true)
	}
	return self.cachedDifficulty
}

func calcluateGridDifficulty(grid Grid, accurate bool) float64 {
	//This can be an extremely expensive method. Do not call repeatedly!
	//returns the difficulty of the grid, which is a number between 0.0 and 1.0.
	//This is a probabilistic measure; repeated calls may return different numbers, although generally we wait for the results to converge.

	//We solve the same puzzle N times, then ask each set of steps for their difficulty, and combine those to come up with the overall difficulty.

	accum := 0.0
	average := 0.0
	lastAverage := 0.0

	grid.HasMultipleSolutions()

	//Since this is so expensive, in testing situations we want to run it in less accurate mode (so it goes fast!)
	maxIterations := _MAX_DIFFICULTY_ITERATIONS
	if !accurate {
		maxIterations = 1
	}

	for i := 0; i < maxIterations; i++ {
		difficulty := gridDifficultyHelper(grid)

		accum += difficulty
		average = accum / (float64(i) + 1.0)

		if math.Abs(average-lastAverage) < _DIFFICULTY_CONVERGENCE {
			//Okay, we've already converged. Just return early!
			return average
		}

		lastAverage = average
	}

	//We weren't converging... oh well!
	return average
}

//This function will HumanSolve _NUM_SOLVES_FOR_DIFFICULTY times, then average the signals together, then
//give the difficulty for THAT. This is more accurate becuase the weights were trained on such averaged signals.
func gridDifficultyHelper(grid Grid) float64 {

	collector := make(chan DifficultySignals, _NUM_SOLVES_FOR_DIFFICULTY)
	//Might as well run all of the human solutions in parallel
	for i := 0; i < _NUM_SOLVES_FOR_DIFFICULTY; i++ {
		go func(gridToUse Grid) {
			solution := gridToUse.HumanSolution(nil)
			if solution == nil {
				log.Println("A generated grid turned out to have mutiple solutions (or otherwise return nil), indicating a very serious error:", gridToUse.DataString())
				os.Exit(1)
			}
			collector <- solution.Signals()
		}(grid)
	}

	combinedSignals := DifficultySignals{}

	for i := 0; i < _NUM_SOLVES_FOR_DIFFICULTY; i++ {
		signals := <-collector
		combinedSignals.sum(signals)
	}

	//Now average all of the signal values
	for key := range combinedSignals {
		combinedSignals[key] /= _NUM_SOLVES_FOR_DIFFICULTY
	}

	return combinedSignals.difficulty()

}
