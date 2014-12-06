package sudoku

import (
	"math"
	"math/rand"
	"testing"
)

const _NUM_RUNS_TEST_WEIGHTED_DISTRIBUTION = 10000
const _ALLOWABLE_DIFF_WEIGHTED_DISTRIBUTION = 0.01

func TestRandomWeightedIndex(t *testing.T) {

	result := randomIndexWithNormalizedWeights([]float64{1.0, 0.0})
	if result != 0 {
		t.Log("Got wrong result with random weights")
		t.Fail()
	}
	result = randomIndexWithNormalizedWeights([]float64{0.5, 0.0, 0.5})
	if result != 0 && result != 2 {
		t.Log("Didn't get one of two legal weights")
		t.Fail()
	}
	result = randomIndexWithNormalizedWeights([]float64{0.0, 0.0, 1.0})
	if result != 2 {
		t.Log("Should have gotten last item in random weights; we didn't")
		t.Fail()
	}

	if weightsNormalized([]float64{1.0, 0.000001}) {
		t.Log("thought weights were normalized when they weren't")
		t.Fail()
	}
	if !weightsNormalized([]float64{0.5, 0.25, 0.25}) {
		t.Log("Didn't think weights were normalized but they were")
		t.Fail()
	}

	if weightsNormalized([]float64{0.5, -0.25, 0.25}) {
		t.Error("A negative weight was considered normal.")
	}

	rand.Seed(1)
	result = randomIndexWithInvertedWeights([]float64{0.0, 0.0, 1.0})
	if result == 2 {
		t.Error("Got the wrong index for inverted weights")
	}

	for i := 0; i < 10; i++ {
		rand.Seed(int64(i))
		result = randomIndexWithInvertedWeights([]float64{0.5, 1.0, 0.0})
		if result == 1 {
			t.Error("RandominzedIndexWithInvertedWeights returned wrong result")
			break
		}
	}

	weightResult := normalizedWeights([]float64{2.0, 1.0, 1.0})
	if weightResult[0] != 0.5 || weightResult[1] != 0.25 || weightResult[2] != 0.25 {
		t.Log("Nomralized weights came back wrong")
		t.Fail()
	}

	weightResult = normalizedWeights([]float64{1.0, 1.0, -0.5})
	if weightResult[0] != 0.5 || weightResult[1] != 0.5 || weightResult[2] != 0 {
		t.Error("Normalized weights with a negative came back wrong: ", weightResult)
	}

	weightResult = normalizedWeights([]float64{-0.25, -0.5, 0.25})
	if weightResult[0] != 0.25 || weightResult[1] != 0 || weightResult[2] != 0.75 {
		t.Error("Normalized weights with two different negative numbers came back wrong: ", weightResult)
	}

	result = randomIndexWithWeights([]float64{1.0, 0.0})
	if result != 0 {
		t.Log("Got wrong result with random weights")
		t.Fail()
	}
	result = randomIndexWithWeights([]float64{5.0, 0.0, 5.0})
	if result != 0 && result != 2 {
		t.Log("Didn't get one of two legal weights")
		t.Fail()
	}
	result = randomIndexWithWeights([]float64{0.0, 0.0, 5.0})
	if result != 2 {
		t.Log("Should have gotten last item in random weights; we didn't")
		t.Fail()
	}
	for i := 0; i < 100; i++ {
		rand.Seed(int64(i))
		result = randomIndexWithWeights([]float64{1.0, 10.0, 0.5, -1.0, 0.0, 6.4})
		if result == 3 {
			t.Error("Random index with weights picked wrong index with seed ", i)
		}
	}
	for i := 0; i < 100; i++ {
		rand.Seed(int64(i))
		result = randomIndexWithWeights([]float64{1.0, 10.0, 0.5, 1.0, 0.0, 6.4, 0.0})
		if result == 4 || result == 6 {
			t.Error("Random index with weights that ended in zero picked wrong index with seed ", i)
		}
	}
}

