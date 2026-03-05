package tui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/armanjr/polyclaude/internal/config"
)

const (
	fieldHomeDir     = 0
	fieldNumAccounts = 1
	fieldAvgDevTime  = 2
	fieldStartTime   = 3
	fieldEndTime     = 4
	fieldWeekdays    = 5
	fieldStrategy    = 6
	fieldSubmit      = 7
	numFields        = 8
)

type inputsModel struct {
	inputs      []textinput.Model
	focusIndex  int
	dryRun      bool
	err         string
	submitted   bool
	weekdays    []string
	weekdaySel  [7]bool // mon-sun
	strategy    int     // 0=spread, 1=bunch
}

var weekdayNames = [7]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
var strategyNames = [2]string{"spread", "bunch"}

func newInputsModel(dryRun bool) inputsModel {
	inputs := make([]textinput.Model, 5)

	homeDir, _ := config.DefaultHomeDir()

	inputs[0] = textinput.New()
	inputs[0].Placeholder = homeDir
	inputs[0].SetWidth(50)
	inputs[0].Prompt = ""
	inputs[0].SetValue(homeDir)
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "3"
	inputs[1].SetWidth(10)
	inputs[1].Prompt = ""
	inputs[1].CharLimit = 2

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "1.0"
	inputs[2].SetWidth(10)
	inputs[2].Prompt = ""
	inputs[2].CharLimit = 4

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "09:00"
	inputs[3].SetWidth(10)
	inputs[3].Prompt = ""
	inputs[3].CharLimit = 5

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "17:00"
	inputs[4].SetWidth(10)
	inputs[4].Prompt = ""
	inputs[4].CharLimit = 5

	return inputsModel{
		inputs:     inputs,
		dryRun:     dryRun,
		weekdaySel: [7]bool{true, true, true, true, true, false, false}, // mon-fri default
	}
}

func (m inputsModel) init() tea.Cmd {
	return textinput.Blink
}

func (m inputsModel) update(msg tea.Msg) (inputsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			m.err = ""
			m = m.nextField()
		case "shift+tab", "up":
			m.err = ""
			m = m.prevField()
		case "enter":
			if m.focusIndex == fieldSubmit {
				return m.submit()
			}
			// Toggle for weekday/strategy fields
			if m.focusIndex == fieldWeekdays {
				// Toggle is handled by space
				m.err = ""
				m = m.nextField()
				return m, nil
			}
			if m.focusIndex == fieldStrategy {
				m.err = ""
				m = m.nextField()
				return m, nil
			}
			m.err = ""
			m = m.nextField()
		case " ":
			if m.focusIndex == fieldWeekdays {
				// Cycle through weekdays - not ideal but simple
				// Actually, let's not do anything for space in weekday
				return m, nil
			}
			if m.focusIndex == fieldStrategy {
				m.strategy = (m.strategy + 1) % 2
				return m, nil
			}
		case "1", "2", "3", "4", "5", "6", "7":
			if m.focusIndex == fieldWeekdays {
				idx := int(msg.String()[0] - '1')
				m.weekdaySel[idx] = !m.weekdaySel[idx]
				return m, nil
			}
		case "left":
			if m.focusIndex == fieldStrategy {
				m.strategy = 0
				return m, nil
			}
		case "right":
			if m.focusIndex == fieldStrategy {
				m.strategy = 1
				return m, nil
			}
		}
	}

	// Update text inputs
	if m.focusIndex < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m inputsModel) nextField() inputsModel {
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Blur()
	}
	m.focusIndex++
	if m.focusIndex >= numFields {
		m.focusIndex = numFields - 1
	}
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Focus()
	}
	return m
}

func (m inputsModel) prevField() inputsModel {
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Blur()
	}
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Focus()
	}
	return m
}

