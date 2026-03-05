package scheduler

import (
	"testing"
)

func TestComputeKc(t *testing.T) {
	tests := []struct {
		name   string
		x      float64
		n, kMax int
		want   int
	}{
		// n=3, x=1: kc = floor((5-1)/(5-3)) + 1 = floor(2) + 1 = 3
		{"n3_x1", 1, 3, 3, 3},
		// n=2, x=1: kc = floor((5-1)/(5-2)) + 1 = floor(1.33) + 1 = 2
		{"n2_x1", 1, 2, 3, 2},
		// n*x >= T: unlimited (kMax)
		{"unlimited", 1, 5, 3, 3},
		{"unlimited2", 2, 3, 3, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeKc(tt.x, tt.n, tt.kMax)
			if got != tt.want {
				t.Errorf("ComputeKc(x=%g, n=%d, kMax=%d) = %d, want %d",
					tt.x, tt.n, tt.kMax, got, tt.want)
			}
		})
	}
}

func TestComputeCMax(t *testing.T) {
	// n=3, x=1, kc=3, W=10: Cmax = min(10, 9) = 9
	got := ComputeCMax(10, 1, 3, 3)
	if !floatEq(got, 9) {
		t.Errorf("ComputeCMax(10, 1, 3, 3) = %g, want 9", got)
	}
}

func TestBunchSchedule(t *testing.T) {
	p := Params{N: 3, X: 1, S: 10, E: 20, W: 10}
	tt, err := BunchSchedule(p)
	if err != nil {
		t.Fatal(err)
	}

	if tt.Strategy != "bunch" {
		t.Errorf("Strategy = %q, want bunch", tt.Strategy)
	}

	// Should have 9 blocks total
	if len(tt.Blocks) != 9 {
		t.Errorf("len(Blocks) = %d, want 9", len(tt.Blocks))
	}

	// First 9 blocks should be contiguous (bunch of 9)
	for i := 1; i < len(tt.Blocks); i++ {
		gap := tt.Blocks[i].Start - tt.Blocks[i-1].End
		if gap > eps {
			t.Errorf("Gap between blocks %d and %d: %g (expected contiguous)", i-1, i, gap)
		}
	}

	// Total coding should be 9h, cooldown 1h
	if !floatEq(tt.Metrics.L, 9) {
		t.Errorf("L = %g, want 9", tt.Metrics.L)
	}
	if !floatEq(tt.Metrics.D, 1) {
		t.Errorf("D = %g, want 1", tt.Metrics.D)
	}
}

func TestBunchScheduleWithGap(t *testing.T) {
	// n=2, x=1, W=10: kMax=3, kc=2, first bunch = 4h, then gap, then more
	p := Params{N: 2, X: 1, S: 10, E: 20, W: 10}
	tt, err := BunchSchedule(p)
	if err != nil {
		t.Fatal(err)
	}

	if len(tt.Blocks) < 5 {
		t.Errorf("len(Blocks) = %d, want >= 5", len(tt.Blocks))
	}

	// Verify all blocks are within window
	for i, b := range tt.Blocks {
		if b.Start < p.S-eps {
			t.Errorf("Block %d starts before window: %g", i, b.Start)
		}
		if b.End > p.E+eps {
			t.Errorf("Block %d ends after window: %g", i, b.End)
		}
	}
}
