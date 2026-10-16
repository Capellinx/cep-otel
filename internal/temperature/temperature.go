package temperature

import "math"

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func CelsiusToFahrenheit(c float64) float64 {
	return round2(c*1.8 + 32)
}

func CelsiusToKelvin(c float64) float64 {
	return round2(c + 273)
}