func (m inputsModel) submit() (inputsModel, tea.Cmd) {
	// Validate all fields
	homeDir := m.inputs[0].Value()
	if homeDir == "" {
		homeDir, _ = config.DefaultHomeDir()
	}

	numStr := m.inputs[1].Value()
	if numStr == "" {
		numStr = "3"
	}
	numAccounts, err := strconv.Atoi(numStr)
	if err != nil {
		m.err = "Invalid number of accounts"
		return m, nil
	}
	if err := config.ValidateNumAccounts(numAccounts); err != nil {
		m.err = err.Error()
		return m, nil
	}

	avgStr := m.inputs[2].Value()
	if avgStr == "" {
		avgStr = "1.0"
	}
	avgDevTime, err := strconv.ParseFloat(avgStr, 64)
	if err != nil {
		m.err = "Invalid average dev time"
		return m, nil
	}
	if err := config.ValidateAvgDevTime(avgDevTime); err != nil {
		m.err = err.Error()
		return m, nil
	}

	startTime := m.inputs[3].Value()
	if startTime == "" {
		startTime = "09:00"
	}
	if err := config.ValidateTimeString(startTime); err != nil {
		m.err = "Invalid start time: " + err.Error()
		return m, nil
	}

	endTime := m.inputs[4].Value()
	if endTime == "" {
		endTime = "17:00"
	}
	if err := config.ValidateTimeString(endTime); err != nil {
		m.err = "Invalid end time: " + err.Error()
		return m, nil
	}

	if err := config.ValidateTimeRange(startTime, endTime); err != nil {
		m.err = err.Error()
		return m, nil
	}

	// Collect selected weekdays
	var weekdays []string
	for i, selected := range m.weekdaySel {
		if selected {
			weekdays = append(weekdays, weekdayNames[i])
		}
	}
	if err := config.ValidateWeekdays(weekdays); err != nil {
		m.err = err.Error()
		return m, nil
	}

	strategy := strategyNames[m.strategy]

	m.submitted = true

	// Return a command that sets the config and advances
	return m, func() tea.Msg {
		return configReadyMsg{
			homeDir:     homeDir,
			numAccounts: numAccounts,
			avgDevTime:  avgDevTime,
			startTime:   startTime,
			endTime:     endTime,
			weekdays:    weekdays,
			strategy:    strategy,
		}
	}
}

type configReadyMsg struct {
	homeDir     string
	numAccounts int
	avgDevTime  float64
	startTime   string
	endTime     string
	weekdays    []string
	strategy    string
}

func (m inputsModel) view() string {
	s := "\n"
	s += titleStyle.Render("  Step 2: Configuration") + "\n\n"

	type field struct {
		label string
		desc  string
	}
	fields := []field{
		{"Home directory", "Where PolyClaude stores its config and account data"},
		{"Number of accounts", "How many Claude Pro accounts do you have?"},
		{"Avg dev time (hours)", "How long until you typically hit the rate limit per cycle?"},
		{"Start time (HH:MM)", "When you usually start coding (24h format)"},
		{"End time (HH:MM)", "When you usually stop coding (24h format)"},
	}

	for i, f := range fields {
		cursor := "  "
		if m.focusIndex == i {
			cursor = highlightStyle.Render("> ")
		}
		s += fmt.Sprintf("  %s%s\n", cursor, f.label)
		s += fmt.Sprintf("    %s\n", mutedStyle.Render(f.desc))
		s += fmt.Sprintf("    %s\n\n", m.inputs[i].View())
	}

	// Weekdays
	cursor := "  "
	if m.focusIndex == fieldWeekdays {
		cursor = highlightStyle.Render("> ")
	}
	s += fmt.Sprintf("  %sWeekdays (press 1-7 to toggle)\n", cursor)
	s += "    " + mutedStyle.Render("Which days of the week do you code?") + "\n"
	s += "    "
	for i, name := range weekdayNames {
		if m.weekdaySel[i] {
			s += successStyle.Render("["+strings.ToUpper(name)+"]") + " "
		} else {
			s += mutedStyle.Render("["+name+"]") + " "
		}
	}
	s += "\n\n"

	// Strategy
	cursor = "  "
	if m.focusIndex == fieldStrategy {
		cursor = highlightStyle.Render("> ")
	}
	s += fmt.Sprintf("  %sStrategy (left/right to change)\n", cursor)
	s += "    " + mutedStyle.Render("Spread: short breaks between blocks (e.g. 9x1h blocks with 6min gaps)") + "\n"
	s += "    " + mutedStyle.Render("Bunch:  maximize unbroken coding (e.g. 9h straight, then 1h cooldown)") + "\n"
	s += "    "
	for i, name := range strategyNames {
		if m.strategy == i {
			s += highlightStyle.Render("● "+name) + "  "
		} else {
			s += mutedStyle.Render("○ "+name) + "  "
		}
	}
	s += "\n\n"

	// Submit
	if m.focusIndex == fieldSubmit {
		s += highlightStyle.Render("  > [ Submit ]") + "\n"
	} else {
		s += mutedStyle.Render("  [ Submit ]") + "\n"
	}

	if m.err != "" {
		s += "\n" + errorStyle.Render("  Error: "+m.err) + "\n"
	}

	s += "\n" + mutedStyle.Render("  tab/shift+tab: navigate  enter: select/submit") + "\n"
	return s
}
