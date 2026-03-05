package scheduler

import "math"

// ComputePreActivation computes the pre-activation time for a single account.
// Given the account's block start positions and cycle length T:
//
//	c0 >= max_j(p_j + x - (j+1)*T)   (lower bound)
//	c0 <= min_j(p_j - j*T)           (upper bound)
//
// Choose the largest valid c0 to minimize lead time before s.
func ComputePreActivation(positions []float64, x float64) float64 {
	if len(positions) == 0 {
		return 0
	}

	upper := math.Inf(1)
	lower := math.Inf(-1)

	for j, pj := range positions {
		u := pj - float64(j)*T
		if u < upper {
			upper = u
		}
		l := pj + x - float64(j+1)*T
		if l > lower {
			lower = l
		}
	}

	if upper >= lower {
		return upper
	}
	// Fallback: constraints can't all be satisfied, use upper bound
	return upper
}

// ComputePreActivations computes pre-activation times for all accounts
// based on the block assignments.
func ComputePreActivations(p Params, blocks []Block) []AccountSchedule {
	// Group blocks by account
	accountBlocks := make(map[int][]Block)
	accountPositions := make(map[int][]float64)
	for _, b := range blocks {
		accountBlocks[b.AccountIndex] = append(accountBlocks[b.AccountIndex], b)
		accountPositions[b.AccountIndex] = append(accountPositions[b.AccountIndex], b.Start)
	}

	accounts := make([]AccountSchedule, p.N)
	for l := 0; l < p.N; l++ {
		positions := accountPositions[l]
		preAct := ComputePreActivation(positions, p.X)
		accounts[l] = AccountSchedule{
			AccountIndex:      l,
			PreActivationTime: preAct,
			Blocks:            accountBlocks[l],
		}
	}

	return accounts
}
