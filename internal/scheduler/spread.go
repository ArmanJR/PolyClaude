package scheduler

import "math"

// ComputeGapLower returns the information-theoretic lower bound on maximum gap.
// Formula: D / (B+1)
func ComputeGapLower(D float64, B int) float64 {
	if B <= 0 {
		return D
	}
	return D / float64(B+1)
}

// ComputeGapCycle returns the cycle-feasibility bound on the gap.
// Formula: max(0, (T*(kMax-2) + x) / (n*(kMax-1)) - x)
// Guard: if kMax <= 1, return 0 (no consecutive same-account blocks).
func ComputeGapCycle(x float64, n, kMax int) float64 {
	if kMax <= 1 {
		return 0
	}
	numerator := T*float64(kMax-2) + x
	denominator := float64(n) * float64(kMax-1)
	g := numerator/denominator - x
	return math.Max(0, g)
}

// ComputeOptimalGap returns the optimal gap for the spread strategy.
// Formula: g* = max(gLower, gCycle)
func ComputeOptimalGap(D, x float64, n, kMax, B int) float64 {
	gLower := ComputeGapLower(D, B)
	gCycle := ComputeGapCycle(x, n, kMax)
	return math.Max(gLower, gCycle)
}

// ComputeSpreadBlocks places B blocks with equal gaps using round-robin assignment.
// Block i starts at: s + (i+1)*g + i*x
func ComputeSpreadBlocks(p Params, metrics CoreMetrics, gap float64) []Block {
	blocks := make([]Block, metrics.B)
	for i := 0; i < metrics.B; i++ {
		start := p.S + float64(i+1)*gap + float64(i)*p.X
		blocks[i] = Block{
			AccountIndex: i % p.N,
			CycleIndex:   i / p.N,
			Start:        start,
			End:          start + p.X,
		}
	}
	return blocks
}

// SpreadSchedule computes the full spread strategy schedule.
func SpreadSchedule(p Params) (*Timetable, error) {
	if err := CheckFeasibility(p); err != nil {
		return nil, err
	}

	metrics := ComputeMetrics(p)
	gap := ComputeOptimalGap(metrics.D, p.X, p.N, metrics.KMax, metrics.B)
	blocks := ComputeSpreadBlocks(p, metrics, gap)

	accounts := ComputePreActivations(p, blocks)

	return &Timetable{
		Strategy: "spread",
		Metrics:  metrics,
		Accounts: accounts,
		Blocks:   blocks,
	}, nil
}
