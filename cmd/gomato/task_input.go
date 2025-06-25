package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type taskInputModel struct {
	inputs  []textinput.Model
	focused int
	err     error
}

type taskCreatedMsg struct {
	title       string
	description string
}

type backMsg struct{}

func NewTaskInputModel() taskInputModel {
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

	return taskInputModel{
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m taskInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m taskInputModel) Update(msg tea.Msg) (taskInputModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				// Last input, create task
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

func (m *taskInputModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
	for i := 0; i < len(m.inputs); i++ {
		if i == m.focused {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Placeholder style
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle()
		}
	}
}

func (m *taskInputModel) prevInput() {
	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
	for i := 0; i < len(m.inputs); i++ {
		if i == m.focused {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle()
		}
	}
}

func (m taskInputModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"Create a new task",
		m.inputs[0].View(),
		m.inputs[1].View(),
		"(press enter to confirm, esc to cancel)",
	)
}
