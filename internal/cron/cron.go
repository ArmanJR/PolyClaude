package cron

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/armanjr/polyclaude/internal/scheduler"
)

const (
	beginMarker = "# BEGIN polyclaude"
	endMarker   = "# END polyclaude"
)

// ReadCrontab reads the current user's crontab.
func ReadCrontab() (string, error) {
	path, _ := CrontabPath()
	if path == "" {
		path = "crontab"
	}
	out, err := exec.Command(path, "-l").Output()
	if err != nil {
		slog.Info("no existing crontab found, starting fresh")
		return "", nil
	}
	slog.Info("read existing crontab", "bytes", len(out))
	return string(out), nil
}

// UpdateCrontab splices new entries between markers, preserving other entries.
func UpdateCrontab(existing string, newEntries []string) string {
	lines := strings.Split(existing, "\n")

	var kept []string
	inside := false
	for _, line := range lines {
		switch line {
		case beginMarker:
			inside = true
		case endMarker:
			inside = false
		default:
			if !inside {
				kept = append(kept, line)
			}
		}
	}

	// Trim trailing blank lines
	for len(kept) > 0 && kept[len(kept)-1] == "" {
		kept = kept[:len(kept)-1]
	}

	if len(newEntries) == 0 {
		if len(kept) == 0 {
			return ""
		}
		return strings.Join(kept, "\n") + "\n"
	}

	block := []string{beginMarker}
	block = append(block, newEntries...)
	block = append(block, endMarker)

	result := append(kept, block...)
	return strings.Join(result, "\n") + "\n"
}

// CrontabPath resolves the crontab binary, falling back to common locations.
// Returns the path and true if found, or empty string and false if not.
func CrontabPath() (string, bool) {
	if path, err := exec.LookPath("crontab"); err == nil {
		return path, true
	}
	for _, p := range []string{"/usr/bin/crontab", "/bin/crontab"} {
		if _, err := exec.LookPath(p); err == nil {
			return p, true
		}
	}
	return "", false
}

// WriteCrontab writes content to the user's crontab.
func WriteCrontab(content string) error {
	path, _ := CrontabPath()
	if path == "" {
		path = "crontab"
	}
	cmd := exec.Command(path, "-")
	cmd.Stdin = bytes.NewBufferString(content)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("writing crontab: %w", err)
	}
	slog.Info("crontab updated successfully")
	return nil
}

// LogDir returns the directory where cron job logs are stored.
func LogDir(homeDir string) string {
	return filepath.Join(homeDir, "logs")
}

// Entry is a single cron entry with an optional descriptive comment.
type Entry struct {
	Comment string // e.g. "# pre-activation: slim-viper"
	Line    string // the executable cron line
}

// Lines returns entries as flat strings (comment + line interleaved) for crontab splicing.
func Lines(entries []Entry) []string {
	out := make([]string, 0, len(entries)*2)
	for _, e := range entries {
		if e.Comment != "" {
			out = append(out, e.Comment)
		}
		out = append(out, e.Line)
	}
	return out
}

// CronTime holds the system-local hour, minute, and day offset resulting from
// converting a user-timezone time to the system timezone.
type CronTime struct {
	Hour      int // 0-23
	Minute    int // 0-59
	DayOffset int // -1, 0, or +1
}

// ConvertToCronTime converts a fractional-hours time in userLoc to systemLoc,
// returning the system-local hour/minute and any day shift.
func ConvertToCronTime(userHours float64, userLoc, systemLoc *time.Location) CronTime {
	ct := scheduler.HoursToClockTime(userHours)
	now := time.Now()
	// Build a time in the user's timezone using today's date
	userTime := time.Date(now.Year(), now.Month(), now.Day(), ct.Hour, ct.Minute, 0, 0, userLoc)
	sysTime := userTime.In(systemLoc)

	dayOffset := sysTime.Day() - userTime.Day()
	// Clamp to -1..+1 (handles month boundaries)
	if dayOffset > 1 {
		dayOffset = -1
	} else if dayOffset < -1 {
		dayOffset = 1
	}

	return CronTime{
		Hour:      sysTime.Hour(),
		Minute:    sysTime.Minute(),
		DayOffset: dayOffset,
	}
}

// weekdayIndex maps 3-letter day names to 0-6 (mon=0 .. sun=6).
var weekdayIndex = map[string]int{
	"mon": 0, "tue": 1, "wed": 2, "thu": 3,
	"fri": 4, "sat": 5, "sun": 6,
}

// weekdayName maps 0-6 back to 3-letter day names.
var weekdayName = [7]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

