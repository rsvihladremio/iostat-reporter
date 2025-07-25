package reporter

import "math"

// CalcScale returns the min, max, and evenly‐spaced interval (over splits)
// across all provided float64 slices.  If data is flat, it expands max = min+1.
func CalcScale(splits int, arrs ...[]float64) (min, max, interval float64) {
	first := true
	for _, arr := range arrs {
		for _, v := range arr {
			if first {
				min, max = v, v
				first = false
			} else {
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
			}
		}
	}
	if first {
		// no data
		return 0, 0, 0
	}
	if max == min {
		max = min + 1
	}
	interval = (max - min) / float64(splits)
	min = math.Round(min*100) / 100
	max = math.Round(max*100) / 100
	interval = math.Round(interval*100) / 100
	return
}
