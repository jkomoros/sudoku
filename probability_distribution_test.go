package sudoku

import (
	"math"
	"math/rand"
	"testing"
)

const _NUM_RUNS_TEST_WEIGHTED_DISTRIBUTION = 10000
const _ALLOWABLE_DIFF_WEIGHTED_DISTRIBUTION = 0.01

func TestInvertingReallyReallyBigDistribution(t *testing.T) {
	//In practical use, Guess technique's non-inverted value is like absurdly
	//large and leads to NaN. We want to make sure that we handle that case
	//reasonably and that all of those huge numbers go to O, not NaN--but only
	//if there are some "normal" valued things int he distribution.

	crazyDistribution := ProbabilityDistribution{
		1.0,
		10.0,
		100.0,
		1000.0,
		1000000000000000000000.0,
	}

	invertedDistribution := crazyDistribution.invert()

	for i, probabability := range invertedDistribution {
		if math.IsNaN(probabability) {
			t.Error("Index", i, "was NaN")
		}
	}
}

func TestAllInfDistribution(t *testing.T) {
	crazyDistribution := ProbabilityDistribution{
		math.Inf(1),
		math.Inf(1),
		math.Inf(1),
	}

	invertedDistribution := crazyDistribution.invert()

	for i, probability := range invertedDistribution {
		if math.IsNaN(probability) {
			t.Error("Index", i, "was NaN")
		}
		if probability != 0.3333333333333333 {
			t.Error("Got wrong value for index", i, "got", probability, "expected 0.333333")
		}
	}
}

func TestRandomWeightedIndex(t *testing.T) {

	result := ProbabilityDistribution{1.0, 0.0}.RandomIndex()
	if result != 0 {
		t.Log("Got wrong result with random weights")
		t.Fail()
	}
	result = ProbabilityDistribution{0.5, 0.0, 0.5}.RandomIndex()
	if result != 0 && result != 2 {
		t.Log("Didn't get one of two legal weights")
		t.Fail()
	}
	result = ProbabilityDistribution{0.0, 0.0, 1.0}.RandomIndex()
	if result != 2 {
		t.Log("Should have gotten last item in random weights; we didn't")
		t.Fail()
	}

	if (ProbabilityDistribution{1.0, 0.000001}.normalized()) {
		t.Log("thought weights were normalized when they weren't")
		t.Fail()
	}
	if !(ProbabilityDistribution{0.5, 0.25, 0.25}.normalized()) {
		t.Log("Didn't think weights were normalized but they were")
		t.Fail()
	}

	if (ProbabilityDistribution{0.5, -0.25, 0.25}.normalized()) {
		t.Error("A negative weight was considered normal.")
	}

	rand.Seed(1)
	result = ProbabilityDistribution{0.0, 0.0, 1.0}.invert().RandomIndex()
	if result == 2 {
		t.Error("Got the wrong index for inverted weights")
	}

	weightResult := ProbabilityDistribution{2.0, 1.0, 1.0}.normalize()
	if weightResult[0] != 0.5 || weightResult[1] != 0.25 || weightResult[2] != 0.25 {
		t.Log("Nomralized weights came back wrong")
		t.Fail()
	}

	weightResult = ProbabilityDistribution{1.0, 1.0, -0.5}.normalize()
	if weightResult[0] != 0.5 || weightResult[1] != 0.5 || weightResult[2] != 0 {
		t.Error("Normalized weights with a negative came back wrong: ", weightResult)
	}

	weightResult = ProbabilityDistribution{-0.25, -0.5, 0.25}.normalize()
	if weightResult[0] != 0.25 || weightResult[1] != 0 || weightResult[2] != 0.75 {
		t.Error("Normalized weights with two different negative numbers came back wrong: ", weightResult)
	}

	result = ProbabilityDistribution{1.0, 0.0}.RandomIndex()
	if result != 0 {
		t.Log("Got wrong result with random weights")
		t.Fail()
	}
	result = ProbabilityDistribution{5.0, 0.0, 5.0}.RandomIndex()
	if result != 0 && result != 2 {
		t.Log("Didn't get one of two legal weights")
		t.Fail()
	}
	result = ProbabilityDistribution{0.0, 0.0, 5.0}.RandomIndex()
	if result != 2 {
		t.Log("Should have gotten last item in random weights; we didn't")
		t.Fail()
	}
	for i := 0; i < 100; i++ {
		rand.Seed(int64(i))
		result = ProbabilityDistribution{1.0, 10.0, 0.5, -1.0, 0.0, 6.4}.RandomIndex()
		if result == 3 {
			t.Error("Random index with weights picked wrong index with seed ", i)
		}
	}
	for i := 0; i < 100; i++ {
		rand.Seed(int64(i))
		result = ProbabilityDistribution{1.0, 10.0, 0.5, 1.0, 0.0, 6.4, 0.0}.RandomIndex()
		if result == 4 || result == 6 {
			t.Error("Random index with weights that ended in zero picked wrong index with seed ", i)
		}
	}
}

