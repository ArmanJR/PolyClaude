package scheduler

import "math"

// ClockTime holds hour and minute components of a time of day.
type ClockTime struct {
	Hour   int
	Minute int
}

// HoursToClockTime converts fractional hours-from-midnight to a ClockTime.
func HoursToClockTime(hours float64) ClockTime {
	totalMinutes := int(math.Round(hours * 60))
	h := totalMinutes / 60
	m := totalMinutes % 60
	if h < 0 {
		h += 24
	}
	if m < 0 {
		m += 60
		h--
		if h < 0 {
			h += 24
		}
	}
	return ClockTime{Hour: h % 24, Minute: m}
}
