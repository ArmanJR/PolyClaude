package accounts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateRandomName(t *testing.T) {
	dir := t.TempDir()
	name := GenerateRandomName(dir, nil)
	if name == "" {
		t.Error("name should not be empty")
	}
	if len(name) < 3 {
		t.Errorf("name too short: %q", name)
	}

	// Should not collide with existing names
	name2 := GenerateRandomName(dir, []string{name})
	if name2 == name {
		t.Errorf("name2 should differ from name1, both are %q", name)
	}
}

func TestGenerateRandomName_Uniqueness(t *testing.T) {
	dir := t.TempDir()
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		existing := make([]string, 0, len(seen))
		for k := range seen {
			existing = append(existing, k)
		}
		name := GenerateRandomName(dir, existing)
		if seen[name] {
			t.Errorf("duplicate name generated: %q", name)
		}
		seen[name] = true
	}
}

func TestCreateAccountDir(t *testing.T) {
	dir := t.TempDir()
	path, err := CreateAccountDir(dir, "test-account")
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(dir, "accounts", "test-account")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestVerifyLogin(t *testing.T) {
	dir := t.TempDir()

	// No .claude.json
	if VerifyLogin(dir) {
		t.Error("should return false without .claude.json")
	}

	// Create .claude.json
	if err := os.WriteFile(filepath.Join(dir, ".claude.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !VerifyLogin(dir) {
		t.Error("should return true with .claude.json")
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"\x1b[31mred\x1b[0m", "red"},
		{"\x1b[1;32mgreen\x1b[0m text", "green text"},
		{"no ansi", "no ansi"},
	}
	for _, tt := range tests {
		got := stripANSI(tt.input)
		if got != tt.want {
			t.Errorf("stripANSI(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
