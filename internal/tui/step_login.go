package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/accounts"
	"github.com/armanjr/polyclaude/internal/config"
)

type loginModel struct {
	config       *config.Config
	dryRun       bool
	currentAcct  int
	accountNames []string
	accountDirs  []string
	loginOk      []bool
	err          string
	allDone      bool
}

func newLoginModel(cfg *config.Config, dryRun bool) loginModel {
	return loginModel{
		config:       cfg,
		dryRun:       dryRun,
		accountNames: make([]string, 0, cfg.NumAccounts),
		accountDirs:  make([]string, 0, cfg.NumAccounts),
		loginOk:      make([]bool, 0, cfg.NumAccounts),
	}
}

type accountCreatedMsg struct {
	name string
	dir  string
	err  error
}

func (m loginModel) init() tea.Cmd {
	return m.createNextAccount()
}

func (m loginModel) createNextAccount() tea.Cmd {
	cfg := m.config
	dryRun := m.dryRun
	existingNames := make([]string, len(m.accountNames))
	copy(existingNames, m.accountNames)

	return func() tea.Msg {
		name := accounts.GenerateRandomName(cfg.HomeDir, existingNames)
		if dryRun {
			return accountCreatedMsg{name: name, dir: cfg.HomeDir + "/accounts/" + name}
		}
		dir, err := accounts.CreateAccountDir(cfg.HomeDir, name)
		return accountCreatedMsg{name: name, dir: dir, err: err}
	}
}

func (m loginModel) update(msg tea.Msg) (loginModel, tea.Cmd) {
	switch msg := msg.(type) {
	case accountCreatedMsg:
		if msg.err != nil {
			m.err = fmt.Sprintf("Failed to create account dir: %v", msg.err)
			return m, nil
		}
		m.accountNames = append(m.accountNames, msg.name)
		m.accountDirs = append(m.accountDirs, msg.dir)
		m.loginOk = append(m.loginOk, false)

	case tea.KeyPressMsg:
		if msg.String() == "enter" {
			if m.allDone {
				return m, func() tea.Msg { return nextStepMsg{} }
			}

			if m.currentAcct >= len(m.accountDirs) {
				return m, nil
			}

			// Check login
			if m.dryRun {
				m.loginOk[m.currentAcct] = true
				m.currentAcct++
				if m.currentAcct < m.config.NumAccounts {
					return m, m.createNextAccount()
				}
				m.allDone = true
				m.finalizeAccounts()
				return m, nil
			}

			dir := m.accountDirs[m.currentAcct]
			if accounts.VerifyLogin(dir) {
				m.loginOk[m.currentAcct] = true
				m.err = ""
				m.currentAcct++
				if m.currentAcct < m.config.NumAccounts {
					return m, m.createNextAccount()
				}
				m.allDone = true
				m.finalizeAccounts()
				return m, nil
			}

			m.err = fmt.Sprintf("Login not detected - .claude.json not found at %s/.claude.json. Please try again.", dir)
		}
	}
	return m, nil
}

func (m *loginModel) finalizeAccounts() {
	m.config.Accounts = make([]config.Account, len(m.accountNames))
	for i := range m.accountNames {
		m.config.Accounts[i] = config.Account{
			Name: m.accountNames[i],
			Dir:  m.accountDirs[i],
		}
	}
}

func (m loginModel) view() string {
	s := "\n"
	s += titleStyle.Render("  Step 3: Account Login") + "\n\n"

	// Show completed accounts
	for i := 0; i < len(m.accountNames) && i < m.currentAcct; i++ {
		s += fmt.Sprintf("  %s Account %d: %s\n",
			successStyle.Render("✓"),
			i+1,
			highlightStyle.Render(m.accountNames[i]))
	}

	if m.allDone {
		s += "\n" + successStyle.Render("  All accounts configured!") + "\n\n"
		s += mutedStyle.Render("  Press Enter to continue") + "\n"
		return s
	}

	if m.currentAcct >= len(m.accountNames) {
		s += "\n  Creating account directory...\n"
		return s
	}

	acctNum := m.currentAcct + 1
	name := m.accountNames[m.currentAcct]
	dir := m.accountDirs[m.currentAcct]

	s += "\n"
	s += boldStyle.Render(fmt.Sprintf("  Account %d of %d: ", acctNum, m.config.NumAccounts))
	s += highlightStyle.Render(fmt.Sprintf("%q", name)) + "\n\n"

	s += "  Open a " + boldStyle.Render("NEW terminal") + " and run this command:\n\n"
	s += codeStyle.Render(fmt.Sprintf("export CLAUDE_CONFIG_DIR=%s && claude", dir)) + "\n\n"
	s += "  Then inside that Claude session, type:  " + boldStyle.Render("/login") + "\n"
	s += "  This will open your browser. Log in with your Claude Pro account " +
		highlightStyle.Render(fmt.Sprintf("#%d", acctNum)) + ".\n"
	s += "  After login succeeds, you can close that terminal.\n\n"
	s += warnStyle.Render("  IMPORTANT:") + " Use a DIFFERENT Claude Pro account for each step.\n\n"

	if m.dryRun {
		s += highlightStyle.Render("  [DRY RUN]") + " Press Enter to auto-advance.\n"
	} else {
		s += mutedStyle.Render("  Press Enter when you've completed the login...") + "\n"
	}

	if m.err != "" {
		s += "\n" + errorStyle.Render("  "+m.err) + "\n"
	}

	return s
}
