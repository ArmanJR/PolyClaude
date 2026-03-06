package cron

import (
	"strings"
	"testing"
	"time"

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

	entries := GenerateEntries(tt, weekdays, dirs, names, claudePath, "", "/home/user/.polyclaude")

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
	// Verify logging wrapper
	if !strings.Contains(entries[0].Line, "/bin/sh -c") {
		t.Errorf("Entry 0 line should use shell wrapper, got: %s", entries[0].Line)
	}
	if !strings.Contains(entries[0].Line, "/home/user/.polyclaude/logs/slim-viper.log") {
		t.Errorf("Entry 0 line should log to account log file, got: %s", entries[0].Line)
	}
	if !strings.Contains(entries[0].Line, "rc=$?") {
		t.Errorf("Entry 0 line should capture exit code, got: %s", entries[0].Line)
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
	// Verify account 1 logs to its own file
	if !strings.Contains(entries[3].Line, "/home/user/.polyclaude/logs/bold-falcon.log") {
		t.Errorf("Entry 3 line should log to bold-falcon.log, got: %s", entries[3].Line)
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
	entries := GenerateEntries(tt, []string{"mon"}, []string{"/dir"}, []string{"solo"}, "/bin/claude", "", "/home/user/.polyclaude")
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
	entries := GenerateEntries(tt, []string{"mon"}, []string{"/dir"}, []string{"night-owl"}, "/bin/claude", "", "/home/user/.polyclaude")
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

func TestConvertToCronTime_SameTimezone(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	ct := ConvertToCronTime(9.0, loc, loc)
	if ct.Hour != 9 || ct.Minute != 0 || ct.DayOffset != 0 {
		t.Errorf("same tz: got %+v, want {Hour:9 Minute:0 DayOffset:0}", ct)
	}
}

func TestConvertToCronTime_ESTtoUTC(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	utc, _ := time.LoadLocation("UTC")

	ct := ConvertToCronTime(9.0, est, utc)

	// EST is UTC-5 (or EDT UTC-4), so 09:00 EST -> 14:00 UTC (or 13:00 in EDT)
	// The offset depends on the current date's DST state.
	now := time.Now()
	ref := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, est)
	expected := ref.In(utc)

	if ct.Hour != expected.Hour() || ct.Minute != expected.Minute() {
		t.Errorf("EST->UTC: got %02d:%02d, want %02d:%02d", ct.Hour, ct.Minute, expected.Hour(), expected.Minute())
	}
	if ct.DayOffset != 0 {
		t.Errorf("EST->UTC 09:00: expected no day shift, got %d", ct.DayOffset)
	}
}

func TestConvertToCronTime_DayCrossForward(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	utc, _ := time.LoadLocation("UTC")

	// 23:00 EST -> next day in UTC
	ct := ConvertToCronTime(23.0, est, utc)

	now := time.Now()
	ref := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, est)
	expected := ref.In(utc)

	if ct.Hour != expected.Hour() || ct.Minute != expected.Minute() {
		t.Errorf("day cross fwd: got %02d:%02d, want %02d:%02d", ct.Hour, ct.Minute, expected.Hour(), expected.Minute())
	}
	if ct.DayOffset != 1 {
		t.Errorf("day cross fwd: expected DayOffset=1, got %d", ct.DayOffset)
	}
}

func TestConvertToCronTime_DayCrossBackward(t *testing.T) {
	utc, _ := time.LoadLocation("UTC")
	est, _ := time.LoadLocation("America/New_York")

	// 01:00 UTC -> previous day in EST
	ct := ConvertToCronTime(1.0, utc, est)

	now := time.Now()
	ref := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, utc)
	expected := ref.In(est)

	if ct.Hour != expected.Hour() || ct.Minute != expected.Minute() {
		t.Errorf("day cross back: got %02d:%02d, want %02d:%02d", ct.Hour, ct.Minute, expected.Hour(), expected.Minute())
	}
	if ct.DayOffset != -1 {
		t.Errorf("day cross back: expected DayOffset=-1, got %d", ct.DayOffset)
	}
}

