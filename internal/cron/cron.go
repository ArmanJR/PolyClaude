package cron

import (
	"bytes"
	"fmt"
	"log/slog"
	"math"
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
	out, err := exec.Command("crontab", "-l").Output()
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

// WriteCrontab writes content to the user's crontab.
func WriteCrontab(content string) error {
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = bytes.NewBufferString(content)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("writing crontab: %w", err)
	}
	slog.Info("crontab updated successfully")
	return nil
}

// GenerateEntries produces cron lines from a timetable.
// Each entry activates an account at its pre-activation time on the specified weekdays.
// claudePath must be the absolute path to the claude binary (cron has a minimal PATH).
func GenerateEntries(tt *scheduler.Timetable, weekdays []string, accountDirs []string, claudePath string) []string {
	dow := strings.Join(weekdays, ",")
	var entries []string

	for _, acct := range tt.Accounts {
		if acct.AccountIndex >= len(accountDirs) {
			continue
		}
		dir := accountDirs[acct.AccountIndex]

		// Convert pre-activation time (hours from midnight) to HH:MM
		totalMinutes := int(math.Round(acct.PreActivationTime * 60))
		hour := totalMinutes / 60
		minute := totalMinutes % 60

		// Handle negative pre-activation times (before midnight)
		if hour < 0 {
			hour += 24
		}
		if minute < 0 {
			minute += 60
			hour--
			if hour < 0 {
				hour += 24
			}
		}

		entry := fmt.Sprintf("%d %d * * %s CLAUDE_CONFIG_DIR=%s %s -p \"say hi\"",
			minute, hour, dow, dir, claudePath)
		entries = append(entries, entry)
	}

	return entries
}
