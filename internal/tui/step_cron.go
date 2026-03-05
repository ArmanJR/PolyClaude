package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/config"
	"github.com/armanjr/polyclaude/internal/cron"
	"github.com/armanjr/polyclaude/internal/scheduler"
)

type cronModel struct {
	config    *config.Config
	timetable *scheduler.Timetable
	dryRun    bool
	entries   []cron.Entry
	confirmed bool
	applied   bool
	err       string
}

func newCronModel(cfg *config.Config, tt *scheduler.Timetable, dryRun bool) cronModel {
	var dirs, names []string
	for _, a := range cfg.Accounts {
		dirs = append(dirs, a.Dir)
		names = append(names, a.Name)
	}

	var entries []cron.Entry
	if tt != nil {
		entries = cron.GenerateEntries(tt, cfg.Weekdays, dirs, names, cfg.ClaudePath)
	}

	return cronModel{
		config:    cfg,
		timetable: tt,
		dryRun:    dryRun,
		entries:   entries,
	}
}

func (m cronModel) update(msg tea.Msg) (cronModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "y", "Y", "enter":
			if m.applied {
				return m, func() tea.Msg { return nextStepMsg{} }
			}
			if !m.confirmed {
				m.confirmed = true
				if m.dryRun {
					m.applied = true
					return m, nil
				}
				return m, m.applyCron()
			}
		case "n", "N":
			if !m.confirmed {
				m.applied = true // skip
				return m, nil
			}
		}
	}

	// Handle cron applied result
	if msg, ok := msg.(cronAppliedMsg); ok {
		m.applied = true
		if msg.err != nil {
			m.err = msg.err.Error()
		}
		return m, nil
	}

	return m, nil
}

type cronAppliedMsg struct {
	err error
}

func (m cronModel) applyCron() tea.Cmd {
	entries := m.entries
	return func() tea.Msg {
		existing, err := cron.ReadCrontab()
		if err != nil {
			return cronAppliedMsg{err: err}
		}
		updated := cron.UpdateCrontab(existing, cron.Lines(entries))
		err = cron.WriteCrontab(updated)
		return cronAppliedMsg{err: err}
	}
}

func (m cronModel) view() string {
	s := "\n"
	s += titleStyle.Render("  Step 6: Install Cron Jobs") + "\n\n"

	if len(m.entries) == 0 {
		s += warnStyle.Render("  No cron entries to install (schedule computation may have failed).") + "\n\n"
		s += mutedStyle.Render("  Press Enter to continue") + "\n"
		return s
	}

	s += boldStyle.Render("  Generated cron entries:") + "\n\n"
	for _, entry := range m.entries {
		if entry.Comment != "" {
			s += "  " + subtitleStyle.Render(entry.Comment) + "\n"
		}
		s += "  " + mutedStyle.Render(entry.Line) + "\n"
	}
	s += "\n"

	if m.dryRun {
		s += highlightStyle.Render("  [DRY RUN]") + " Would install the above cron entries.\n\n"
		if !m.applied {
			s += mutedStyle.Render("  Press Enter to continue") + "\n"
		} else {
			s += mutedStyle.Render("  Press Enter to continue") + "\n"
		}
		return s
	}

	if m.applied {
		if m.err != "" {
			s += errorStyle.Render("  Error: "+m.err) + "\n\n"
		} else if m.confirmed {
			s += successStyle.Render("  ✓ Cron jobs installed successfully!") + "\n\n"
		} else {
			s += mutedStyle.Render("  Skipped cron installation.") + "\n\n"
		}
		s += mutedStyle.Render("  Press Enter to continue") + "\n"
		return s
	}

	s += fmt.Sprintf("  Install these cron jobs? %s/%s ",
		highlightStyle.Render("Y"),
		mutedStyle.Render("n"))
	s += "\n"
	return s
}

