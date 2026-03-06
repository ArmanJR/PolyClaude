package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/config"
	"github.com/armanjr/polyclaude/internal/cron"
)

type finishModel struct {
	config *config.Config
	dryRun bool
}

func newFinishModel(cfg *config.Config, dryRun bool) finishModel {
	return finishModel{config: cfg, dryRun: dryRun}
}

func (m finishModel) update(msg tea.Msg) (finishModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "enter", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m finishModel) view() string {
	s := "\n"
	s += titleStyle.Render("  Setup Complete!") + "\n\n"

	if m.dryRun {
		s += highlightStyle.Render("  [DRY RUN]") + " No changes were made.\n\n"
	}

	s += successStyle.Render("  ✓") + " Configuration saved to " +
		mutedStyle.Render(config.ConfigPath(m.config.HomeDir)) + "\n"
	s += successStyle.Render("  ✓") + fmt.Sprintf(" %d account(s) configured\n", len(m.config.Accounts))
	s += successStyle.Render("  ✓") + " Cron jobs installed\n"
	s += successStyle.Render("  ✓") + " Logs: " +
		mutedStyle.Render(cron.LogDir(m.config.HomeDir)) + "\n\n"

	s += "  Useful commands:\n"
	s += "  " + codeStyle.Render("crontab -l") + "  View installed cron jobs\n"
	s += "  " + codeStyle.Render("tail -f "+cron.LogDir(m.config.HomeDir)+"/*.log") + "  Watch cron job logs\n"
	s += "  " + codeStyle.Render("polyclaude --dry-run") + "  Re-run without side effects\n\n"

	s += mutedStyle.Render("  github.com/armanjr/polyclaude") + "\n\n"
	s += mutedStyle.Render("  Press Enter or q to quit") + "\n"
	return s
}
