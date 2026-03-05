package cron

import (
	"strings"
	"testing"

	"github.com/armanjr/polyclaude/internal/scheduler"
)

func TestUpdateCrontab_Insert(t *testing.T) {
	existing := "0 * * * * some-job\n"
	entries := []string{"30 5 * * mon,tue claude -p \"say hi\""}
	result := UpdateCrontab(existing, entries)

	if !strings.Contains(result, beginMarker) {
		t.Error("missing begin marker")
	}
	if !strings.Contains(result, endMarker) {
		t.Error("missing end marker")
	}
	if !strings.Contains(result, "30 5 * * mon,tue claude") {
		t.Error("missing new entry")
	}
	if !strings.Contains(result, "some-job") {
		t.Error("existing job was removed")
	}
}

func TestUpdateCrontab_Replace(t *testing.T) {
	existing := "0 * * * * some-job\n" + beginMarker + "\nold entry\n" + endMarker + "\n"
	entries := []string{"new entry"}
	result := UpdateCrontab(existing, entries)

	if strings.Contains(result, "old entry") {
		t.Error("old entry was not removed")
	}
	if !strings.Contains(result, "new entry") {
		t.Error("new entry not found")
	}
	if !strings.Contains(result, "some-job") {
		t.Error("existing job was removed")
	}
}

func TestUpdateCrontab_Remove(t *testing.T) {
	existing := "0 * * * * some-job\n" + beginMarker + "\nold entry\n" + endMarker + "\n"
	result := UpdateCrontab(existing, nil)

	if strings.Contains(result, beginMarker) {
		t.Error("begin marker should be removed")
	}
	if strings.Contains(result, "old entry") {
		t.Error("old entry should be removed")
	}
	if !strings.Contains(result, "some-job") {
		t.Error("existing job was removed")
	}
}

func TestUpdateCrontab_EmptyCrontab(t *testing.T) {
	result := UpdateCrontab("", []string{"new entry"})
	if !strings.Contains(result, "new entry") {
		t.Error("new entry not found")
	}
}

func TestUpdateCrontab_Idempotent(t *testing.T) {
	entries := []string{"30 5 * * mon claude -p \"say hi\""}
	r1 := UpdateCrontab("", entries)
	r2 := UpdateCrontab(r1, entries)
	if r1 != r2 {
		t.Errorf("not idempotent:\nfirst:  %q\nsecond: %q", r1, r2)
	}
}

func TestGenerateEntries(t *testing.T) {
	tt := &scheduler.Timetable{
		Accounts: []scheduler.AccountSchedule{
			{AccountIndex: 0, PreActivationTime: 5.5},
			{AccountIndex: 1, PreActivationTime: 6.75},
		},
	}
	weekdays := []string{"mon", "tue", "wed"}
	dirs := []string{"/home/user/.polyclaude/accounts/acct1", "/home/user/.polyclaude/accounts/acct2"}

	claudePath := "/usr/local/bin/claude"
	entries := GenerateEntries(tt, weekdays, dirs, claudePath)
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}

	// First entry: 5.5h = 5:30
	if !strings.Contains(entries[0], "30 5") {
		t.Errorf("Entry 0 should have minute=30 hour=5, got: %s", entries[0])
	}
	if !strings.Contains(entries[0], "mon,tue,wed") {
		t.Errorf("Entry 0 should have weekdays, got: %s", entries[0])
	}
	if !strings.Contains(entries[0], "CLAUDE_CONFIG_DIR=/home/user/.polyclaude/accounts/acct1") {
		t.Errorf("Entry 0 should have config dir, got: %s", entries[0])
	}
	if !strings.Contains(entries[0], claudePath) {
		t.Errorf("Entry 0 should have absolute claude path, got: %s", entries[0])
	}

	// Second entry: 6.75h = 6:45
	if !strings.Contains(entries[1], "45 6") {
		t.Errorf("Entry 1 should have minute=45 hour=6, got: %s", entries[1])
	}
}
