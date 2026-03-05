package tui

import (
	"fmt"
	"math"

	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/config"
	"github.com/armanjr/polyclaude/internal/scheduler"
)

type scheduleModel struct {
	config    *config.Config
	dryRun    bool
	timetable *scheduler.Timetable
	err       string
	computed  bool
}

type scheduleComputedMsg struct {
	timetable *scheduler.Timetable
	err       error
}

func newScheduleModel(cfg *config.Config, dryRun bool) scheduleModel {
	return scheduleModel{config: cfg, dryRun: dryRun}
}

func (m scheduleModel) init() tea.Cmd {
	cfg := m.config
	return func() tea.Msg {
		startHours, _ := config.ParseTimeToHours(cfg.StartTime)
		endHours, _ := config.ParseTimeToHours(cfg.EndTime)

		p := scheduler.Params{
			N: cfg.NumAccounts,
			X: cfg.AvgDevTime,
			S: startHours,
			E: endHours,
			W: endHours - startHours,
		}

		tt, err := scheduler.Schedule(p, cfg.Strategy)
		return scheduleComputedMsg{timetable: tt, err: err}
	}
}

func (m scheduleModel) update(msg tea.Msg) (scheduleModel, tea.Cmd) {
	switch msg := msg.(type) {
	case scheduleComputedMsg:
		m.computed = true
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.timetable = msg.timetable
		}
	case tea.KeyPressMsg:
		if msg.String() == "enter" && m.computed {
			return m, func() tea.Msg { return nextStepMsg{} }
		}
	}
	return m, nil
}

func (m scheduleModel) view() string {
	s := "\n"
	s += titleStyle.Render("  Step 5: Computed Schedule") + "\n\n"

	if !m.computed {
		s += "  Computing schedule...\n"
		return s
	}

	if m.err != "" {
		s += errorStyle.Render("  Error: "+m.err) + "\n\n"
		s += mutedStyle.Render("  Press Enter to continue") + "\n"
		return s
	}

	tt := m.timetable
	met := tt.Metrics

	s += boldStyle.Render("  Summary") + "\n"
	s += fmt.Sprintf("  Strategy:          %s\n", highlightStyle.Render(tt.Strategy))
	s += fmt.Sprintf("  Cycles/account:    %d\n", met.KMax)
	s += fmt.Sprintf("  Total blocks:      %d\n", met.B)
	s += fmt.Sprintf("  Total coding:      %s\n", successStyle.Render(fmt.Sprintf("%.1fh", met.L)))
	s += fmt.Sprintf("  Total cooldown:    %s\n", warnStyle.Render(fmt.Sprintf("%.1fh", met.D)))
	s += "\n"

	// Pre-activation times
	s += boldStyle.Render("  Pre-Activation Times") + "\n"
	for _, acct := range tt.Accounts {
		name := fmt.Sprintf("Account %d", acct.AccountIndex+1)
		if acct.AccountIndex < len(m.config.Accounts) {
			name = m.config.Accounts[acct.AccountIndex].Name
		}
		s += fmt.Sprintf("  %-20s %s\n", name, highlightStyle.Render(formatTime(acct.PreActivationTime)))
	}
	s += "\n"

	// Block timeline
	s += boldStyle.Render("  Block Timeline") + "\n"
	s += fmt.Sprintf("  %-8s %-8s %-20s %s\n",
		mutedStyle.Render("Start"),
		mutedStyle.Render("End"),
		mutedStyle.Render("Account"),
		mutedStyle.Render("Cycle"))
	for _, block := range tt.Blocks {
		name := fmt.Sprintf("Account %d", block.AccountIndex+1)
		if block.AccountIndex < len(m.config.Accounts) {
			name = m.config.Accounts[block.AccountIndex].Name
		}
		s += fmt.Sprintf("  %-8s %-8s %-20s %d\n",
			formatTime(block.Start),
			formatTime(block.End),
			name,
			block.CycleIndex+1)
	}

	s += "\n" + mutedStyle.Render("  Press Enter to continue") + "\n"
	return s
}

// formatTime converts hours-from-midnight to HH:MM string.
func formatTime(hours float64) string {
	totalMinutes := int(math.Round(hours * 60))
	h := totalMinutes / 60
	m := totalMinutes % 60
	if h < 0 {
		h += 24
	}
	if m < 0 {
		m += 60
		h--
		if h < 0 {
			h += 24
		}
	}
	return fmt.Sprintf("%02d:%02d", h%24, m)
}
