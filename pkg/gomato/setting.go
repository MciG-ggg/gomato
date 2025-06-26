package gomato

import (
	"fmt"
	"gomato/pkg/common"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

const (
	pomodoro   = 0
	shortBreak = 1
	longBreak  = 2
	cycle      = 3
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
	helpStyle    = blurredStyle

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type SettingModel struct {
	Tabs       []string
	ActiveTab  int
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	Settings   common.Settings
}

func NewSettingModel() SettingModel {
	tabs := []string{"Timer", "Appearance", "Notifications"}

	settings, err := common.LoadSettings()
	if err != nil {
		fmt.Println("could not load settings:", err)
	}

	m := SettingModel{
		Tabs:     tabs,
		inputs:   make([]textinput.Model, 4),
		Settings: settings,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 3
		t.Validate = func(s string) error {
			if s == "" {
				return nil
			}
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be a number")
			}
			return nil
		}

		switch i {
		case pomodoro:
			t.SetValue(fmt.Sprintf("%d", m.Settings.Pomodoro))
			t.Placeholder = "25"
			t.Focus()
			t.Prompt = "Pomodoro: "
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case shortBreak:
			t.SetValue(fmt.Sprintf("%d", m.Settings.ShortBreak))
			t.Placeholder = "5"
			t.Prompt = "Short Break: "
		case longBreak:
			t.SetValue(fmt.Sprintf("%d", m.Settings.LongBreak))
			t.Placeholder = "15"
			t.Prompt = "Long Break: "
		case cycle:
			t.SetValue(fmt.Sprintf("%d", m.Settings.Cycle))
			t.Placeholder = "4"
			t.Prompt = "Cycle (每周期工作/短休息次数): "
		}

		m.inputs[i] = t
	}

	return m
}

func (m SettingModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SettingModel) Update(msg tea.Msg) (SettingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			return m, func() tea.Msg { return backMsg{} }
		case "enter":
			// Persist the settings
			p, err := strconv.Atoi(m.inputs[pomodoro].Value())
			if err != nil {
				p = int(m.Settings.Pomodoro) // fallback to existing
			}
			sb, err := strconv.Atoi(m.inputs[shortBreak].Value())
			if err != nil {
				sb = int(m.Settings.ShortBreak) // fallback to existing
			}
			lb, err := strconv.Atoi(m.inputs[longBreak].Value())
			if err != nil {
				lb = int(m.Settings.LongBreak) // fallback to existing
			}
			cy, err := strconv.Atoi(m.inputs[cycle].Value())
			if err != nil {
				cy = int(m.Settings.Cycle)
			}

			m.Settings.Pomodoro = uint(p)
			m.Settings.ShortBreak = uint(sb)
			m.Settings.LongBreak = uint(lb)
			m.Settings.Cycle = uint(cy)
			m.Settings.Save()

			return m, func() tea.Msg { return backMsg{} }

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		case "right", "l":
			m.ActiveTab = min(m.ActiveTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "h":
			m.ActiveTab = max(m.ActiveTab-1, 0)
			return m, nil
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *SettingModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func (m SettingModel) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.ActiveTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	var windowContent string
	if m.ActiveTab == 0 {
		var b strings.Builder
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}

		button := &blurredButton
		if m.focusIndex == len(m.inputs) {
			button = &focusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)
		windowContent = b.String()
	} else {
		// Just placeholder content for other tabs for now
		windowContent = fmt.Sprintf("%s Content", m.Tabs[m.ActiveTab])
	}

	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(windowContent))
	doc.WriteString("\n" + helpStyle.Render("  ↑/↓: navigate • tab: next field • enter: confirm • q: quit"))
	return docStyle.Render(doc.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *SettingModel) ReloadInputsFromSettings() {
	settings, err := common.LoadSettings()
	if err == nil {
		m.Settings = settings
	}
	for i := range m.inputs {
		switch i {
		case pomodoro:
			m.inputs[i].SetValue(fmt.Sprintf("%d", m.Settings.Pomodoro))
		case shortBreak:
			m.inputs[i].SetValue(fmt.Sprintf("%d", m.Settings.ShortBreak))
		case longBreak:
			m.inputs[i].SetValue(fmt.Sprintf("%d", m.Settings.LongBreak))
		case cycle:
			m.inputs[i].SetValue(fmt.Sprintf("%d", m.Settings.Cycle))
		}
	}
}
