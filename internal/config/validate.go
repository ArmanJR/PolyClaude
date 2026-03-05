package config

import (
	"fmt"
	"strconv"
	"strings"
)

func ValidateNumAccounts(n int) error {
	if n < 1 {
		return fmt.Errorf("number of accounts must be at least 1, got %d", n)
	}
	return nil
}

func ValidateAvgDevTime(x float64) error {
	if x <= 0 {
		return fmt.Errorf("average dev time must be positive, got %g", x)
	}
	if x > 5.0 {
		return fmt.Errorf("average dev time cannot exceed 5.0 hours (cycle length), got %g", x)
	}
	return nil
}

func ValidateTimeString(s string) error {
	_, err := ParseTimeToMinutes(s)
	return err
}

func ValidateTimeRange(start, end string) error {
	startMin, err := ParseTimeToMinutes(start)
	if err != nil {
		return fmt.Errorf("start time: %w", err)
	}
	endMin, err := ParseTimeToMinutes(end)
	if err != nil {
		return fmt.Errorf("end time: %w", err)
	}
	if endMin <= startMin {
		return fmt.Errorf("end time (%s) must be after start time (%s)", end, start)
	}
	return nil
}

var validWeekdays = map[string]bool{
	"mon": true, "tue": true, "wed": true, "thu": true,
	"fri": true, "sat": true, "sun": true,
}

func ValidateWeekdays(days []string) error {
	if len(days) == 0 {
		return fmt.Errorf("at least one weekday must be selected")
	}
	for _, d := range days {
		if !validWeekdays[strings.ToLower(d)] {
			return fmt.Errorf("invalid weekday: %q", d)
		}
	}
	return nil
}

func ValidateStrategy(s string) error {
	switch s {
	case "spread", "bunch":
		return nil
	default:
		return fmt.Errorf("strategy must be \"spread\" or \"bunch\", got %q", s)
	}
}

// ParseTimeToMinutes parses "HH:MM" (24h) and returns total minutes from midnight.
func ParseTimeToMinutes(s string) (int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time %q, expected HH:MM", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, fmt.Errorf("invalid hour in %q", s)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid minute in %q", s)
	}
	return h*60 + m, nil
}

// ParseTimeToHours parses "HH:MM" and returns hours as float64.
func ParseTimeToHours(s string) (float64, error) {
	min, err := ParseTimeToMinutes(s)
	if err != nil {
		return 0, err
	}
	return float64(min) / 60.0, nil
}