func TestWeightedRandomDistribution(t *testing.T) {

	//We're just going to bother testing randomIndexWithInvertedWeights since that's the one we actually use
	//in HumanSolve.

	type distributionTestCase struct {
		input       []float64
		expected    []float64
		description string
	}

	cases := []distributionTestCase{
		{
			[]float64{
				0.0,
				1.0,
				2.0,
			},
			[]float64{
				0.66666666666,
				0.33333333333,
				0.0,
			},
			"0 1 2",
		},
		{
			[]float64{
				0.0,
			},
			[]float64{
				1.0,
			},
			"0.0",
		},
		{
			[]float64{
				10.0,
			},
			[]float64{
				1.0,
			},
			"10.0",
		},
		{
			[]float64{
				0.5,
				0.5,
				1.0,
			},
			[]float64{
				0.5,
				0.5,
				0.0,
			},
			"0.5, 0.5, 1.0",
		},
		{
			[]float64{
				1.0,
				100.0,
				0.5,
				-1.0,
				0.0,
				6.4,
			},
			[]float64{
				0.20077063475968,
				0.0,
				0.20178462786453,
				0.20482660717907,
				0.20279862096938,
				0.18981950922734,
			},
			"1.0, 100.0, 0.5, -1.0, 0.0, 6.4",
		},
		{
			[]float64{
				3.0,
				3.0,
				4.0,
				4.0,
				4.0,
				4.0,
				100.0,
				100.0,
				400.0,
			},
			//TODO: think about this hard... This probability seems wrong for what we're trying to accomplish.
			// The precense of the very high numbers makes it so that the early items are picked at a similar
			//rate to the later ones that should be an order of magnitude less likely. This kind of condition
			//happens a lot in HumanSolve (all you need is one high weighted technique), and would explain why
			//we have kind of odd selections of steps.
			[]float64{
				0.13331094694426,
				0.13331094694426,
				0.13297515110813,
				0.13297515110813,
				0.13297515110813,
				0.13297515110813,
				0.10073875083949,
				0.10073875083949,
				0.0,
			},
			"Many at same weight; exponential incrase",
		},
		//This demonstrates the same problem as the case above, but is more pure
		{
			[]float64{
				0.0,
				1.0,
				2.0,
				4.0,
				8.0,
				16.0,
			},
			/*
				//What I'd expect the weights to be:
				//Just the weights of the input non inverted, reversed.
				//It's easier to reason about in this case since each one clearly has a dual; not
				//immediately clear how to generalize it.
				//
				//Based on a lot of thinking, and looking carefully at the graphs, I think the answer might be:
				//* De-negativize
				//* Remove duplicates (for now)
				//* Sort the weights in ascending order.
				//* Set the lowest one to what is the highest number.
				//* Working back from the highest to lowest number, the slope from the lowest to highest should match
				//... so if len = 4, to compute the value of 1, it would be the value of 0 + diff from 4 to 3.
				//... In this step you obviously put back in repeated versions with the same value
				//* Now you normalize THAT.
				//... WAIT, I think this might only work if the numbers are smoothly distributed across the range.
					// To reason about that, consider {2,3,8}
				[]float64{
					0.51612903225806,
					0.25806451612903,
					0.12903225806452,
					0.06451612903226,
					0.03225806451613,
					0.0,
				},
			*/
			[]float64{
				0.24615384615385,
				0.23076923076923,
				0.21538461538462,
				0.18461538461538,
				0.12307692307692,
				0.0,
			},
			"Straight power of two increase 31",
		},
	}

	for _, testCase := range cases {
		randomIndexDistributionHelper(
			t,
			randomIndexWithInvertedWeights,
			testCase.input,
			testCase.expected,
			testCase.description)
	}

}

func randomIndexDistributionHelper(t *testing.T, theFunc func([]float64) int, input []float64, expectedDistribution []float64, testCase string) {

	if len(input) != len(expectedDistribution) {
		t.Fatal("Given differently sized input and expected distribution")
	}

	//collect the results
	results := make([]int, len(expectedDistribution))
	for i := 0; i < _NUM_RUNS_TEST_WEIGHTED_DISTRIBUTION; i++ {
		rand.Seed(int64(i))
		result := theFunc(input)
		results[result]++
	}

	//normalize the results and then calculate the diffs from expected.

	diffAccum := 0.0

	normalizedResults := make([]float64, len(results))

	for i, result := range results {
		normalizedResults[i] = float64(result) / _NUM_RUNS_TEST_WEIGHTED_DISTRIBUTION
		diffAccum += math.Abs(normalizedResults[i] - expectedDistribution[i])
	}

	if diffAccum > _ALLOWABLE_DIFF_WEIGHTED_DISTRIBUTION {
		t.Error("More than allowable difference observed in weighted random distribution:", diffAccum, testCase, "Got", normalizedResults, "Expected", expectedDistribution)
	}

}
