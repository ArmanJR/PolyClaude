package scheduler

import (
	"fmt"
	"math"
)

// CheckFeasibility validates that the parameters allow a valid schedule.
func CheckFeasibility(p Params) error {
	if p.X > T {
		return fmt.Errorf("average dev time (%.2f) exceeds cycle length (%.2f)", p.X, T)
	}
	if p.X > p.W {
		return fmt.Errorf("average dev time (%.2f) exceeds coding window (%.2f)", p.X, p.W)
	}
	if p.X <= 0 {
		return fmt.Errorf("average dev time must be positive, got %.2f", p.X)
	}
	if p.N < 1 {
		return fmt.Errorf("number of accounts must be at least 1, got %d", p.N)
	}
	kmax := ComputeKMax(p.W, p.X)
	if kmax < 1 {
		return fmt.Errorf("no usable cycle fits in the coding window (kMax=%d)", kmax)
	}
	return nil
}

// ComputeKMax returns the maximum usable cycles per account.
// Formula: floor((W + 2(T-x)) / T)
func ComputeKMax(W, x float64) int {
	return int(math.Floor((W + 2*(T-x)) / T))
}

// ComputeB returns the total number of coding blocks.
func ComputeB(n, kMax int) int {
	return n * kMax
}

// ComputeL returns the total coding hours from full cycles only.
func ComputeL(W float64, B int, x float64) float64 {
	return math.Min(W, float64(B)*x)
}

// ComputeD returns the total cooldown hours.
func ComputeD(W, L float64) float64 {
	return math.Max(0, W-L)
}

// ComputeLSingle returns the total coding hours for a single account
// including partial cycle contributions.
// Formula: x + floor((W-x)/T)*x + min(x, (W-x) mod T)
func ComputeLSingle(W, x float64) float64 {
	if W <= x {
		return W
	}
	remaining := W - x
	fullCycles := math.Floor(remaining / T)
	partial := math.Mod(remaining, T)
	return x + fullCycles*x + math.Min(x, partial)
}

// ComputeMetrics computes all core metrics from parameters.
func ComputeMetrics(p Params) CoreMetrics {
	kMax := ComputeKMax(p.W, p.X)
	B := ComputeB(p.N, kMax)

	// Use partial cycle refinement for more accurate L
	lSingle := ComputeLSingle(p.W, p.X)
	L := math.Min(p.W, float64(p.N)*lSingle)
	D := ComputeD(p.W, L)

	return CoreMetrics{
		KMax: kMax,
		B:    B,
		L:    L,
		D:    D,
	}
}
