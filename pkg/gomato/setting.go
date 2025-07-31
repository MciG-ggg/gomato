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

// 支持依赖注入 settings 的构造函数
func NewSettingModelWithSettings(settings common.Settings) SettingModel {
	tabs := []string{"General", "Timer", "Appearance", "Notifications"}

	m := SettingModel{
		Tabs:               tabs,
		inputs:             make([]textinput.Model, 4),
		Settings:           settings,
		timeDisplayOptions: []string{"ANSI艺术显示", "普通数字显示"},
		languageOptions:    []string{"中文", "English"},
	}

	// 设置时间显示方式的初始索引
	if m.Settings.TimeDisplayMode == "normal" {
		m.timeDisplayIndex = 1
	} else {
		m.timeDisplayIndex = 0 // 默认ANSI
	}

	// 设置语言的初始索引
	if m.Settings.Language == "en" {
		m.languageIndex = 1
	} else {
		m.languageIndex = 0 // 默认中文
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
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
		case 0:
			t.SetValue(fmt.Sprintf("%d", m.Settings.Pomodoro))
			t.Placeholder = "25"
			t.Prompt = "Pomodoro: "
		case 1:
			t.SetValue(fmt.Sprintf("%d", m.Settings.ShortBreak))
			t.Placeholder = "5"
			t.Prompt = "Short Break: "
		case 2:
			t.SetValue(fmt.Sprintf("%d", m.Settings.LongBreak))
			t.Placeholder = "15"
			t.Prompt = "Long Break: "
		case 3:
			t.SetValue(fmt.Sprintf("%d", m.Settings.Cycle))
			t.Placeholder = "4"
			t.Prompt = "Cycle (每周期工作/短休息次数): "
		}

		m.inputs[i] = t
	}
	return m
}

type (
	errMsg error
)

const (
	pomodoro    = 0
	shortBreak  = 1
	longBreak   = 2
	cycle       = 3
	timeDisplay = 4
	languageOpt = 5 // 新增语言选项索引
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
	Tabs               []string
	ActiveTab          int
	focusIndex         int
	inputs             []textinput.Model
	cursorMode         cursor.Mode
	Settings           common.Settings
	timeDisplayOptions []string
	timeDisplayIndex   int
	languageOptions    []string // 新增语言选项
	languageIndex      int      // 当前语言索引
}

func NewSettingModel() SettingModel {
	tabs := []string{"General", "Timer", "Appearance", "Notifications"}

	settings, err := common.LoadSettings()
	if err != nil {
		fmt.Println("could not load settings:", err)
	}

	m := SettingModel{
		Tabs:               tabs,
		inputs:             make([]textinput.Model, 4),
		Settings:           settings,
		timeDisplayOptions: []string{"ANSI艺术显示", "普通数字显示"},
		languageOptions:    []string{"中文", "English"},
	}

	// 设置时间显示方式的初始索引
	if m.Settings.TimeDisplayMode == "normal" {
		m.timeDisplayIndex = 1
	} else {
		m.timeDisplayIndex = 0 // 默认ANSI
	}

	// 设置语言的初始索引
	if m.Settings.Language == "en" {
		m.languageIndex = 1
	} else {
		m.languageIndex = 0 // 默认中文
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
			return m, func() tea.Msg { return quitMsg{} }
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

			// 保存时间显示方式
			if m.timeDisplayIndex == 1 {
				m.Settings.TimeDisplayMode = "normal"
			} else {
				m.Settings.TimeDisplayMode = "ansi"
			}

			// 保存语言
			if m.languageIndex == 1 {
				m.Settings.Language = "en"
			} else {
				m.Settings.Language = "zh"
			}

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

			// 更新焦点范围，包含时间显示方式和语言选择器
			maxFocusIndex := len(m.inputs) + 2 // +1 for time display, +1 for language
			if m.focusIndex > maxFocusIndex {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = maxFocusIndex
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
			if m.focusIndex == timeDisplay {
				m.timeDisplayIndex = min(len(m.timeDisplayOptions)-1, m.timeDisplayIndex+1)
			} else if m.focusIndex == languageOpt {
				m.languageIndex = min(len(m.languageOptions)-1, m.languageIndex+1)
			} else {
				m.ActiveTab = min(m.ActiveTab+1, len(m.Tabs)-1)
			}
			return m, nil
		case "left", "h":
			if m.focusIndex == timeDisplay {
				m.timeDisplayIndex = max(0, m.timeDisplayIndex-1)
			} else if m.focusIndex == languageOpt {
				m.languageIndex = max(0, m.languageIndex-1)
			} else {
				m.ActiveTab = max(m.ActiveTab-1, 0)
			}
			return m, nil
		// 语言选择快捷键
		case "1", "2":
			if m.focusIndex == timeDisplay {
				if msg.String() == "1" {
					m.timeDisplayIndex = 0
				} else if msg.String() == "2" {
					m.timeDisplayIndex = 1
				}
			} else if m.focusIndex == languageOpt {
				if msg.String() == "1" {
					m.languageIndex = 0
				} else if msg.String() == "2" {
					m.languageIndex = 1
				}
			}
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
		// General 标签页，仅显示语言设置
		var b strings.Builder
		languagePrompt := "语言(Language): "
		if m.focusIndex == languageOpt {
			languagePrompt = focusedStyle.Render(languagePrompt)
		}
		b.WriteString(languagePrompt)
		b.WriteRune('\n')
		for i, option := range m.languageOptions {
			prefix := "  "
			if i == m.languageIndex {
				prefix = "> "
			}
			if m.focusIndex == languageOpt {
				if i == m.languageIndex {
					b.WriteString(focusedStyle.Render(prefix + option))
				} else {
					b.WriteString(prefix + option)
				}
			} else {
				if i == m.languageIndex {
					b.WriteString(focusedStyle.Render(prefix + option))
				} else {
					b.WriteString(prefix + option)
				}
			}
			b.WriteRune('\n')
		}
		button := &blurredButton
		if m.focusIndex == len(m.inputs)+2 {
			button = &focusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)
		windowContent = b.String()
	} else if m.ActiveTab == 1 {
		// Timer 标签页
		var b strings.Builder
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		// 添加时间显示方式选择器
		b.WriteRune('\n')
		timeDisplayPrompt := "时间显示方式: "
		if m.focusIndex == timeDisplay {
			timeDisplayPrompt = focusedStyle.Render(timeDisplayPrompt)
		}
		b.WriteString(timeDisplayPrompt)
		b.WriteRune('\n')
		for i, option := range m.timeDisplayOptions {
			prefix := "  "
			if i == m.timeDisplayIndex {
				prefix = "> "
			}
			if m.focusIndex == timeDisplay {
				if i == m.timeDisplayIndex {
					b.WriteString(focusedStyle.Render(prefix + option))
				} else {
					b.WriteString(prefix + option)
				}
			} else {
				if i == m.timeDisplayIndex {
					b.WriteString(focusedStyle.Render(prefix + option))
				} else {
					b.WriteString(prefix + option)
				}
			}
			b.WriteRune('\n')
		}
		button := &blurredButton
		if m.focusIndex == len(m.inputs)+2 {
			button = &focusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)
		windowContent = b.String()
	} else {
		// 其他标签页
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

	// 重新加载时间显示方式设置
	if m.Settings.TimeDisplayMode == "normal" {
		m.timeDisplayIndex = 1
	} else {
		m.timeDisplayIndex = 0 // 默认ANSI
	}
	// 重新加载语言设置
	if m.Settings.Language == "en" {
		m.languageIndex = 1
	} else {
		m.languageIndex = 0
	}
}
