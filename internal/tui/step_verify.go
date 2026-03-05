package tui

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/cron"
)

type verifyModel struct {
	claudePath  string
	claudeFound bool
	crontabPath  string
	crontabFound bool
	checked      bool
}

func newVerifyModel() verifyModel {
	return verifyModel{}
}

type depsCheckMsg struct {
	claudePath   string
	claudeErr    error
	crontabPath  string
	crontabFound bool
}

func (m verifyModel) init() tea.Cmd {
	return func() tea.Msg {
		claudePath, claudeErr := exec.LookPath("claude")
		crontabPath, crontabFound := cron.CrontabPath()
		return depsCheckMsg{
			claudePath:   claudePath,
			claudeErr:    claudeErr,
			crontabPath:  crontabPath,
			crontabFound: crontabFound,
		}
	}
}

func (m verifyModel) update(msg tea.Msg) (verifyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case depsCheckMsg:
		m.checked = true
		if msg.claudeErr == nil {
			m.claudeFound = true
			m.claudePath = msg.claudePath
		}
		m.crontabFound = msg.crontabFound
		m.crontabPath = msg.crontabPath
	case tea.KeyPressMsg:
		if msg.String() == "enter" {
			if m.claudeFound && m.crontabFound {
				return m, func() tea.Msg { return nextStepMsg{} }
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m verifyModel) view() string {
	s := "\n"
	s += titleStyle.Render("  Step 1: Verify Dependencies") + "\n\n"

	if !m.checked {
		s += "  Checking dependencies...\n"
		return s
	}

	// Claude CLI status
	if m.claudeFound {
		s += successStyle.Render("  ✓") + fmt.Sprintf(" Claude CLI found at %s\n", m.claudePath)
	} else {
		s += errorStyle.Render("  ✗") + " Claude CLI not found in PATH\n"
	}

	// Crontab status
	if m.crontabFound {
		s += successStyle.Render("  ✓") + fmt.Sprintf(" crontab found at %s\n", m.crontabPath)
	} else {
		s += errorStyle.Render("  ✗") + " crontab not found in PATH\n"
	}
	s += "\n"

	allGood := m.claudeFound && m.crontabFound

	// Show install instructions for missing deps
	if !m.claudeFound {
		s += "  Install Claude CLI:\n"
		s += "  " + codeStyle.Render("curl -fsSL https://claude.ai/install.sh | bash") + "\n\n"
	}

	if !m.crontabFound {
		s += "  Install crontab:\n"
		s += crontabInstallHint()
		s += "\n"
	}

	if allGood {
		s += mutedStyle.Render("  Press Enter to continue")
	} else {
		s += mutedStyle.Render("  Press Enter to quit")
	}
	s += "\n"
	return s
}

func crontabInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "  crontab is included with macOS. If missing, reinstall Command Line Tools:\n" +
			"  " + codeStyle.Render("xcode-select --install") + "\n"
	case "linux":
		return "  " + codeStyle.Render("# Debian/Ubuntu") + "\n" +
			"  " + codeStyle.Render("sudo apt install cron && sudo service cron start") + "\n\n" +
			"  " + codeStyle.Render("# RHEL/Fedora/CentOS") + "\n" +
			"  " + codeStyle.Render("sudo dnf install cronie && sudo service crond start") + "\n\n" +
			"  " + codeStyle.Render("# Alpine") + "\n" +
			"  " + codeStyle.Render("apk add dcron && rc-update add dcron default && rc-service dcron start") + "\n\n" +
			"  " + codeStyle.Render("# Arch") + "\n" +
			"  " + codeStyle.Render("sudo pacman -S cronie && sudo service cronie start") + "\n"
	default:
		return "  Install a cron implementation for your OS to provide the crontab command.\n"
	}
}
