package scheduler

// T is the fixed cycle length in hours.
const T = 5.0

// Params holds the input parameters for schedule computation.
type Params struct {
	N int     // Number of accounts
	X float64 // Average time before hitting cycle limit (hours)
	S float64 // Coding start time (hours from midnight)
	E float64 // Coding end time (hours from midnight)
	W float64 // Coding window length W = E - S (hours)
}

// CoreMetrics holds the fundamental derived quantities.
type CoreMetrics struct {
	KMax int     // Maximum usable cycles per account
	B    int     // Total coding blocks (N * KMax)
	L    float64 // Total coding hours in window
	D    float64 // Total cooldown hours in window
}

// Block represents a single coding block assigned to an account-cycle pair.
type Block struct {
	AccountIndex int     // Which account (0-indexed)
	CycleIndex   int     // Which cycle of this account (0-indexed)
	Start        float64 // Start time (hours from midnight)
	End          float64 // End time (hours from midnight)
}

// AccountSchedule holds the schedule for a single account.
type AccountSchedule struct {
	AccountIndex      int     // Which account (0-indexed)
	PreActivationTime float64 // When to send pre-activation prompt (hours from midnight)
	Blocks            []Block // Blocks assigned to this account
}

// Timetable is the complete computed schedule.
type Timetable struct {
	Strategy string            // "spread" or "bunch"
	Metrics  CoreMetrics       // Summary metrics
	Accounts []AccountSchedule // Per-account schedules
	Blocks   []Block           // All blocks in chronological order
}
