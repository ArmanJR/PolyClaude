package tui

import (
	"fmt"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

type verifyModel struct {
	claudePath string
	found      bool
	checked    bool
}

func newVerifyModel() verifyModel {
	return verifyModel{}
}

type claudeCheckMsg struct {
	path string
	err  error
}

func (m verifyModel) init() tea.Cmd {
	return func() tea.Msg {
		path, err := exec.LookPath("claude")
		return claudeCheckMsg{path: path, err: err}
	}
}

func (m verifyModel) update(msg tea.Msg) (verifyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case claudeCheckMsg:
		m.checked = true
		if msg.err == nil {
			m.found = true
			m.claudePath = msg.path
		}
	case tea.KeyPressMsg:
		if msg.String() == "enter" {
			if m.found {
				return m, func() tea.Msg { return nextStepMsg{} }
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m verifyModel) view() string {
	s := "\n"
	s += titleStyle.Render("  Step 1: Verify Claude CLI") + "\n\n"

	if !m.checked {
		s += "  Checking for Claude CLI...\n"
		return s
	}

	if m.found {
		s += successStyle.Render("  ✓") + fmt.Sprintf(" Claude CLI found at %s\n\n", m.claudePath)
		s += mutedStyle.Render("  Press Enter to continue")
	} else {
		s += errorStyle.Render("  ✗") + " Claude CLI not found in PATH\n\n"
		s += "  Install Claude CLI first:\n"
		s += codeStyle.Render("npm install -g @anthropic-ai/claude-code") + "\n\n"
		s += mutedStyle.Render("  Press Enter to quit")
	}
	s += "\n"
	return s
}
