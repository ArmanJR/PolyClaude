# PolyClaude: Optimal Scheduling for Multiple Claude Code Accounts

## Intro

Claude Code offers three subscription tiers: **Pro** ($20/mo), **Max 5x** ($100/mo), and **Max 20x** ($200/mo). On the Pro plan, a 5-hour usage cycle begins with your first prompt — and with a capable model like Opus 4.6, most developers hit the rate limit well before that cycle resets.

The obvious fix is upgrading to Max, but there's a steep gap: nothing exists between $20 and $100. If your usage falls somewhere in between, you're either overpaying or constantly waiting.

A cheaper alternative is running multiple Pro accounts and switching when one hits its limit. Two accounts at $40/mo total give you roughly 2x the capacity at less than half the cost of Max 5x. Three or four accounts scale the same way.

But naively switching between accounts leaves significant capacity on the table. When you activate an account matters just as much as which account you use — a single throwaway prompt sent hours before you start coding can unlock an entire extra usage cycle.

**PolyClaude** answers a simple question: given *n* Pro accounts and a fixed coding window, what's the optimal way to schedule them so you spend more time coding and less time waiting?

---

## When Combinatorial Optimization saves you money

---

## 1. Problem Statement

Johnny is a developer who relies on **Claude Code** for his daily coding sessions. He has access to **n** Claude Code Pro accounts. Each account has two operational constraints:

1. **Usage limit:** Claude Code rate-limits by tokens, but given a developer's consistent coding style, this maps to an average active time per cycle — we call it **x**.
2. **Cycle reset:** From the moment an account is first used (even for a single prompt), a fixed **T = 5 hour** countdown begins. When the timer expires, the account fully resets its rate limits.

Johnny is not a machine. He doesn't code 24hrs. Let's say, he codes during a fixed window **[s, e]**, where **W = e − s** is the total coding duration. Critically, Johnny is allowed to briefly send a prompt from an account outside his coding window to start its timer early (a "pre-activation"), giving him control over when each account's cycles align with his coding time.

### The Core Question

Given n, x, s, and e, what is the optimal strategy for scheduling account usage to minimize downtime during the coding window?

**Example:** Johnny has 3 Pro accounts, with his coding-style, his token budget lasts about 1 hour before hitting the limit. He codes from 10:00 to 20:00 (W = 10h). Without any optimization, he gets 6 hours of coding and 4 hours of waiting. With optimal scheduling, he gets 9 hours of coding and only 1 hour of waiting — same accounts, same limits, just smarter timing.

"Optimal" has two distinct interpretations, and the best strategy differs significantly between them:

- **Spread strategy:** Minimize the longest cooldown gap (uniform interruptions).
  *e.g. 9 one-hour coding blocks separated by 6-minute breaks throughout the day. so you can take a quick break.*
- **Bunch strategy:** Maximize the longest unbroken coding stretch (fewest interruptions).
  *e.g. 9 hours of continuous, focused coding followed by a single 1-hour break.*

---

## 2. Notation and Parameters

| Symbol | Meaning | Domain |
|--------|---------|--------|
| n | Number of Claude Pro accounts | n ≥ 1 (integer) |
| x | Average time before hitting cycle limit | 0 < x ≤ T |
| T | Cycle length (fixed) | T = 5 hours |
| s | Coding start time | 0 ≤ s < 24 |
| e | Coding end time | s < e ≤ 24 |
| W | Coding window length, W = e − s | W > 0 |
| k | Number of usable cycles per account | derived |
| B | Total number of coding blocks | B = n · k |
| L | Total coding hours in window | L = min(W, B · x) |
| D | Total cooldown hours in window | D = W − L |

---

## 3. Cycle Mechanics

### 3.1 Single Account Lifecycle

When an account is first used at time **a** (even a single prompt), its timeline becomes:

```
Time:    a          a+T         a+2T        a+3T
         |──cycle 0──|──cycle 1──|──cycle 2──|── ...
         [===x===]   [===x===]   [===x===]
          ↑ usable    ↑ usable    ↑ usable
```

Within each cycle **[a + jT, a + (j+1)T)**, the account can provide exactly **x** hours of active coding assistance. The placement of the x-hour block within the cycle is a free variable — Johnny chooses when to start coding with that account within each cycle.

### 3.2 Pre-Activation

Johnny may send a throwaway prompt from an account before his coding window begins to start its timer. This costs him nothing in terms of coding time but shifts the account's cycle boundaries.

If an account is pre-activated at time **a < s**, its first cycle boundary during coding time falls at **a + T** (or **a + 2T**, etc.), which may be well before the coding window ends. This allows the account to complete more cycles during the coding window than if it were first used at time **s**.

