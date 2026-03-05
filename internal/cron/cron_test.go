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
			{
				AccountIndex:      0,
				PreActivationTime: 5.5,
				Blocks: []scheduler.Block{
					{AccountIndex: 0, CycleIndex: 0, Start: 6.0, End: 8.5},
					{AccountIndex: 0, CycleIndex: 1, Start: 13.5, End: 16.0},
				},
			},
			{
				AccountIndex:      1,
				PreActivationTime: 6.75,
				Blocks: []scheduler.Block{
					{AccountIndex: 1, CycleIndex: 0, Start: 8.5, End: 11.0},
				},
			},
		},
	}
	weekdays := []string{"mon", "tue", "wed"}
	dirs := []string{"/home/user/.polyclaude/accounts/acct1", "/home/user/.polyclaude/accounts/acct2"}
	names := []string{"slim-viper", "bold-falcon"}
	claudePath := "/usr/local/bin/claude"

	entries := GenerateEntries(tt, weekdays, dirs, names, claudePath)

	// Account 0: 1 pre-act + 2 post-cycle = 3
	// Account 1: 1 pre-act + 1 post-cycle = 2
	if len(entries) != 5 {
		t.Fatalf("len(entries) = %d, want 5", len(entries))
	}

	// Pre-activation for account 0: 5.5h = 5:30
	if !strings.Contains(entries[0].Line, "30 5") {
		t.Errorf("Entry 0 line should have minute=30 hour=5, got: %s", entries[0].Line)
	}
	if !strings.Contains(entries[0].Line, "mon,tue,wed") {
		t.Errorf("Entry 0 line should have weekdays, got: %s", entries[0].Line)
	}
	if !strings.Contains(entries[0].Line, "CLAUDE_CONFIG_DIR=/home/user/.polyclaude/accounts/acct1") {
		t.Errorf("Entry 0 line should have config dir, got: %s", entries[0].Line)
	}
	if !strings.Contains(entries[0].Line, claudePath) {
		t.Errorf("Entry 0 line should have absolute claude path, got: %s", entries[0].Line)
	}
	if entries[0].Comment != "# pre-activation: slim-viper" {
		t.Errorf("Entry 0 comment = %q, want %q", entries[0].Comment, "# pre-activation: slim-viper")
	}

	// Post-cycle for account 0, cycle 1: end=8.5 + 1/60 ≈ 8:31
	if !strings.Contains(entries[1].Line, "31 8") {
		t.Errorf("Entry 1 line should have minute=31 hour=8, got: %s", entries[1].Line)
	}
	if entries[1].Comment != "# post-cycle: slim-viper, cycle 1" {
		t.Errorf("Entry 1 comment = %q, want %q", entries[1].Comment, "# post-cycle: slim-viper, cycle 1")
	}

	// Post-cycle for account 0, cycle 2: end=16.0 + 1/60 ≈ 16:01
	if !strings.Contains(entries[2].Line, "1 16") {
		t.Errorf("Entry 2 line should have minute=1 hour=16, got: %s", entries[2].Line)
	}
	if entries[2].Comment != "# post-cycle: slim-viper, cycle 2" {
		t.Errorf("Entry 2 comment = %q, want %q", entries[2].Comment, "# post-cycle: slim-viper, cycle 2")
	}

	// Pre-activation for account 1: 6.75h = 6:45
	if !strings.Contains(entries[3].Line, "45 6") {
		t.Errorf("Entry 3 line should have minute=45 hour=6, got: %s", entries[3].Line)
	}
	if entries[3].Comment != "# pre-activation: bold-falcon" {
		t.Errorf("Entry 3 comment = %q, want %q", entries[3].Comment, "# pre-activation: bold-falcon")
	}

	// Post-cycle for account 1, cycle 1: end=11.0 + 1/60 ≈ 11:01
	if !strings.Contains(entries[4].Line, "1 11") {
		t.Errorf("Entry 4 line should have minute=1 hour=11, got: %s", entries[4].Line)
	}
}

func TestGenerateEntries_NoBlocks(t *testing.T) {
	tt := &scheduler.Timetable{
		Accounts: []scheduler.AccountSchedule{
			{AccountIndex: 0, PreActivationTime: 7.0},
		},
	}
	entries := GenerateEntries(tt, []string{"mon"}, []string{"/dir"}, []string{"solo"}, "/bin/claude")
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1 (pre-activation only)", len(entries))
	}
	if entries[0].Comment != "# pre-activation: solo" {
		t.Errorf("comment = %q, want pre-activation", entries[0].Comment)
	}
}

func TestGenerateEntries_MidnightWrap(t *testing.T) {
	// Block ending at 23:59 (23 + 59/60 ≈ 23.9833)
	tt := &scheduler.Timetable{
		Accounts: []scheduler.AccountSchedule{
			{
				AccountIndex:      0,
				PreActivationTime: 20.0,
				Blocks: []scheduler.Block{
					{AccountIndex: 0, CycleIndex: 0, Start: 21.0, End: 23.0 + 59.0/60.0},
				},
			},
		},
	}
	entries := GenerateEntries(tt, []string{"mon"}, []string{"/dir"}, []string{"night-owl"}, "/bin/claude")
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	// Post-cycle: 23:59 + 1 min = 00:00
	if !strings.Contains(entries[1].Line, "0 0 ") {
		t.Errorf("midnight wrap failed, got: %s", entries[1].Line)
	}
}

func TestLines(t *testing.T) {
	entries := []Entry{
		{Comment: "# comment1", Line: "line1"},
		{Comment: "", Line: "line2"},
		{Comment: "# comment3", Line: "line3"},
	}
	lines := Lines(entries)
	expected := []string{"# comment1", "line1", "line2", "# comment3", "line3"}
	if len(lines) != len(expected) {
		t.Fatalf("len(lines) = %d, want %d", len(lines), len(expected))
	}
	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("lines[%d] = %q, want %q", i, lines[i], want)
		}
	}
}
