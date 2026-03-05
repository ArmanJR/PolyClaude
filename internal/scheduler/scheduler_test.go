package scheduler

import "testing"

func TestScheduleDispatch(t *testing.T) {
	p := Params{N: 3, X: 1, S: 10, E: 20, W: 10}

	// Spread
	tt, err := Schedule(p, "spread")
	if err != nil {
		t.Fatal(err)
	}
	if tt.Strategy != "spread" {
		t.Errorf("Strategy = %q, want spread", tt.Strategy)
	}

	// Bunch
	tt, err = Schedule(p, "bunch")
	if err != nil {
		t.Fatal(err)
	}
	if tt.Strategy != "bunch" {
		t.Errorf("Strategy = %q, want bunch", tt.Strategy)
	}

	// Unknown
	_, err = Schedule(p, "unknown")
	if err == nil {
		t.Error("Expected error for unknown strategy")
	}
}

func TestScheduleEdgeCases(t *testing.T) {
	// Single account
	p := Params{N: 1, X: 1, S: 10, E: 20, W: 10}
	tt, err := Schedule(p, "spread")
	if err != nil {
		t.Fatal(err)
	}
	if tt.Metrics.KMax != 3 {
		t.Errorf("Single account KMax = %d, want 3", tt.Metrics.KMax)
	}

	// x = T
	p2 := Params{N: 2, X: 5, S: 10, E: 20, W: 10}
	tt2, err := Schedule(p2, "spread")
	if err != nil {
		t.Fatal(err)
	}
	if tt2.Metrics.KMax != 2 {
		t.Errorf("x=T KMax = %d, want 2", tt2.Metrics.KMax)
	}
}
