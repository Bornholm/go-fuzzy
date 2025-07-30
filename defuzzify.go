package fuzzy

import "math"

func Centroid(steps int) func(m Membership, min, max float64) float64 {
	return func(m Membership, min, max float64) float64 {
		var (
			num float64
			den float64
		)

		if math.IsInf(min, 0) || math.IsInf(max, 0) || min >= max {
			return 0
		}

		step := math.Max(1.0/float64(steps), (max-min)/float64(steps))

		for x := min; x <= max; x += step {
			y := m.Value(x)
			num += y * x
			den += y
		}

		if den == 0 {
			return (min + max) / 2
		}

		return num / den
	}
}

func MeanOfMaximum(steps int) func(m Membership, min, max float64) float64 {
	return func(m Membership, min, max float64) float64 {
		if math.IsInf(min, 0) || math.IsInf(max, 0) || min >= max {
			return (min + max) / 2
		}

		step := (max - min) / float64(steps)

		maxMembershipValue := 0.0
		for x := min; x <= max; x += step {
			y := m.Value(x)
			if y > maxMembershipValue {
				maxMembershipValue = y
			}
		}

		if maxMembershipValue == 0 {
			return (min + max) / 2
		}

		var maxValues []float64
		const epsilon = 1e-9
		for x := min; x <= max; x += step {
			y := m.Value(x)
			if math.Abs(y-maxMembershipValue) < epsilon {
				maxValues = append(maxValues, x)
			}
		}

		if len(maxValues) == 0 {
			return (min + max) / 2
		}

		sum := 0.0
		for _, v := range maxValues {
			sum += v
		}

		return sum / float64(len(maxValues))
	}
}