func TestWeightedRandomDistribution(t *testing.T) {

	//We're just going to bother testing randomIndexWithInvertedWeights since that's the one we actually use
	//in HumanSolve.

	type distributionTestCase struct {
		input       ProbabilityDistribution
		expected    ProbabilityDistribution
		description string
	}

	cases := []distributionTestCase{
		{
			ProbabilityDistribution{
				0.0,
				1.0,
				2.0,
			},
			ProbabilityDistribution{
				0.3505,
				0.33333333333,
				0.3162,
			},
			"0 1 2",
		},
		{
			ProbabilityDistribution{
				0.0,
			},
			ProbabilityDistribution{
				1.0,
			},
			"0.0",
		},
		{
			ProbabilityDistribution{
				10.0,
			},
			ProbabilityDistribution{
				1.0,
			},
			"10.0",
		},
		{
			ProbabilityDistribution{
				0.5,
				0.5,
				1.0,
			},
			ProbabilityDistribution{
				0.337,
				0.337,
				0.327,
			},
			"0.5, 0.5, 1.0",
		},
		{
			ProbabilityDistribution{
				1.0,
				100.0,
				0.5,
				-1.0,
				0.0,
				6.4,
			},
			ProbabilityDistribution{
				0.2019,
				0.0012,
				0.2072,
				0.2232,
				0.2119,
				0.1546,
			},
			"1.0, 100.0, 0.5, -1.0, 0.0, 6.4",
		},
		{
			ProbabilityDistribution{
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
			ProbabilityDistribution{
				0.1721,
				0.1721,
				0.1635,
				0.1624,
				0.1638,
				0.1637,
				0.0014,
				0.0011,
				0.0,
			},
			"Many at same weight; exponential incrase",
		},
		//This demonstrates the same problem as the case above, but is more pure
		{
			ProbabilityDistribution{
				0.0,
				1.0,
				2.0,
				4.0,
				8.0,
				16.0,
			},
			ProbabilityDistribution{
				0.2086,
				0.1979,
				0.1891,
				0.1714,
				0.1402,
				0.0928,
			},
			"Straight power of two increase 31",
		},
		{
			ProbabilityDistribution{
				1.0,
				2.0,
				3.0,
				4.0,
				10.0,
				1000.0,
			},
			ProbabilityDistribution{
				0.2291,
				0.219,
				0.2075,
				0.1979,
				0.1465,
				0.0,
			},
			"Small numbers and very big one",
		},
	}

	for _, testCase := range cases {
		randomIndexDistributionHelper(
			t,
			testCase.input,
			testCase.expected,
			testCase.description)
	}

	for _, testCase := range cases {
		distribution := testCase.input.invert()

		for i, num := range distribution {
			if math.Abs(num-testCase.expected[i]) > 0.01 {
				t.Error("Got wrong distribution for", testCase.description, "at", i, "Got", distribution, "Wanted", testCase.expected)
			}
		}
	}

}

func randomIndexDistributionHelper(t *testing.T, input ProbabilityDistribution, expectedDistribution ProbabilityDistribution, testCase string) {

	if len(input) != len(expectedDistribution) {
		t.Fatal("Given differently sized input and expected distribution")
	}

	//collect the results
	results := make([]int, len(expectedDistribution))
	for i := 0; i < _NUM_RUNS_TEST_WEIGHTED_DISTRIBUTION; i++ {
		rand.Seed(int64(i))
		result := input.invert().RandomIndex()
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
