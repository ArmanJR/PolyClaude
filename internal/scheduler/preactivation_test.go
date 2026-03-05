package scheduler

import (
	"testing"
)

func TestComputePreActivation(t *testing.T) {
	// Single block at position 10: c0 <= 10, c0 >= 10+1-5 = 6
	// Choose largest: c0 = 10
	got := ComputePreActivation([]float64{10}, 1)
	if !floatEq(got, 10) {
		t.Errorf("Single block: pre-act = %g, want 10", got)
	}

	// Two blocks at 10 and 15 (same account, 5h apart = 1 cycle):
	// c0 <= min(10-0, 15-5) = min(10, 10) = 10
	// c0 >= max(10+1-5, 15+1-10) = max(6, 6) = 6
	// Choose largest: c0 = 10
	got = ComputePreActivation([]float64{10, 15}, 1)
	if !floatEq(got, 10) {
		t.Errorf("Two blocks 5h apart: pre-act = %g, want 10", got)
	}

	// Empty positions
	got = ComputePreActivation([]float64{}, 1)
	if !floatEq(got, 0) {
		t.Errorf("Empty: pre-act = %g, want 0", got)
	}
}

func TestComputePreActivations(t *testing.T) {
	p := Params{N: 3, X: 1, S: 10, E: 20, W: 10}
	blocks := []Block{
		{AccountIndex: 0, Start: 10.1, End: 11.1},
		{AccountIndex: 1, Start: 11.1, End: 12.1},
		{AccountIndex: 2, Start: 12.1, End: 13.1},
		{AccountIndex: 0, Start: 13.1, End: 14.1},
		{AccountIndex: 1, Start: 14.1, End: 15.1},
		{AccountIndex: 2, Start: 15.1, End: 16.1},
	}

	accounts := ComputePreActivations(p, blocks)
	if len(accounts) != 3 {
		t.Fatalf("len(accounts) = %d, want 3", len(accounts))
	}

	// Each account should have 2 blocks
	for _, acct := range accounts {
		if len(acct.Blocks) != 2 {
			t.Errorf("Account %d: len(Blocks) = %d, want 2", acct.AccountIndex, len(acct.Blocks))
		}
	}
}

func TestPreActivationC0Bounds(t *testing.T) {
	// Verify that c0 bounds are satisfied for computed schedules
	p := Params{N: 3, X: 1, S: 10, E: 20, W: 10}
	tt, err := SpreadSchedule(p)
	if err != nil {
		t.Fatal(err)
	}

	for _, acct := range tt.Accounts {
		c0 := acct.PreActivationTime
		for j, block := range acct.Blocks {
			// c0 + j*T <= block.Start (block starts after cycle start)
			cycleStart := c0 + float64(j)*T
			if cycleStart > block.Start+eps {
				t.Errorf("Account %d block %d: cycle start %g > block start %g",
					acct.AccountIndex, j, cycleStart, block.Start)
			}
			// block.End <= c0 + (j+1)*T (block ends before cycle end)
			cycleEnd := c0 + float64(j+1)*T
			if block.End > cycleEnd+eps {
				t.Errorf("Account %d block %d: block end %g > cycle end %g",
					acct.AccountIndex, j, block.End, cycleEnd)
			}
		}
	}
}
