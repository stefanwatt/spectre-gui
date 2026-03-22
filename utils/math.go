package utils

import "math"

func CalculatePercentage(value, total int) float64 {
	if total == 0 {
		return 100.0
	}
	percentage := float64(value) / float64(total) * 10000.0
	return math.Round(percentage) / 100
}
