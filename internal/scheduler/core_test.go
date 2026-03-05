package scheduler

import (
	"math"
	"testing"
)

const eps = 1e-9

func floatEq(a, b float64) bool {
	return math.Abs(a-b) < eps
}

func TestComputeKMax(t *testing.T) {
	tests := []struct {
		name string
		W, x float64
		want int
	}{
		{"n=3,x=1,W=10", 10, 1, 3},
		{"n=2,x=1.5,W=8", 8, 1.5, 3},
		{"n=1,x=3,W=10", 10, 3, 2},
		{"x=T", 10, 5, 2},
		{"small_window", 2, 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeKMax(tt.W, tt.x)
			if got != tt.want {
				t.Errorf("ComputeKMax(W=%g, x=%g) = %d, want %d", tt.W, tt.x, got, tt.want)
			}
		})
	}
}

func TestComputeB(t *testing.T) {
	if got := ComputeB(3, 3); got != 9 {
		t.Errorf("ComputeB(3, 3) = %d, want 9", got)
	}
	if got := ComputeB(2, 3); got != 6 {
		t.Errorf("ComputeB(2, 3) = %d, want 6", got)
	}
}

func TestComputeL(t *testing.T) {
	// n=3, x=1, W=10: B=9, L=min(10,9)=9
	if got := ComputeL(10, 9, 1); !floatEq(got, 9) {
		t.Errorf("ComputeL(10, 9, 1) = %g, want 9", got)
	}
	// n=2, x=1.5, W=8: B=6, L=min(8,9)=8
	if got := ComputeL(8, 6, 1.5); !floatEq(got, 8) {
		t.Errorf("ComputeL(8, 6, 1.5) = %g, want 8", got)
	}
}

func TestComputeD(t *testing.T) {
	if got := ComputeD(10, 9); !floatEq(got, 1) {
		t.Errorf("ComputeD(10, 9) = %g, want 1", got)
	}
	if got := ComputeD(8, 8); !floatEq(got, 0) {
		t.Errorf("ComputeD(8, 8) = %g, want 0", got)
	}
}

func TestComputeLSingle(t *testing.T) {
	tests := []struct {
		name string
		W, x float64
		want float64
	}{
		// n=1, x=3, W=10: L_single = 3 + 1*3 + min(3,2) = 8
		{"partial_cycle", 10, 3, 8},
		// x=1, W=10: L_single = 1 + 1*1 + min(1,4) = 3 (wait, let me recalculate)
		// L_single = x + floor((W-x)/T)*x + min(x, (W-x) mod T)
		// = 1 + floor(9/5)*1 + min(1, 9 mod 5) = 1 + 1 + min(1,4) = 3
		{"full_coverage_single", 10, 1, 3},
		// W=5, x=1: L_single = 1 + 0*1 + min(1, 4 mod 5) = 1 + 0 + 1 = 2
		// Actually: floor((5-1)/5) = floor(0.8) = 0
		// (W-x) mod T = 4 mod 5 = 4
		// L_single = 1 + 0 + min(1, 4) = 2
		{"one_cycle", 5, 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeLSingle(tt.W, tt.x)
			if !floatEq(got, tt.want) {
				t.Errorf("ComputeLSingle(W=%g, x=%g) = %g, want %g", tt.W, tt.x, got, tt.want)
			}
		})
	}
}

func TestComputeMetrics(t *testing.T) {
	// Example 1: n=3, x=1, W=10
	p := Params{N: 3, X: 1, S: 10, E: 20, W: 10}
	m := ComputeMetrics(p)
	if m.KMax != 3 {
		t.Errorf("KMax = %d, want 3", m.KMax)
	}
	if m.B != 9 {
		t.Errorf("B = %d, want 9", m.B)
	}
	if !floatEq(m.L, 9) {
		t.Errorf("L = %g, want 9", m.L)
	}
	if !floatEq(m.D, 1) {
		t.Errorf("D = %g, want 1", m.D)
	}

	// Example 2: n=2, x=1.5, W=8 → full coverage
	p2 := Params{N: 2, X: 1.5, S: 10, E: 18, W: 8}
	m2 := ComputeMetrics(p2)
	if m2.KMax != 3 {
		t.Errorf("KMax = %d, want 3", m2.KMax)
	}
	if !floatEq(m2.D, 0) {
		t.Errorf("D = %g, want 0", m2.D)
	}
}

func TestCheckFeasibility(t *testing.T) {
	tests := []struct {
		name    string
		p       Params
		wantErr bool
	}{
		{"valid", Params{N: 3, X: 1, S: 10, E: 20, W: 10}, false},
		{"x>T", Params{N: 1, X: 6, S: 10, E: 20, W: 10}, true},
		{"x>W", Params{N: 1, X: 3, S: 10, E: 12, W: 2}, true},
		{"n<1", Params{N: 0, X: 1, S: 10, E: 20, W: 10}, true},
		{"x<=0", Params{N: 1, X: 0, S: 10, E: 20, W: 10}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckFeasibility(tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckFeasibility() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
