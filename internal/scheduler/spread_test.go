package scheduler

import (
	"testing"
)

func TestComputeGapLower(t *testing.T) {
	// n=3, x=1, W=10: D=1, B=9, g_lower = 1/10 = 0.1h = 6min
	got := ComputeGapLower(1, 9)
	if !floatEq(got, 0.1) {
		t.Errorf("ComputeGapLower(1, 9) = %g, want 0.1", got)
	}
}

func TestComputeGapCycle(t *testing.T) {
	// n=3, x=1, kMax=3: g_cycle = max(0, (5*1+1)/(3*2) - 1) = max(0, 6/6 - 1) = 0
	got := ComputeGapCycle(1, 3, 3)
	if !floatEq(got, 0) {
		t.Errorf("ComputeGapCycle(x=1, n=3, kMax=3) = %g, want 0", got)
	}

	// n=1, x=1, kMax=3: g_cycle = max(0, (5*1+1)/(1*2) - 1) = max(0, 3-1) = 2
	got = ComputeGapCycle(1, 1, 3)
	if !floatEq(got, 2) {
		t.Errorf("ComputeGapCycle(x=1, n=1, kMax=3) = %g, want 2", got)
	}

	// kMax=1: should return 0 (division by zero guard)
	got = ComputeGapCycle(1, 3, 1)
	if !floatEq(got, 0) {
		t.Errorf("ComputeGapCycle(x=1, n=3, kMax=1) = %g, want 0", got)
	}
}

func TestComputeOptimalGap(t *testing.T) {
	// n=3, x=1, W=10: g* = max(0.1, 0) = 0.1
	got := ComputeOptimalGap(1, 1, 3, 3, 9)
	if !floatEq(got, 0.1) {
		t.Errorf("ComputeOptimalGap for n=3 = %g, want 0.1", got)
	}

	// n=1, x=1, W=10: D=7, B=3
	// g_lower = 7/4 = 1.75, g_cycle = 2, g* = 2
	got = ComputeOptimalGap(7, 1, 1, 3, 3)
	if !floatEq(got, 2) {
		t.Errorf("ComputeOptimalGap for n=1 = %g, want 2", got)
	}
}

func TestSpreadSchedule(t *testing.T) {
	p := Params{N: 3, X: 1, S: 10, E: 20, W: 10}
	tt, err := SpreadSchedule(p)
	if err != nil {
		t.Fatal(err)
	}

	if tt.Strategy != "spread" {
		t.Errorf("Strategy = %q, want spread", tt.Strategy)
	}
	if len(tt.Blocks) != 9 {
		t.Errorf("len(Blocks) = %d, want 9", len(tt.Blocks))
	}
	if len(tt.Accounts) != 3 {
		t.Errorf("len(Accounts) = %d, want 3", len(tt.Accounts))
	}

	// Verify round-robin assignment
	for i, b := range tt.Blocks {
		expectedAcct := i % 3
		if b.AccountIndex != expectedAcct {
			t.Errorf("Block %d: AccountIndex = %d, want %d", i, b.AccountIndex, expectedAcct)
		}
	}

	// Verify blocks are within window
	for i, b := range tt.Blocks {
		if b.Start < p.S-eps {
			t.Errorf("Block %d starts before window: %g < %g", i, b.Start, p.S)
		}
		if b.End > p.E+eps {
			t.Errorf("Block %d ends after window: %g > %g", i, b.End, p.E)
		}
	}

	// Verify pre-activation times are before coding window
	for _, acct := range tt.Accounts {
		if acct.PreActivationTime > p.S+eps {
			t.Errorf("Account %d pre-activation %g is after window start %g",
				acct.AccountIndex, acct.PreActivationTime, p.S)
		}
	}
}

func TestSpreadScheduleFullCoverage(t *testing.T) {
	// n=2, x=1.5, W=8: full coverage
	p := Params{N: 2, X: 1.5, S: 10, E: 18, W: 8}
	tt, err := SpreadSchedule(p)
	if err != nil {
		t.Fatal(err)
	}

	if tt.Metrics.D > eps {
		t.Errorf("D = %g, want ~0 (full coverage)", tt.Metrics.D)
	}
}
