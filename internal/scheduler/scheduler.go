package scheduler

import "fmt"

// Schedule dispatches to the appropriate strategy and returns a complete timetable.
func Schedule(p Params, strategy string) (*Timetable, error) {
	switch strategy {
	case "spread":
		return SpreadSchedule(p)
	case "bunch":
		return BunchSchedule(p)
	default:
		return nil, fmt.Errorf("unknown strategy: %q (must be \"spread\" or \"bunch\")", strategy)
	}
}
