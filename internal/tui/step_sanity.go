package tui

import (
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/armanjr/polyclaude/internal/accounts"
	"github.com/armanjr/polyclaude/internal/config"
)

type sanityModel struct {
	config   *config.Config
	dryRun   bool
	spinner  spinner.Model
	results  []sanityResult
	checking int // -1 = not started, 0..n-1 = checking, n = done
	allDone  bool
}

type sanityResult struct {
	passed bool
	err    string
}

type sanityCheckDoneMsg struct {
	index  int
	passed bool
	err    string
}

func newSanityModel(cfg *config.Config, dryRun bool) sanityModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return sanityModel{
		config:   cfg,
		dryRun:   dryRun,
		spinner:  s,
		results:  make([]sanityResult, len(cfg.Accounts)),
		checking: -1,
	}
}

func (m sanityModel) init() tea.Cmd {
	if m.dryRun {
		m.allDone = true
		return func() tea.Msg {
			return sanityCheckDoneMsg{index: -1}
		}
	}
	return tea.Batch(m.spinner.Tick, m.runCheck(0))
}

func (m sanityModel) runCheck(index int) tea.Cmd {
	if index >= len(m.config.Accounts) {
		return nil
	}
	dir := m.config.Accounts[index].Dir
	return func() tea.Msg {
		passed, err := accounts.RunSanityCheck(dir, m.config.ClaudePath)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		return sanityCheckDoneMsg{index: index, passed: passed, err: errStr}
	}
}

func (m sanityModel) update(msg tea.Msg) (sanityModel, tea.Cmd) {
	switch msg := msg.(type) {
	case sanityCheckDoneMsg:
		if msg.index == -1 {
			// Dry run - mark all as passed
			for i := range m.results {
				m.results[i] = sanityResult{passed: true}
			}
			m.allDone = true
			return m, nil
		}
		m.results[msg.index] = sanityResult{passed: msg.passed, err: msg.err}
		m.checking = msg.index + 1
		if m.checking >= len(m.config.Accounts) {
			m.allDone = true
			return m, nil
		}
		return m, m.runCheck(m.checking)

	case tea.KeyPressMsg:
		if msg.String() == "enter" && m.allDone {
			return m, func() tea.Msg { return nextStepMsg{} }
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m sanityModel) view() string {
	s := "\n"
	s += titleStyle.Render("  Step 4: Sanity Checks") + "\n\n"

	if m.dryRun {
		s += highlightStyle.Render("  [DRY RUN]") + " Skipping sanity checks.\n\n"
		for i, acct := range m.config.Accounts {
			s += fmt.Sprintf("  %s Account %d (%s): would check\n",
				mutedStyle.Render("~"), i+1, acct.Name)
		}
		s += "\n" + mutedStyle.Render("  Press Enter to continue") + "\n"
		return s
	}

	for i, acct := range m.config.Accounts {
		if i < len(m.results) && (m.results[i].passed || m.results[i].err != "") {
			if m.results[i].passed {
				s += fmt.Sprintf("  %s Account %d (%s): passed\n",
					successStyle.Render("✓"), i+1, acct.Name)
			} else {
				s += fmt.Sprintf("  %s Account %d (%s): %s\n",
					errorStyle.Render("✗"), i+1, acct.Name,
					errorStyle.Render(m.results[i].err))
			}
		} else if i == m.checking || (m.checking == -1 && i == 0) {
			s += fmt.Sprintf("  %s Account %d (%s): checking...\n",
				m.spinner.View(), i+1, acct.Name)
		} else {
			s += fmt.Sprintf("  %s Account %d (%s): waiting\n",
				mutedStyle.Render("○"), i+1, acct.Name)
		}
	}

	if m.allDone {
		s += "\n" + mutedStyle.Render("  Press Enter to continue") + "\n"
	}
	return s
}