func TestShiftWeekdays(t *testing.T) {
	tests := []struct {
		name     string
		days     []string
		offset   int
		expected []string
	}{
		{"no shift", []string{"mon", "wed", "fri"}, 0, []string{"mon", "wed", "fri"}},
		{"+1", []string{"mon", "wed", "fri"}, 1, []string{"tue", "thu", "sat"}},
		{"-1", []string{"mon", "wed", "fri"}, -1, []string{"sun", "tue", "thu"}},
		{"wrap forward sun", []string{"sun"}, 1, []string{"mon"}},
		{"wrap backward mon", []string{"mon"}, -1, []string{"sun"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShiftWeekdays(tt.days, tt.offset)
			if len(got) != len(tt.expected) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("day[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestLogDir(t *testing.T) {
	got := LogDir("/home/user/.polyclaude")
	want := "/home/user/.polyclaude/logs"
	if got != want {
		t.Errorf("LogDir = %q, want %q", got, want)
	}
}

func TestBuildEntry_LoggingWrapper(t *testing.T) {
	entry := buildEntry(9.5, []string{"mon", "fri"}, "/cfg/dir", "swift-fox", "/bin/claude",
		"# pre-activation: swift-fox", false, nil, nil, "/home/.polyclaude/logs")

	checks := []struct {
		desc    string
		substr  string
		present bool
	}{
		{"starts with cron schedule", "30 9 * * mon,fri", true},
		{"uses shell wrapper", "/bin/sh -c", true},
		{"redirects to account log file", "/home/.polyclaude/logs/swift-fox.log", true},
		{"appends stdout+stderr", "exec >>", true},
		{"logs START marker with date", `echo "=== $(date) START`, true},
		{"logs END marker with date", `echo "=== $(date) END`, true},
		{"captures exit code", "rc=$?", true},
		{"includes exit code in END", "rc=$rc", true},
		{"sets CLAUDE_CONFIG_DIR", "CLAUDE_CONFIG_DIR=/cfg/dir", true},
		{"invokes claude binary", "/bin/claude -p", true},
		{"description strips # prefix", "START pre-activation: swift-fox", true},
		{"no mkdir in cron line", "mkdir", false},
	}
	for _, c := range checks {
		got := strings.Contains(entry.Line, c.substr)
		if got != c.present {
			t.Errorf("%s: Contains(%q) = %v, want %v\n  line: %s", c.desc, c.substr, got, c.present, entry.Line)
		}
	}
}

func TestBuildEntry_NoCrontabPercentChars(t *testing.T) {
	// '%' in crontab is interpreted as newline; generated lines must not contain any.
	tt := &scheduler.Timetable{
		Accounts: []scheduler.AccountSchedule{
			{
				AccountIndex:      0,
				PreActivationTime: 10.0,
				Blocks: []scheduler.Block{
					{AccountIndex: 0, CycleIndex: 0, Start: 11.0, End: 13.0},
				},
			},
		},
	}
	entries := GenerateEntries(tt, []string{"mon", "tue", "wed", "thu", "fri"},
		[]string{"/dir"}, []string{"acct"}, "/bin/claude", "", "/home/.polyclaude")

	for i, e := range entries {
		if strings.Contains(e.Line, "%") {
			t.Errorf("entry %d contains '%%' which crontab interprets as newline:\n  %s", i, e.Line)
		}
	}
}

func TestBuildEntry_PerAccountLogFile(t *testing.T) {
	tt := &scheduler.Timetable{
		Accounts: []scheduler.AccountSchedule{
			{AccountIndex: 0, PreActivationTime: 8.0},
			{AccountIndex: 1, PreActivationTime: 9.0},
		},
	}
	dirs := []string{"/dir/a", "/dir/b"}
	names := []string{"alpha-one", "beta-two"}
	entries := GenerateEntries(tt, []string{"mon"}, dirs, names, "/bin/claude", "", "/home/.polyclaude")

	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if !strings.Contains(entries[0].Line, "/home/.polyclaude/logs/alpha-one.log") {
		t.Errorf("entry 0 should log to alpha-one.log, got: %s", entries[0].Line)
	}
	if !strings.Contains(entries[1].Line, "/home/.polyclaude/logs/beta-two.log") {
		t.Errorf("entry 1 should log to beta-two.log, got: %s", entries[1].Line)
	}
	// Ensure they don't reference each other's log file
	if strings.Contains(entries[0].Line, "beta-two.log") {
		t.Errorf("entry 0 should not reference beta-two.log")
	}
	if strings.Contains(entries[1].Line, "alpha-one.log") {
		t.Errorf("entry 1 should not reference alpha-one.log")
	}
}

func TestBuildEntry_PostCycleDescription(t *testing.T) {
	entry := buildEntry(14.0, []string{"wed"}, "/dir", "acct", "/bin/claude",
		"# post-cycle: acct, cycle 2", false, nil, nil, "/logs")

	if !strings.Contains(entry.Line, "START post-cycle: acct, cycle 2") {
		t.Errorf("description should include post-cycle info, got: %s", entry.Line)
	}
}

func TestGenerateEntries_WithTimezone(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	utc, _ := time.LoadLocation("UTC")
	_ = est
	_ = utc

	tt := &scheduler.Timetable{
		Accounts: []scheduler.AccountSchedule{
			{
				AccountIndex:      0,
				PreActivationTime: 9.0, // 09:00 in user TZ
				Blocks: []scheduler.Block{
					{AccountIndex: 0, CycleIndex: 0, Start: 10.0, End: 12.0},
				},
			},
		},
	}

	entries := GenerateEntries(tt, []string{"mon", "fri"}, []string{"/dir"}, []string{"acct"}, "/bin/claude", "America/New_York", "/home/user/.polyclaude")
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}

	// Verify entries contain converted times (exact values depend on system TZ)
	// At minimum, verify the entries were generated and contain expected structure
	for _, e := range entries {
		if !strings.Contains(e.Line, "/bin/claude") {
			t.Errorf("entry missing claude path: %s", e.Line)
		}
		if !strings.Contains(e.Line, "CLAUDE_CONFIG_DIR=/dir") {
			t.Errorf("entry missing config dir: %s", e.Line)
		}
	}
}
