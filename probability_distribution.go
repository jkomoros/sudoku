package sudoku

import (
	"math"
	"math/rand"
)

//ProbabilityDistribution represents a distribution of probabilities over
//indexes.
type ProbabilityDistribution []float64

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

	allInfinite := true
	for _, weight := range d {
		if !math.IsInf(weight, 0) {
			allInfinite = false
			break
		}
	}

	//If all of the items are infinite then we want to special case to spread
	//probability evenly by putting all numbers to lowest possible numbers.
	//However if there is a mix of some infinite and some non-infinite we want
	//to basically ignore the infinite ones.
	if allInfinite {
		for i, _ := range weights {
			weights[i] = math.SmallestNonzeroFloat64
		}
	} else {
		//Invert
		for i, weight := range invertedWeights {
			weights[i] = invertWeight(weight)
		}
	}

	//But now you need to renormalize since they won't sum to 1.
	return weights.normalize()
}

//invertWeight is the primary logic used to invert a positive weight.
func invertWeight(inverted float64) float64 {

	if math.IsInf(inverted, 0) {
		//This will only happen if there's a mix of Inf and non-Inf in the
		//input distribution--in which case the expected beahvior is that
		//Inf's are basically ignore.d
		return 0.0
	}

	result := 1 / math.Exp(inverted/10)

	if math.IsInf(result, 0) {
		result = math.SmallestNonzeroFloat64
	} else if result == 0.0 {
		//Some very large numbers will come out as 0.0, but that's wrong.
		result = math.SmallestNonzeroFloat64
	}
	return result
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
