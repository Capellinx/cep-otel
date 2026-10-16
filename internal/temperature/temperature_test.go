package temperature_test

import (
	"testing"

	"github.com/capella/cep-otel/internal/temperature"
)

func TestCelsiusToFahrenheit(t *testing.T) {
	cases := map[float64]float64{0: 32, 28.5: 83.3, 100: 212, -40: -40}
	for c, want := range cases {
		if got := temperature.CelsiusToFahrenheit(c); got != want {
			t.Errorf("CelsiusToFahrenheit(%v) = %v, quer %v", c, got, want)
		}
	}
}

func TestCelsiusToKelvin(t *testing.T) {
	cases := map[float64]float64{0: 273, 28.5: 301.5, 100: 373}
	for c, want := range cases {
		if got := temperature.CelsiusToKelvin(c); got != want {
			t.Errorf("CelsiusToKelvin(%v) = %v, quer %v", c, got, want)
		}
	}
}
