package sudoku

import (
	"testing"
)

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

	result = randomIndexWithInvertedWeights([]float64{0.0, 0.0, 1.0})
	if result == 2 {
		t.Log("Got the wrong index for inverted weights")
		t.Fail()
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
}