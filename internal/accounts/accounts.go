package accounts

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var adjectives = []string{
	"bright", "calm", "dark", "eager", "fair", "glad", "happy", "idle",
	"jolly", "keen", "lame", "merry", "neat", "odd", "pale", "quick",
	"rich", "safe", "tall", "vast", "warm", "bold", "cold", "deep",
	"fast", "gold", "high", "kind", "lean", "mild", "nice", "open",
	"pure", "rare", "slim", "true", "wise", "blue", "cool", "fine",
}

var nouns = []string{
	"falcon", "tiger", "eagle", "panda", "otter", "raven", "viper", "whale",
	"cobra", "crane", "shark", "bison", "cedar", "cliff", "creek", "delta",
	"ember", "flame", "glade", "haven", "ivory", "jewel", "knoll", "larch",
	"maple", "north", "oasis", "pearl", "quill", "ridge", "storm", "thorn",
	"ultra", "vivid", "brook", "stone", "frost", "grove", "heron", "lumen",
}

// GenerateRandomName creates a random adjective-noun name.
// It checks against existingNames and existing dirs to avoid collisions.
func GenerateRandomName(homeDir string, existingNames []string) string {
	existing := make(map[string]bool)
	for _, n := range existingNames {
		existing[n] = true
	}
	// Also check existing directories
	accountsDir := filepath.Join(homeDir, "accounts")
	entries, _ := os.ReadDir(accountsDir)
	for _, e := range entries {
		if e.IsDir() {
			existing[e.Name()] = true
		}
	}

	for attempt := 0; attempt < 10; attempt++ {
		adj := adjectives[rand.Intn(len(adjectives))]
		noun := nouns[rand.Intn(len(nouns))]
		name := adj + "-" + noun
		if attempt > 0 {
			name = fmt.Sprintf("%s-%s-%04d", adj, noun, rand.Intn(10000))
		}
		if !existing[name] {
			return name
		}
	}

	// Fallback to hex name
	return fmt.Sprintf("account-%08x", rand.Int31())
}

// CreateAccountDir creates the account directory structure.
func CreateAccountDir(homeDir, name string) (string, error) {
	dir := filepath.Join(homeDir, "accounts", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating account dir %s: %w", dir, err)
	}
	slog.Info("created account directory", "path", dir)
	return dir, nil
}

// VerifyLogin checks if settings.json exists in the account directory,
// indicating a successful claude /login.
func VerifyLogin(accountDir string) bool {
	_, err := os.Stat(filepath.Join(accountDir, "settings.json"))
	return err == nil
}

// RunSanityCheck runs `claude -p "say hi"` with the account's config dir
// and checks the output contains "hi".
func RunSanityCheck(accountDir string) (bool, error) {
	cmd := exec.Command("claude", "-p", "say hi")
	cmd.Env = append(os.Environ(), "CLAUDE_CONFIG_DIR="+accountDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("running sanity check: %w (output: %s)", err, string(out))
	}

	// Strip ANSI escape codes for matching
	cleaned := stripANSI(string(out))
	if strings.Contains(strings.ToLower(cleaned), "hi") {
		slog.Info("sanity check passed", "account", accountDir)
		return true, nil
	}

	slog.Warn("sanity check output did not contain 'hi'", "account", accountDir, "output", cleaned)
	return false, nil
}

// stripANSI removes ANSI escape sequences from a string.
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until we find a letter
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++ // skip the final letter
			}
			i = j
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}
