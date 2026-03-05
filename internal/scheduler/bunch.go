package scheduler

import "math"

// ComputeKc returns the maximum number of consecutive cycles an account
// can participate in within a round-robin bunch.
// Formula: floor((T-x)/(T-n*x)) + 1 when n*x < T; unlimited (kMax) when n*x >= T.
func ComputeKc(x float64, n, kMax int) int {
	nx := float64(n) * x
	if nx >= T {
		return kMax
	}
	return int(math.Floor((T-x)/(T-nx))) + 1
}

// ComputeCMax returns the maximum continuous coding stretch.
// Formula: min(W, n * kc * x)
func ComputeCMax(W, x float64, n, kc int) float64 {
	return math.Min(W, float64(n)*float64(kc)*x)
}

// ComputeBunchBlocks implements the full bunch algorithm:
// 1. Place first bunch of contiguous round-robin blocks from s
// 2. Compute pre-activation times from first bunch
// 3. Compute cooldown gap from cycle timing
// 4. Repeat for subsequent bunches
func ComputeBunchBlocks(p Params, metrics CoreMetrics) []Block {
	kc := ComputeKc(p.X, p.N, metrics.KMax)
	var allBlocks []Block

	// Track how many cycles each account has used
	accountCyclesUsed := make([]int, p.N)
	totalBlocksPlaced := 0
	bunchStart := p.S

	// Track first-bunch block positions per account for pre-activation computation
	type accountBlockInfo struct {
		positions []float64
	}
	accountInfo := make([]accountBlockInfo, p.N)

	bunchNumber := 0
	for totalBlocksPlaced < metrics.B && bunchStart+p.X <= p.E+1e-9 {
		// Determine how many blocks fit in this bunch
		blocksThisBunch := 0
		for i := 0; i < p.N; i++ {
			remaining := metrics.KMax - accountCyclesUsed[i]
			availableForThisAccount := min(remaining, kc)
			if bunchNumber > 0 {
				// After first bunch, kc for subsequent bunches may be different
				// Use remaining cycles
				availableForThisAccount = remaining
			}
			blocksThisBunch += availableForThisAccount
		}

		// Cap by total blocks remaining and window
		blocksThisBunch = min(blocksThisBunch, metrics.B-totalBlocksPlaced)

		// Place blocks contiguously with round-robin
		var bunchBlocks []Block
		pos := bunchStart
		for len(bunchBlocks) < blocksThisBunch {
			accountIdx := len(bunchBlocks) % p.N
			if accountCyclesUsed[accountIdx] >= metrics.KMax {
				// This account is exhausted, skip
				// This means we can't do perfect round-robin; break
				break
			}
			if pos+p.X > p.E+1e-9 {
				break
			}
			block := Block{
				AccountIndex: accountIdx,
				CycleIndex:   accountCyclesUsed[accountIdx],
				Start:        pos,
				End:          pos + p.X,
			}
			bunchBlocks = append(bunchBlocks, block)
			accountInfo[accountIdx].positions = append(accountInfo[accountIdx].positions, pos)
			accountCyclesUsed[accountIdx]++
			pos += p.X
		}

		if len(bunchBlocks) == 0 {
			break
		}

		allBlocks = append(allBlocks, bunchBlocks...)
		totalBlocksPlaced += len(bunchBlocks)
		bunchEnd := pos

		if totalBlocksPlaced >= metrics.B {
			break
		}

		// Compute when next bunch can start based on cycle timing
		// After the first bunch, each account l has used some cycles.
		// Account l's pre-activation time determines when its next cycle starts.
		// We need to compute pre-activation times from the first bunch to determine the gap.

		// Compute c0 for each account from all blocks so far
		nextAvailable := math.Inf(1)
		for l := 0; l < p.N; l++ {
			if len(accountInfo[l].positions) == 0 {
				continue
			}
			if accountCyclesUsed[l] >= metrics.KMax {
				continue
			}
			// Compute c0 for this account
			c0 := computeC0(accountInfo[l].positions, p.X)
			// Next cycle starts at c0 + accountCyclesUsed[l] * T
			nextCycleStart := c0 + float64(accountCyclesUsed[l])*T
			if nextCycleStart < nextAvailable {
				nextAvailable = nextCycleStart
			}
		}

		if math.IsInf(nextAvailable, 1) {
			break
		}

		bunchStart = math.Max(bunchEnd, nextAvailable)
		bunchNumber++
	}

	return allBlocks
}

// computeC0 computes the first cycle start time from block positions.
// c0 <= min_j(p_j - j*T) and c0 >= max_j(p_j + x - (j+1)*T)
// Returns the largest valid c0 to minimize lead time.
func computeC0(positions []float64, x float64) float64 {
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
	// Choose largest valid c0
	if upper < lower {
		// Constraints can't all be satisfied; use upper as best effort
		return upper
	}
	return upper
}

// BunchSchedule computes the full bunch strategy schedule.
func BunchSchedule(p Params) (*Timetable, error) {
	if err := CheckFeasibility(p); err != nil {
		return nil, err
	}

	metrics := ComputeMetrics(p)
	blocks := ComputeBunchBlocks(p, metrics)

	// Recompute metrics based on actual blocks placed
	actualL := 0.0
	for _, b := range blocks {
		actualL += b.End - b.Start
	}
	metrics.L = math.Min(p.W, actualL)
	metrics.D = p.W - metrics.L
	metrics.B = len(blocks)

	accounts := ComputePreActivations(p, blocks)

	// Fix KMax to reflect actual minimum cycles any account received
	minCycles := metrics.KMax
	for _, acct := range accounts {
		if len(acct.Blocks) < minCycles {
			minCycles = len(acct.Blocks)
		}
	}
	metrics.KMax = minCycles

	// Ensure consistent pre-activation gaps across accounts.
	// Accounts with fewer cycles have looser constraints and may get a
	// shorter gap. Normalize them to match the most-constrained accounts.
	maxCycles := 0
	for _, acct := range accounts {
		if len(acct.Blocks) > maxCycles {
			maxCycles = len(acct.Blocks)
		}
	}
	var referenceGap float64
	for _, acct := range accounts {
		if len(acct.Blocks) == maxCycles && len(acct.Blocks) > 0 {
			referenceGap = acct.Blocks[0].Start - acct.PreActivationTime
			break
		}
	}
	for i, acct := range accounts {
		if len(acct.Blocks) > 0 && len(acct.Blocks) < maxCycles {
			desired := acct.Blocks[0].Start - referenceGap
			valid := true
			for j, block := range acct.Blocks {
				if desired+float64(j)*T > block.Start+1e-9 {
					valid = false
					break
				}
				if block.End > desired+float64(j+1)*T+1e-9 {
					valid = false
					break
				}
			}
			if valid {
				accounts[i].PreActivationTime = desired
			}
		}
	}

	return &Timetable{
		Strategy: "bunch",
		Metrics:  metrics,
		Accounts: accounts,
		Blocks:   blocks,
	}, nil
}
