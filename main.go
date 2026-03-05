package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/config"
	"github.com/armanjr/polyclaude/internal/tui"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	dryRun := flag.Bool("dry-run", false, "Walk through the wizard without making any changes")
	flag.Parse()

	if *showVersion {
		fmt.Println("polyclaude " + version)
		os.Exit(0)
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
	homeDir, _ := config.DefaultHomeDir()
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
