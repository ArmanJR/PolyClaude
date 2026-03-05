package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Account struct {
	Name string `yaml:"name"`
	Dir  string `yaml:"dir"`
}

type Config struct {
	HomeDir     string    `yaml:"home_dir"`
	NumAccounts int       `yaml:"num_accounts"`
	AvgDevTime  float64   `yaml:"avg_dev_time"`
	StartTime   string    `yaml:"start_time"`
	EndTime     string    `yaml:"end_time"`
	Weekdays    []string  `yaml:"weekdays"`
	Strategy    string    `yaml:"strategy"`
	ClaudePath  string    `yaml:"claude_path"`
	Accounts    []Account `yaml:"accounts"`
}

func DefaultHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home dir: %w", err)
	}
	return filepath.Join(home, ".polyclaude"), nil
}

func ConfigPath(homeDir string) string {
	return filepath.Join(homeDir, "config.yaml")
}

func Load(homeDir string) (*Config, error) {
	path := ConfigPath(homeDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path := ConfigPath(cfg.HomeDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

func Exists(homeDir string) bool {
	_, err := os.Stat(ConfigPath(homeDir))
	return err == nil
}
