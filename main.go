package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/config"
	"github.com/armanjr/polyclaude/internal/tui"
	"github.com/armanjr/polyclaude/internal/updater"
)

var version = "dev"

func printUsage() {
	fmt.Printf(`PolyClaude %s
Schedule multiple Claude Code Pro accounts to minimize rate-limit downtime.

Usage:
  polyclaude              Launch the interactive setup wizard
  polyclaude update       Download and install the latest version
  polyclaude --dry-run    Preview the wizard without making changes
  polyclaude --version    Print version and exit
  polyclaude --help       Show this help
`, version)
}

func main() {
	// Handle "update" subcommand before flag parsing
	if len(os.Args) > 1 && os.Args[1] == "update" {
		if err := updater.SelfUpdate(version); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle flags manually for clean help output
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-version":
			fmt.Println("polyclaude " + version)
			os.Exit(0)
		case "--help", "-help", "-h":
			printUsage()
			os.Exit(0)
		}
	}

	dryRun := flag.Bool("dry-run", false, "")
	flag.Usage = printUsage
	flag.Parse()

	// Check for updates (cached, non-blocking)
	homeDir, _ := config.DefaultHomeDir()
	if latest := updater.CheckCached(version, homeDir); latest != "" {
		fmt.Printf("Update available: v%s -> %s (run `polyclaude update` to upgrade)\n\n", version, latest)
	}

	// Set up debug logging to file (bubbletea owns the terminal)
	f, err := tea.LogToFile("polyclaude-debug.log", "polyclaude")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error setting up log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	slog.Info("starting polyclaude", "dry_run", *dryRun)

	// Check for existing config
	if config.Exists(homeDir) && !*dryRun {
		fmt.Print("Existing configuration found at " + config.ConfigPath(homeDir) + "\n")
		fmt.Print("Start fresh? [Y/n] ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "" && answer != "y" && answer != "Y" {
			fmt.Println("Exiting.")
			os.Exit(0)
		}
	}

	m := tui.New(*dryRun)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		slog.Error("program error", "error", err)
		os.Exit(1)
	}
}
