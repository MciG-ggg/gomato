package gomato

import (
	"fmt"
	"gomato/pkg/common"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	statusMessageStyle = common.StatusMessageStyle
)

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func handleTick(m *App) tea.Cmd {
	if m.timeModel.TimerIsRunning && m.timeModel.TimerRemaining > 0 {
		m.timeModel.TimerRemaining--
		if m.timeModel.TimerRemaining == 0 {
			m.timeModel.TimerIsRunning = false
			if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
				m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
				m.taskManager.Save()
			}
			return m.list.NewStatusMessage(statusMessageStyle("计时结束！"))
		}
		statusMsg := fmt.Sprintf("剩余时间: %02d:%02d", m.timeModel.TimerRemaining/60, m.timeModel.TimerRemaining%60)
		statusCmd := m.list.NewStatusMessage(statusMessageStyle(statusMsg))
		if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
			m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
			m.taskManager.Save()
		}
		return tea.Batch(statusCmd, tick())
	}
	return nil
}

func updateTimeView(m *App, msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(keyMsg, m.timeViewKeys.Back):
		if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
			m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
			m.taskManager.Save()
		}
		m.currentView = taskListView
		return nil
	case key.Matches(keyMsg, m.timeViewKeys.StartPause):
		m.timeModel.TimerIsRunning = !m.timeModel.TimerIsRunning
		if m.timeModel.TimerIsRunning && m.timeModel.TimerRemaining > -2 {
			return tick()
		}
		return nil
	case key.Matches(keyMsg, m.timeViewKeys.Reset):
		m.timeModel.TimerIsRunning = false
		m.timeModel.TimerRemaining = int(m.settingModel.Settings.Pomodoro) * 60
		if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
			m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
			m.taskManager.Save()
		}
		m.currentView = taskListView
		return nil
	}
	return nil
}
