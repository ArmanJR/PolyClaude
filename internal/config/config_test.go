package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateNumAccounts(t *testing.T) {
	tests := []struct {
		n       int
		wantErr bool
	}{
		{0, true},
		{-1, true},
		{1, false},
		{5, false},
		{100, false},
	}
	for _, tt := range tests {
		err := ValidateNumAccounts(tt.n)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateNumAccounts(%d) error = %v, wantErr %v", tt.n, err, tt.wantErr)
		}
	}
}

func TestValidateAvgDevTime(t *testing.T) {
	tests := []struct {
		x       float64
		wantErr bool
	}{
		{0, true},
		{-1, true},
		{0.5, false},
		{1.0, false},
		{5.0, false},
		{5.1, true},
	}
	for _, tt := range tests {
		err := ValidateAvgDevTime(tt.x)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateAvgDevTime(%g) error = %v, wantErr %v", tt.x, err, tt.wantErr)
		}
	}
}

func TestValidateTimeString(t *testing.T) {
	tests := []struct {
		s       string
		wantErr bool
	}{
		{"09:00", false},
		{"00:00", false},
		{"23:59", false},
		{"24:00", true},
		{"9:00", false},
		{"abc", true},
		{"", true},
		{"12", true},
		{"12:60", true},
		{"-1:00", true},
	}
	for _, tt := range tests {
		err := ValidateTimeString(tt.s)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateTimeString(%q) error = %v, wantErr %v", tt.s, err, tt.wantErr)
		}
	}
}

func TestValidateTimeRange(t *testing.T) {
	tests := []struct {
		start, end string
		wantErr    bool
	}{
		{"09:00", "17:00", false},
		{"17:00", "09:00", true},
		{"09:00", "09:00", true},
		{"00:00", "23:59", false},
	}
	for _, tt := range tests {
		err := ValidateTimeRange(tt.start, tt.end)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateTimeRange(%q, %q) error = %v, wantErr %v", tt.start, tt.end, err, tt.wantErr)
		}
	}
}

func TestValidateWeekdays(t *testing.T) {
	tests := []struct {
		days    []string
		wantErr bool
	}{
		{[]string{"mon", "tue"}, false},
		{[]string{"MON"}, false},
		{[]string{}, true},
		{[]string{"monday"}, true},
		{[]string{"mon", "invalid"}, true},
	}
	for _, tt := range tests {
		err := ValidateWeekdays(tt.days)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateWeekdays(%v) error = %v, wantErr %v", tt.days, err, tt.wantErr)
		}
	}
}

func TestValidateStrategy(t *testing.T) {
	tests := []struct {
		s       string
		wantErr bool
	}{
		{"spread", false},
		{"bunch", false},
		{"other", true},
		{"", true},
	}
	for _, tt := range tests {
		err := ValidateStrategy(tt.s)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateStrategy(%q) error = %v, wantErr %v", tt.s, err, tt.wantErr)
		}
	}
}

func TestParseTimeToMinutes(t *testing.T) {
	tests := []struct {
		s       string
		want    int
		wantErr bool
	}{
		{"09:00", 540, false},
		{"17:00", 1020, false},
		{"00:00", 0, false},
		{"23:59", 1439, false},
	}
	for _, tt := range tests {
		got, err := ParseTimeToMinutes(tt.s)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseTimeToMinutes(%q) error = %v, wantErr %v", tt.s, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("ParseTimeToMinutes(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestParseTimeToHours(t *testing.T) {
	got, err := ParseTimeToHours("10:30")
	if err != nil {
		t.Fatal(err)
	}
	if got != 10.5 {
		t.Errorf("ParseTimeToHours(\"10:30\") = %g, want 10.5", got)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		HomeDir:     dir,
		NumAccounts: 3,
		AvgDevTime:  1.0,
		StartTime:   "10:00",
		EndTime:     "20:00",
		Weekdays:    []string{"mon", "tue", "wed", "thu", "fri"},
		Strategy:    "spread",
		Accounts: []Account{
			{Name: "test-account", Dir: filepath.Join(dir, "accounts", "test-account")},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.NumAccounts != cfg.NumAccounts {
		t.Errorf("NumAccounts = %d, want %d", loaded.NumAccounts, cfg.NumAccounts)
	}
	if loaded.Strategy != cfg.Strategy {
		t.Errorf("Strategy = %q, want %q", loaded.Strategy, cfg.Strategy)
	}
	if len(loaded.Accounts) != 1 {
		t.Errorf("len(Accounts) = %d, want 1", len(loaded.Accounts))
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if Exists(dir) {
		t.Error("Exists should return false for empty dir")
	}

	cfg := &Config{HomeDir: dir, NumAccounts: 1}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	if !Exists(dir) {
		t.Error("Exists should return true after Save")
	}
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		tz      string
		wantErr bool
	}{
		{"America/New_York", false},
		{"UTC", false},
		{"Europe/London", false},
		{"Asia/Kolkata", false},
		{"Invalid/Nowhere", true},
		{"not-a-timezone", true},
	}
	for _, tt := range tests {
		err := ValidateTimezone(tt.tz)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateTimezone(%q) error = %v, wantErr %v", tt.tz, err, tt.wantErr)
		}
	}
}

func TestDetectSystemTimezone(t *testing.T) {
	tz := DetectSystemTimezone()
	if tz == "" {
		t.Error("DetectSystemTimezone() returned empty string")
	}
	// The result must be a valid IANA name
	if err := ValidateTimezone(tz); err != nil {
		t.Errorf("DetectSystemTimezone() returned invalid timezone %q: %v", tz, err)
	}
}

func TestDefaultHomeDir(t *testing.T) {
	dir, err := DefaultHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".polyclaude")
	if dir != expected {
		t.Errorf("DefaultHomeDir() = %q, want %q", dir, expected)
	}
}
