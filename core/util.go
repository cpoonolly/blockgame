package core

import (
	"math"
)

func f32Abs(num float32) float32 {
	return float32(math.Abs(float64(num)))
}

func f32Round(num float32, decimalPlaces uint16) float32 {
	multiplier := math.Pow10(int(decimalPlaces))
	return float32(math.Round(float64(num)*multiplier) / multiplier)
}

func f32Max(num1, num2 float32) float32 {
	return float32(math.Max(float64(num1), float64(num2)))
}

func f32Min(num1, num2 float32) float32 {
	return float32(math.Min(float64(num1), float64(num2)))
}

func f32LimitBetween(num, min, max float32) float32 {
	return f32Min(f32Max(num, min), max)
}
