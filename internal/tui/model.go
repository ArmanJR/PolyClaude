package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/config"
	"github.com/armanjr/polyclaude/internal/scheduler"
)

const (
	stepIntro    = 0
	stepVerify   = 1
	stepInputs   = 2
	stepLogin    = 3
	stepSanity   = 4
	stepSchedule = 5
	stepCron     = 6
	stepFinish   = 7
)

type Model struct {
	step      int
	dryRun    bool
	config    *config.Config
	timetable *scheduler.Timetable
	quitting  bool

	intro    introModel
	verify   verifyModel
	inputs   inputsModel
	login    loginModel
	sanity   sanityModel
	schedule scheduleModel
	cron     cronModel
	finish   finishModel
}

func New(dryRun bool) Model {
	return Model{
		step:   stepIntro,
		dryRun: dryRun,
		config: &config.Config{},
		intro:  newIntroModel(dryRun),
		verify: newVerifyModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type nextStepMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global quit handling
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Handle step transitions
	if _, ok := msg.(nextStepMsg); ok {
		// Capture claude path when leaving verify step
		if m.step == stepVerify {
			m.config.ClaudePath = m.verify.claudePath
		}
		return m.advanceStep()
	}

	// Intercept config ready from inputs step
	if msg, ok := msg.(configReadyMsg); ok {
		m.config.HomeDir = msg.homeDir
		m.config.NumAccounts = msg.numAccounts
		m.config.AvgDevTime = msg.avgDevTime
		m.config.StartTime = msg.startTime
		m.config.EndTime = msg.endTime
		m.config.Weekdays = msg.weekdays
		m.config.Strategy = msg.strategy
		slog.Info("config ready",
			"home_dir", msg.homeDir,
			"num_accounts", msg.numAccounts,
			"strategy", msg.strategy)

		if !m.dryRun {
			if err := config.Save(m.config); err != nil {
				slog.Error("failed to save config", "error", err)
			}
		}
		return m.advanceStep()
	}

	// Intercept schedule computed from schedule step
	if msg, ok := msg.(scheduleComputedMsg); ok {
		if msg.err == nil {
			m.timetable = msg.timetable
			slog.Info("schedule computed",
				"strategy", msg.timetable.Strategy,
				"blocks", msg.timetable.Metrics.B)
		}
	}

	// Dispatch to active step
	var cmd tea.Cmd
	switch m.step {
	case stepIntro:
		m.intro, cmd = m.intro.update(msg)
	case stepVerify:
		m.verify, cmd = m.verify.update(msg)
	case stepInputs:
		m.inputs, cmd = m.inputs.update(msg)
	case stepLogin:
		m.login, cmd = m.login.update(msg)
	case stepSanity:
		m.sanity, cmd = m.sanity.update(msg)
	case stepSchedule:
		m.schedule, cmd = m.schedule.update(msg)
	case stepCron:
		m.cron, cmd = m.cron.update(msg)
	case stepFinish:
		m.finish, cmd = m.finish.update(msg)
	}
	return m, cmd
}

func (m Model) advanceStep() (tea.Model, tea.Cmd) {
	m.step++
	var cmd tea.Cmd
	switch m.step {
	case stepVerify:
		m.verify = newVerifyModel()
		cmd = m.verify.init()
	case stepInputs:
		m.inputs = newInputsModel(m.dryRun)
		cmd = m.inputs.init()
	case stepLogin:
		m.login = newLoginModel(m.config, m.dryRun)
		cmd = m.login.init()
	case stepSanity:
		m.sanity = newSanityModel(m.config, m.dryRun)
		cmd = m.sanity.init()
	case stepSchedule:
		m.schedule = newScheduleModel(m.config, m.dryRun)
		cmd = m.schedule.init()
	case stepCron:
		m.cron = newCronModel(m.config, m.timetable, m.dryRun)
	case stepFinish:
		m.finish = newFinishModel(m.config, m.dryRun)
	}
	return m, cmd
}

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("\n  Goodbye!\n\n")
	}

	var s string
	switch m.step {
	case stepIntro:
		s = m.intro.view()
	case stepVerify:
		s = m.verify.view()
	case stepInputs:
		s = m.inputs.view()
	case stepLogin:
		s = m.login.view()
	case stepSanity:
		s = m.sanity.view()
	case stepSchedule:
		s = m.schedule.view()
	case stepCron:
		s = m.cron.view()
	case stepFinish:
		s = m.finish.view()
	default:
		s = "Unknown step"
	}

	return tea.NewView(s)
}
