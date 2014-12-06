package sudoku

import (
	"math/rand"
)

func weightsNormalized(weights []float64) bool {
	var sum float64
	for _, weight := range weights {
		if weight < 0 {
			return false
		}
		sum += weight
	}
	return sum == 1.0
}

func normalizedWeights(weights []float64) []float64 {
	if weightsNormalized(weights) {
		return weights
	}
	var sum float64
	var lowestNegative float64

	//Check for a negative.
	for _, weight := range weights {
		if weight < 0 {
			if weight < lowestNegative {
				lowestNegative = weight
			}
		}
	}

	fixedWeights := make([]float64, len(weights))

	copy(fixedWeights, weights)

	if lowestNegative != 0 {
		//Found a negative. Need to fix up all weights.
		for i := 0; i < len(weights); i++ {
			fixedWeights[i] = weights[i] - lowestNegative
		}

	}

	for _, weight := range fixedWeights {
		sum += weight
	}

	result := make([]float64, len(weights))
	for i, weight := range fixedWeights {
		result[i] = weight / sum
	}
	return result
}

func randomIndexWithInvertedWeights(invertedWeights []float64) int {
	//TODO: this function means that the worst weighted item will have a weight of 0.0. Isn't that wrong? Maybe it should be +1 to everythign?
	weights := make([]float64, len(invertedWeights))

	//Invert
	for i, weight := range invertedWeights {
		weights[i] = weight * -1
	}

	//But now you need to renormalize since they won't sum to 1.
	reNormalizedWeights := normalizedWeights(weights)

	return randomIndexWithNormalizedWeights(reNormalizedWeights)
}

func randomIndexWithWeights(weights []float64) int {
	//TODO: shouldn't this be called weightedRandomIndex?
	return randomIndexWithNormalizedWeights(normalizedWeights(weights))
}

func randomIndexWithNormalizedWeights(weights []float64) int {
	//assumes that weights is normalized--that is, weights all sum to 1.
	sample := rand.Float64()
	var counter float64
	for i, weight := range weights {
		counter += weight
		if sample <= counter {
			return i
		}
	}
	//This shouldn't happen if the weights are properly normalized.
	return len(weights) - 1
}
