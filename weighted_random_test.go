package sudoku

import (
	"math"
	"math/rand"
	"testing"
)

const _NUM_RUNS_TEST_WEIGHTED_DISTRIBUTION = 1000
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