// ShiftWeekdays shifts each weekday name by offset using modular arithmetic.
func ShiftWeekdays(weekdays []string, offset int) []string {
	if offset == 0 {
		return weekdays
	}
	shifted := make([]string, len(weekdays))
	for i, day := range weekdays {
		idx := weekdayIndex[strings.ToLower(day)]
		newIdx := ((idx + offset) % 7 + 7) % 7
		shifted[i] = weekdayName[newIdx]
	}
	return shifted
}

// GenerateEntries produces cron entries from a timetable.
// Each account gets a pre-activation entry and post-block entries.
// Post-block entries fire at c₀ + j*(T + 1min) for j=1..K, where c₀ is the
// pre-activation time and K is the number of coding cycles. Each activation
// starts a fresh T-hour rate-limit timer, so the 1-minute delays compound.
// claudePath must be the absolute path to the claude binary (cron has a minimal PATH).
// userTZ is the IANA timezone of the user's schedule times. If empty, no conversion is done.
func GenerateEntries(tt *scheduler.Timetable, weekdays []string, accountDirs []string, accountNames []string, claudePath string, userTZ string, homeDir string) []Entry {
	logDir := LogDir(homeDir)

	var userLoc, systemLoc *time.Location
	convertTZ := false
	if userTZ != "" {
		var err error
		userLoc, err = time.LoadLocation(userTZ)
		if err != nil {
			slog.Error("failed to load user timezone, skipping conversion", "timezone", userTZ, "error", err)
		} else {
			systemLoc = time.Now().Location()
			if userLoc.String() != systemLoc.String() {
				convertTZ = true
				slog.Info("timezone conversion enabled",
					"user_tz", userLoc.String(),
					"system_tz", systemLoc.String())
				slog.Warn("DST caveat: cron entries use current UTC offset; if DST changes the offset, entries will be off by ~1 hour until re-run")
			}
		}
	}

	var entries []Entry
	for _, acct := range tt.Accounts {
		if acct.AccountIndex >= len(accountDirs) {
			continue
		}
		dir := accountDirs[acct.AccountIndex]
		label := accountNames[acct.AccountIndex]

		// Pre-activation
		entries = append(entries, buildEntry(
			acct.PreActivationTime, weekdays, dir, label, claudePath,
			fmt.Sprintf("# pre-activation: %s", label),
			convertTZ, userLoc, systemLoc, logDir,
		))

		// Post-block: c₀ + j*(T + 1/60) for each cycle j=1..K
		// Each activation starts a new T-hour timer, so delays compound.
		for j := 1; j <= len(acct.Blocks); j++ {
			postTime := acct.PreActivationTime + float64(j)*(scheduler.T+1.0/60.0)
			entries = append(entries, buildEntry(
				postTime, weekdays, dir, label, claudePath,
				fmt.Sprintf("# post-block: %s, block %d", label, j),
				convertTZ, userLoc, systemLoc, logDir,
			))
		}

		slog.Info("generated cron entries", "account", label,
			"pre_activation", 1, "post_block", len(acct.Blocks))
	}
	return entries
}

func buildEntry(hours float64, weekdays []string, dir, label, claudePath, comment string, convertTZ bool, userLoc, systemLoc *time.Location, logDir string) Entry {
	var minute, hour int
	dow := strings.Join(weekdays, ",")

	if convertTZ {
		ct := ConvertToCronTime(hours, userLoc, systemLoc)
		minute = ct.Minute
		hour = ct.Hour
		if ct.DayOffset != 0 {
			shifted := ShiftWeekdays(weekdays, ct.DayOffset)
			dow = strings.Join(shifted, ",")
		}
	} else {
		ct := scheduler.HoursToClockTime(hours)
		minute = ct.Minute
		hour = ct.Hour
	}

	desc := strings.TrimPrefix(comment, "# ")
	logFile := filepath.Join(logDir, label+".log")

	// Wrap the command in a shell that logs output with timestamps.
	// Uses plain `date` (no format specifiers) to avoid crontab % escaping issues.
	// Log directory and files are pre-created during cron installation.
	line := fmt.Sprintf(
		`%d %d * * %s /bin/sh -c 'exec >>%s 2>&1; echo "=== $(date) START %s ==="; CLAUDE_CONFIG_DIR=%s %s -p "say hi"; rc=$?; echo "=== $(date) END rc=$rc ==="'`,
		minute, hour, dow, logFile, desc, dir, claudePath,
	)

	return Entry{
		Comment: comment,
		Line:    line,
	}
}
