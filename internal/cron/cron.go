package cron

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

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

// GenerateEntries produces cron entries from a timetable.
// Each account gets a pre-activation entry and post-cycle entries (1 min after each block ends).
// claudePath must be the absolute path to the claude binary (cron has a minimal PATH).
func GenerateEntries(tt *scheduler.Timetable, weekdays []string, accountDirs []string, accountNames []string, claudePath string) []Entry {
	dow := strings.Join(weekdays, ",")
	var entries []Entry

	for _, acct := range tt.Accounts {
		if acct.AccountIndex >= len(accountDirs) {
			continue
		}
		dir := accountDirs[acct.AccountIndex]
		label := accountNames[acct.AccountIndex]

		// Pre-activation
		ct := scheduler.HoursToClockTime(acct.PreActivationTime)
		entries = append(entries, Entry{
			Comment: fmt.Sprintf("# pre-activation: %s", label),
			Line: fmt.Sprintf("%d %d * * %s CLAUDE_CONFIG_DIR=%s %s -p \"say hi\"",
				ct.Minute, ct.Hour, dow, dir, claudePath),
		})

		// Post-cycle: 1 min after each block ends
		for _, block := range acct.Blocks {
			pt := scheduler.HoursToClockTime(block.End + 1.0/60.0)
			entries = append(entries, Entry{
				Comment: fmt.Sprintf("# post-cycle: %s, cycle %d", label, block.CycleIndex+1),
				Line: fmt.Sprintf("%d %d * * %s CLAUDE_CONFIG_DIR=%s %s -p \"say hi\"",
					pt.Minute, pt.Hour, dow, dir, claudePath),
			})
		}

		slog.Info("generated cron entries", "account", label,
			"pre_activation", 1, "post_cycle", len(acct.Blocks))
	}
	return entries
}
