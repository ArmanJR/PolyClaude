package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

type introModel struct {
	dryRun bool
}

func newIntroModel(dryRun bool) introModel {
	return introModel{dryRun: dryRun}
}

func (m introModel) update(msg tea.Msg) (introModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if msg.String() == "enter" {
			return m, func() tea.Msg { return nextStepMsg{} }
		}
	}
	return m, nil
}

func (m introModel) view() string {
	s := "\n"
	s += titleStyle.Render("  PolyClaude") + "\n"
	s += subtitleStyle.Render("  github.com/armanjr/polyclaude") + "\n\n"
	s += "  Schedule multiple Claude Pro accounts to minimize rate-limit downtime.\n\n"
	s += warnStyle.Render("  Warning:") + " This tool will install cron jobs that send prompts to Claude\n"
	s += "  at scheduled times. Your machine must be awake for cron jobs to run.\n\n"

	if m.dryRun {
		s += highlightStyle.Render("  [DRY RUN]") + " No files will be written, no cron jobs will be installed.\n\n"
	}

	s += mutedStyle.Render(fmt.Sprintf("  Press Enter to continue %s", mutedStyle.Render("(ctrl+c to quit)")))
	s += "\n"
	return s
}