This is the single most powerful insight in the problem: pre-activation unlocks additional cycles.

---

## 4. Fundamental Analysis

### 4.1 Maximum Usable Cycles Per Account

An account's cycle **j** (spanning **[a + jT, a + (j+1)T)**) is "usable" if an entire x-hour block can fit within both the cycle and the coding window **[s, e]**. Formally:

```
Cycle j is usable  ⟺  max(a + jT, s) + x ≤ min(a + (j+1)T, e)
```

By choosing the pre-activation time **a** optimally, we maximize the number of usable cycles. The optimal count is:

$$
\boxed{k_{\max} = \left\lfloor \frac{W + 2(T - x)}{T} \right\rfloor}
$$

**Derivation:** Consider an account with first cycle start **c₀ = a + j₀T** and last cycle start **cₖ = c₀ + (k−1)T**. For the first cycle to be usable: **c₀ ≥ s − (T − x)** (the block can end by the cycle's end and start at or after s). For the last cycle to be usable: **c₀ + (k−1)T + x ≤ e + (T − x)** isn't quite right — more precisely, **cₖ ≤ e − x** (so the block fits before e). Combining:

```
c₀ ≥ s − (T − x)
c₀ + (k−1)T ≤ e − x

⟹  (k−1)T ≤ (e − x) − (s − (T − x)) = W − 2x + T

⟹  k ≤ (W − 2x + T)/T + 1 = (W + 2T − 2x) / T = (W + 2(T − x)) / T
```

Taking the floor gives **k_max**.

### 4.2 Total Available Coding Time and Cooldown

```
Total coding blocks:    B = n · k_max
Total coding hours:     L = min(W, B · x) = min(W, n · k_max · x)
Total cooldown hours:   D = W − L = max(0, W − n · k_max · x)
```

These values use full cycles only. See Section 4.5 for the refined estimate that includes partial cycle contributions.

### 4.3 Full Coverage Condition

Zero cooldown is achievable if and only if:

$$
n \cdot k_{\max} \cdot x \geq W
$$

This means there is enough total account capacity (from full cycles) to fill the entire coding window. When this holds, both optimization strategies achieve D = 0 and the problem is trivially solved by covering [s, e] completely.

**Note:** This is a sufficient condition using full cycles only. Including partial cycle contributions (Section 4.5), full coverage may be achievable even when n · k_max · x < W.

### 4.4 Numerical Examples

**Example 1:** n = 3, x = 1, W = 10 (coding 10:00–20:00)

```
k_max = ⌊(10 + 2·4) / 5⌋ = ⌊18/5⌋ = 3
B = 9,  L = 9h,  D = 1h
```

**Example 2:** n = 2, x = 1.5, W = 8

```
k_max = ⌊(8 + 2·3.5) / 5⌋ = ⌊15/5⌋ = 3
B = 6,  L = min(8, 9) = 8h  →  D = 0 (full coverage!)
```

**Example 3:** n = 3, x = 1, W = 10, no pre-activation (a = s)

```
k = ⌊(W − x) / T⌋ + 1 = ⌊9 / 5⌋ + 1 = 2  (only 2 cycles each!)
B = 6,  L = 6h,  D = 4h  (much worse)
```

This comparison highlights the power of pre-activation: from 4 hours of cooldown down to 1 hour, simply by sending a throwaway prompt in advance.

### 4.5 Partial Cycle Refinement

The k_max formula counts only cycles where a **full** x-hour block fits within both the cycle and the coding window. However, one additional cycle may partially overlap the coding window, providing fewer than x but still usable hours. Ignoring this partial contribution causes L to underestimate.

For a single account with optimal pre-activation, the total coding contribution is:

$$
L_{\text{single}} = x + \left\lfloor \frac{W - x}{T} \right\rfloor \cdot x + \min\!\left(x,\; (W - x) \bmod T\right)
$$

The three terms correspond to: (1) the pre-loaded first cycle, (2) all subsequent full cycles that fit within the remaining window, and (3) the partial contribution from the boundary cycle (capped at x).

The **refined total** across all accounts:

$$
L = \min(W, \; n \cdot L_{\text{single}}), \qquad D = W - L
$$

**When does this matter?** The partial cycle contributes r = min(x, (W − x) mod T) hours. When (W − x) mod T ≥ x, the boundary cycle fits a full x-hour block already counted by k_max, so the partial formula adds nothing beyond the full-cycle estimate. For the running example (x = 1, W = 10), (W − x) mod T = 4 ≥ x = 1, so the full-cycle formula is exact. But for larger x, the difference can be significant:

**Example:** n = 1, x = 3, W = 10

```
Full cycles only:  k_max = ⌊(10 + 4) / 5⌋ = 2  →  L = 6h,  D = 4h
With partial:      L_single = 3 + 1·3 + min(3, 2) = 8h  →  L = 8h,  D = 2h
```

The boundary cycle overlaps [s, e] by 2 hours — enough for a partial block, saving 2 additional hours.

**For the strategy analysis in Sections 5–6**, the formulas use B = n · k_max (full blocks of x hours each). The partial cycle contributions can further reduce total downtime D. In practice, partial blocks are best placed at the window boundaries (start or end) where they naturally arise.

---

## 5. Strategy A — Spread (Minimize Maximum Gap)

### 5.1 Objective

Distribute cooldown as uniformly as possible to minimize the longest interruption. This is ideal for sustained coding tasks like debugging or incremental feature work, where short pauses are tolerable but long gaps break flow.

### 5.2 Optimal Solution

Distribute **B** coding blocks across the window with **B + 1** equal gaps (before the first block, between consecutive blocks, and after the last block):

```
Gap size:        g = D / (B + 1)
Block i starts:  s + (i + 1)·g + i·x,    i = 0, 1, ..., B−1
```

Each block is anchored to an account-cycle pair, with pre-activation times chosen so that each block falls within a valid cycle.

**Practical variant:** If Johnny prefers to start and end his window with active coding, the B blocks can be anchored at the boundaries (first block at s, last block ending at e). This produces B − 1 interior gaps of size D/(B − 1) — slightly larger than the optimal, but eliminating idle time at the window edges.

### 5.3 Optimal Gap Size

Two constraints jointly determine the minimum achievable maximum gap:

**Information-theoretic bound.** B blocks create at most B + 1 cooldown segments totaling D hours. By the pigeonhole principle:

$$
g_{\text{lower}} = \frac{D}{B + 1}
$$

**Cycle-feasibility bound.** In a round-robin block assignment (account l gets every n-th block), consecutive same-account blocks are spaced n·(x + g) apart. For these to fit within consecutive T-hour cycles, the cumulative positional drift must not exceed T − x across k_max − 1 steps:

$$
g_{\text{cycle}} = \max\!\left(0,\; \frac{T(k_{\max} - 2) + x}{n(k_{\max} - 1)} - x \right)
$$

**Optimal gap.** The true minimum-maximum gap is the larger of the two bounds:

$$
\boxed{g^* = \max\!\left(\frac{D}{B + 1},\; g_{\text{cycle}}\right)}
$$

**For the original problem** (n = 3, x = 1, W = 10):

```
g_lower = 1/10 h = 6 minutes
g_cycle = max(0, (5·1 + 1)/(3·2) − 1) = max(0, 0) = 0
g* = max(6 min, 0) = 6 minutes
```

**When does g_cycle bind?** For small n or large k_max, the cycle constraint may force a larger gap. For example, with n = 1, x = 1, W = 10: g_lower = 7/4 = 1.75h but g_cycle = 2h, giving g* = 2h. For n ≥ 2 in typical configurations, g_cycle ≤ g_lower and the information-theoretic bound is tight.

### 5.4 Optimality Proof

The pigeonhole bound D/(B + 1) is unconditional: B blocks divide the window into at most B + 1 cooldown segments totaling D hours, so at least one segment has length ≥ D/(B + 1).

The cycle-feasibility bound arises from the constraint that each account's blocks must occupy valid cycles. With round-robin assignment and equal gaps g, same-account spacing is n(x + g). The block's position within its cycle drifts by n(x + g) − T per step. After k − 1 steps, this drift must stay within the cycle's usable window [0, T − x]:

```
|n(x + g) − T| · (k_max − 1) ≤ T − x
```

Solving for g (in the typical case where n(x + g) < T) yields g_cycle. The equal-spacing construction from Section 5.2, using g = g*, achieves the optimal gap exactly.

### 5.5 Pre-Activation Schedule

For the spread strategy, each account's pre-activation time must be chosen so that its cycle boundaries accommodate all assigned blocks. For account **l** (0-indexed) with assigned block start positions p₀, p₁, ..., p_{k-1}:

```
The first cycle start c₀(l) must satisfy:
  c₀(l) ≥ max over j of (pⱼ + x − (j+1)·T)     (lower bound)
  c₀(l) ≤ min over j of (pⱼ − j·T)               (upper bound)

Pre-activation time:  a(l) = c₀(l)
```

Choose the largest valid c₀ to minimize lead time before s. If c₀ < s, the pre-activation happens before the coding window begins, which is the typical case.

---

## 6. Strategy B — Bunch (Maximize Continuous Coding)

### 6.1 Objective

Maximize the length of the longest unbroken stretch of Claude Code availability. This is ideal for deep-focus tasks like building a new feature end-to-end or working through a complex refactor, where interruptions are very costly and it's better to have one long session followed by a break than many short sessions.

### 6.2 Continuous Stretch Limit

When accounts are used in round-robin fashion (Acc 1, Acc 2, ..., Acc n, Acc 1, Acc 2, ...), each account's next turn arrives n·x hours after its previous turn. An account can sustain this pattern across consecutive cycles only if the gap between consecutive uses is less than T:

```
n·x < T  →  limited continuous stretch
n·x ≥ T  →  unlimited continuous coverage (up to W)
```

When **n·x < T**, the number of consecutive cycles a single account can participate in before needing a reset is:

$$
\boxed{k_c = \left\lfloor \frac{T - x}{T - nx} \right\rfloor + 1}
$$

**Derivation:** In round-robin, account l's j-th usage starts at time s + (l + jn)·x. This must fall within cycle j of the account, which started at a + jT. The constraint is:

```
a + jT ≤ s + (l + jn)·x ≤ a + jT + (T − x)

The time between consecutive starts for the same account: n·x
The cycle length: T
Each cycle, the account's usage drifts by n·x − T (relative to cycle start).

For j cycles:  drift = j·(n·x − T)
Must satisfy:  |drift| ≤ T − x

⟹  j ≤ (T − x) / (T − n·x)     [when n·x < T, drift is negative]
⟹  k_c = ⌊(T − x) / (T − n·x)⌋ + 1
```

### 6.3 Maximum Continuous Stretch

$$
\boxed{C_{\max} = \min(W, \; n \cdot k_c \cdot x)}
$$

After this continuous block, there is a forced cooldown while at least one account resets. The remaining cycles (k_max − k_c per account) can fill in additional blocks later in the window.

### 6.4 Numerical Example

For n = 3, x = 1, W = 10:

```
n·x = 3 < 5 = T
k_c = ⌊(5 − 1) / (5 − 3)⌋ + 1 = ⌊2⌋ + 1 = 3
C_max = min(10, 3·3·1) = min(10, 9) = 9 hours!
```

Johnny can code for 9 continuous hours, with a 1-hour cooldown at the end.

Compare with the spread strategy on the same inputs: 9 blocks with 6-minute gaps. Same total coding time, radically different distribution.

---

## 7. Feasibility Conditions

For a valid schedule to exist, the following must all hold:

| Condition | Meaning |
|-----------|---------|
| x ≤ T | A single block must fit within one cycle |
| x ≤ W | A single block must fit within the coding window |
| k_max ≥ 1 | At least one cycle is usable |

The third condition expands to:

```
W + 2(T − x) ≥ T  ⟹  W ≥ 2x − T
```

Since x ≤ T, this gives W ≥ 2x − T. For x ≤ T/2 this is automatically satisfied (any positive W works). For x > T/2, we need W ≥ 2x − T.

---

## 8. The Trade-Off Space

The two strategies represent endpoints of a continuous trade-off:

```
     Spread                                         Bunch
   (uniform gaps)                           (one long stretch)
       ←──────────────────────────────────────────→
  Short max gap                            Long max gap
  Many interruptions                       Few interruptions
  Constant rhythm                          Deep focus + long break
  Best for: debugging,                     Best for: building
    incremental tasks,                       new features,
    code reviews,                            complex refactors,
    writing tests                            prototyping
```

Both strategies achieve the same total coding time L and total cooldown D. The only difference is the distribution of cooldown across the window.

### 8.1 Intermediate Strategies

Hybrid schedules are possible. For example, "2 bunched stretches with a single gap" can be modeled by partitioning the B blocks into 2 groups and solving each sub-problem. The maximum stretch would be approximately C_max/2 with a gap of approximately D, offering a middle ground.

More generally, if Johnny wants exactly **m** breaks, the gap size becomes D/m and the maximum stretch becomes approximately B·x/(m+1), subject to cycle feasibility constraints.

---

## 9. Comparison Table — Original Problem

For the concrete case **n = 3, x = 1h, T = 5h, coding window [10:00, 20:00]**:

| Strategy | Pre-act? | Cycles/account | Total Coding | Total Cooldown | Max Gap | Max Stretch |
|----------|----------|----------------|--------------|----------------|---------|-------------|
| Naïve sequential | No | 2 | 6h | 4h | 2h | 3h |
| Staggered (∞-horizon) | No | 2 | 6h | 4h | 40min | 1h |
| Spread (optimal) | Yes | 3 | 9h | 1h | 6min | 1h |
| Bunch (optimal) | Yes | 3 | 9h | 1h | 1h | 9h |

The jump from 2 to 3 cycles per account (enabled by pre-activation) is the decisive improvement. The choice between spread and bunch is then a matter of preference.

---

## 10. Special Cases and Edge Conditions

### 10.1 When n·x ≥ T (Continuous Coverage Regime)

If the combined account capacity per cycle meets or exceeds the cycle length, the accounts can sustain continuous coding indefinitely through round-robin. In this regime:

```
k_c = ∞  (no limit on continuous stretch)
C_max = W  (full window coverage, if k_max sufficient)
```

Both strategies converge to the same result: zero cooldown. This occurs, for example, with n = 5, x = 1, T = 5.

### 10.2 When x = T (Maximum Usage)

Each account can be used for the entire cycle duration. In this case:

```
k_max = ⌊(W + 0) / T⌋ = ⌊W/T⌋
```

Pre-activation provides no benefit (T − x = 0), and the problem reduces to simple tiling of T-length blocks.

### 10.3 When W ≤ n·x (Window Smaller Than One Round)

A single round-robin pass covers the entire window. Only one cycle per account is needed, and the problem is trivially solved.

### 10.4 When n = 1 (Single Account)

With one account, there is no possibility of filling gaps between cycles. The account provides x hours of coding time every T hours, with forced cooldown of T − x hours between cycles.

```
Spread: gaps of (T − x) between every block
Bunch: identical to spread (only 1 block per cycle)
```

This is the worst case, and illustrates why additional accounts are so valuable.

---

## 11. Summary of Key Equations

### Maximum usable cycles per account

$$k_{\max} = \left\lfloor \frac{W + 2(T - x)}{T} \right\rfloor$$

### Total coding time and cooldown (full cycles)

$$L_{\text{full}} = \min(W, \; n \cdot k_{\max} \cdot x), \qquad D_{\text{full}} = W - L_{\text{full}}$$

### Single-account capacity (including partial cycles)

$$L_{\text{single}} = x + \left\lfloor \frac{W - x}{T} \right\rfloor \cdot x + \min\!\left(x,\; (W - x) \bmod T\right), \qquad L = \min(W,\; n \cdot L_{\text{single}}), \qquad D = W - L$$

### Spread strategy — optimal gap

$$g_{\text{lower}} = \frac{D}{B + 1}, \qquad g_{\text{cycle}} = \max\!\left(0,\; \frac{T(k_{\max}\!-\!2) + x}{n(k_{\max}\!-\!1)} - x\right), \qquad g^{*} = \max(g_{\text{lower}},\; g_{\text{cycle}})$$

### Bunch strategy — continuous stretch limit

$$k_c = \left\lfloor \frac{T - x}{T - nx} \right\rfloor + 1 \quad (\text{when } nx < T), \qquad C_{\max} = \min(W, \; n \cdot k_c \cdot x)$$

### Full coverage condition

$$n \cdot k_{\max} \cdot x \geq W$$

### Pre-activation timing (per account)

For account l with assigned block start positions p₀ < p₁ < ⋯ < p_{k-1}, the first cycle start c₀ must satisfy:

$$\max_j\!\big(p_j + x - (j\!+\!1)T\big) \;\leq\; c_0 \;\leq\; \min_j\!\big(p_j - jT\big)$$

The pre-activation time is **a = c₀** (send a throwaway prompt at time c₀ to start the cycle timer). Choose the largest valid c₀ to minimize lead time.

---

## 12. Conclusion

The Claude Code scheduling problem is a constrained interval-packing problem with periodic renewal. Its key features are:

1. **Pre-activation** is the single most impactful technique, often unlocking an additional cycle per account and dramatically reducing total cooldown. A throwaway prompt sent hours before coding begins can mean the difference between 4 hours of downtime and 1 hour.
2. **The spread vs. bunch trade-off** is fundamental and irreducible. Total cooldown is fixed by the parameters; only its distribution can be optimized.
3. **The pigeonhole principle** applied to the B + 1 cooldown segments provides a tight lower bound D/(B + 1) on the maximum gap in the spread strategy, and the equal-spacing solution achieves this bound exactly.
4. **The continuous stretch limit** in the bunch strategy arises from the drift between round-robin period (n·x) and cycle length (T), creating a geometric constraint on how many consecutive cycles an account can participate in.

The general formulas presented here allow Johnny (or any developer managing multiple Claude Code accounts) to compute the optimal strategy for any combination of parameters, and to make an informed choice between minimizing interruptions and maximizing deep-focus coding time.
