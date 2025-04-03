package shared

import "math"

func RoundUp(x float64) int {
	pow := math.Pow(10, float64(0))
	return int(math.Ceil(x*pow) / pow)
}
