package gomato

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TaskInputModel struct {
	inputs       []textinput.Model
	focused      int
	err          error
	submitButton string
}

type taskCreatedMsg struct {
	title       string
	description string
}

type backMsg struct{}

func NewTaskInputModel() TaskInputModel {
	var inputs = make([]textinput.Model, 2)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Title"
	inputs[0].Focus()
	inputs[0].CharLimit = 156
	inputs[0].Width = 20

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Description"
	inputs[1].CharLimit = 156
	inputs[1].Width = 50

	return TaskInputModel{
		inputs:       inputs,
		focused:      0,
		err:          nil,
		submitButton: "Create",
	}
}

func (m TaskInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TaskInputModel) Update(msg tea.Msg) (TaskInputModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs) {
				return m, func() tea.Msg {
					return taskCreatedMsg{
						title:       m.inputs[0].Value(),
						description: m.inputs[1].Value(),
					}
				}
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, func() tea.Msg { return backMsg{} }
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
	}

	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *TaskInputModel) nextInput() {
	m.focused = (m.focused + 1) % (len(m.inputs) + 1)
	if m.focused == len(m.inputs) {
		m.inputs[0].Blur()
		m.inputs[1].Blur()
		return
	}
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focused {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle()
		}
	}
}

func (m *TaskInputModel) prevInput() {
	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs)
	}

	if m.focused == len(m.inputs) {
		m.inputs[0].Blur()
		m.inputs[1].Blur()
		return
	}

	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focused {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle()
		}
	}
}

func (m TaskInputModel) View() string {
	var b strings.Builder

	b.WriteString("Create a new task\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := "> " + m.submitButton
	if m.focused == len(m.inputs) {
		button = "> " + lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.submitButton)
	}

	fmt.Fprintf(&b, "\n\n%s\n\n", button)

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(press esc to cancel)"))

	return b.String()
}

func handleBack(m *App) (tea.Model, tea.Cmd) {
	m.currentView = taskListView
	m.taskInput = NewTaskInputModel()
	return m, nil
}
