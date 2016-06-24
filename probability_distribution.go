package sudoku

import (
	"math"
	"math/rand"
)

//ProbabilityDistribution represents a distribution of probabilities over
//indexes.
type ProbabilityDistribution []float64

//probabiliyDistributionTweak represents tweaks to make to a
//ProbabilityDistribution via tweak(). 1.0 is no effect.
type probabilityDistributionTweak []float64

//Normalized returns true if the distribution is normalized: that is, the
//distribution sums to 1.0
func (d ProbabilityDistribution) normalized() bool {
	var sum float64
	for _, weight := range d {
		if weight < 0 {
			return false
		}
		sum += weight
	}
	return sum == 1.0
}

//Normalize returns a new probability distribution based on this one that is
//normalized.
func (d ProbabilityDistribution) normalize() ProbabilityDistribution {
	if d.normalized() {
		return d
	}
	var sum float64

	fixedWeights := d.denegativize()

	for _, weight := range fixedWeights {
		sum += weight
	}

	result := make(ProbabilityDistribution, len(d))
	for i, weight := range fixedWeights {
		result[i] = weight / sum
	}
	return result
}

//Denegativize returns a new probability distribution like this one, but with
//no negatives. If a negative is found, the entire distribution is shifted up
//by that amount.
func (d ProbabilityDistribution) denegativize() ProbabilityDistribution {
	var lowestNegative float64

	//Check for a negative.
	for _, weight := range d {
		if weight < 0 {
			if weight < lowestNegative {
				lowestNegative = weight
			}
		}
	}

	fixedWeights := make(ProbabilityDistribution, len(d))

	copy(fixedWeights, d)

	if lowestNegative != 0 {
		//Found a negative. Need to fix up all weights.
		for i := 0; i < len(d); i++ {
			fixedWeights[i] = d[i] - lowestNegative
		}

	}

	return fixedWeights
}

//tweak takes an amount to tweak each probability and returns a new
//ProbabilityDistribution where each probability is multiplied by tweak and
//the entire distribution is normalized. Useful for strengthening some
//probabilities and reducing others.
func (d ProbabilityDistribution) tweak(tweak probabilityDistributionTweak) ProbabilityDistribution {
	//sanity check that the tweak amounts are same length
	if len(d) != len(tweak) {
		return d
	}
	result := make(ProbabilityDistribution, len(d))

	for i, num := range d {
		result[i] = num * tweak[i]
	}

	return result.normalize()
}

//invert returns a new probability distribution like this one, but "flipped"
//so low values have a high chance of occurring and high values have a low
//chance. The curve used to invert is expoential.
func (d ProbabilityDistribution) invert() ProbabilityDistribution {
	//TODO: this function means that the worst weighted item will have a weight of 0.0. Isn't that wrong? Maybe it should be +1 to everythign?
	weights := make(ProbabilityDistribution, len(d))

	invertedWeights := d.denegativize()

	//This inversion isn't really an inversion, and it's tied closely to the shape of the weightings we expect to be given in the sudoku problem domain.

	//In sudoku, the weights for different techniques go up exponentially. So when we invert, we want to see a similar exponential shape of the
	//flipped values.

	//We flip with 1/x, and the bottom is an exponent, but softened some.

	//I don't know if this math makes any sense, but in the test distributions the outputs FEEL right.

	//Invert
	for i, weight := range invertedWeights {
		weights[i] = invertWeight(weight)
	}

	//But now you need to renormalize since they won't sum to 1.
	return weights.normalize()
}

//invertWeight is the primary logic used to invert a positive weight.
func invertWeight(inverted float64) float64 {
	return 1 / math.Exp(inverted/20)
}

//RandomIndex returns a random index based on the probability distribution.
func (d ProbabilityDistribution) RandomIndex() int {
	weightsToUse := d
	if !d.normalized() {
		weightsToUse = d.normalize()
	}
	sample := rand.Float64()
	var counter float64
	for i, weight := range weightsToUse {
		counter += weight
		if sample <= counter {
			return i
		}
	}
	//This shouldn't happen if the weights are properly normalized.
	return len(weightsToUse) - 1
}
