package gomato

import (
	"fmt"
	"gomato/pkg/common"
	"gomato/pkg/logging"
	"gomato/pkg/notice"
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
		logging.Log(fmt.Sprintf("[Tick] Timer ticked, remaining: %d", m.timeModel.TimerRemaining))
		if m.timeModel.TimerRemaining == 0 {
			if m.timeModel.IsWorkSession {
				// 工作结束，cycle计数+1
				m.CurrentCycleCount++
				logging.Log(fmt.Sprintf("[Cycle] 完成一次工作，当前cycle计数: %d/%d", m.CurrentCycleCount, m.settingModel.Settings.Cycle))
				if m.CurrentCycleCount < int(m.settingModel.Settings.Cycle) {
					// 进入短休息
					m.timeModel.IsWorkSession = false
					m.timeModel.TimerRemaining = int(m.settingModel.Settings.ShortBreak) * 60
					statusMsg := fmt.Sprintf("工作结束，开始休息！\n现在是休息时间！(第%d/%d次)", m.CurrentCycleCount, m.settingModel.Settings.Cycle)
					// 通知：工作结束
					notice.SendNotification("番茄钟", "工作时间结束，开始休息！")
					if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
						m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
						m.taskManager.Save()
					}
					return tea.Batch(
						m.list.NewStatusMessage(statusMessageStyle(statusMsg)),
						tick(),
					)
				} else {
					// 达到cycle，进入长休息
					logging.Log("[Cycle] 达到cycle上限，进入长休息，重置cycle计数")
					m.CurrentCycleCount = 0
					m.timeModel.IsWorkSession = false
					m.timeModel.TimerRemaining = int(m.settingModel.Settings.LongBreak) * 60
					statusMsg := "本周期已完成，进入长休息！"
					// 通知：本周期已完成，进入长休息
					notice.SendNotification("番茄钟", "本周期已完成，进入长休息！")
					if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
						m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
						m.taskManager.Save()
					}
					return tea.Batch(
						m.list.NewStatusMessage(statusMessageStyle(statusMsg)),
						tick(),
					)
				}
			} else {
				// 休息结束，回到工作
				m.timeModel.IsWorkSession = true
				m.timeModel.TimerIsRunning = true // 自动开始新一轮
				m.timeModel.TimerRemaining = int(m.settingModel.Settings.Pomodoro) * 60
				logging.Log(fmt.Sprintf("[Cycle] 休息结束，开始新一轮工作。当前cycle计数: %d/%d", m.CurrentCycleCount, m.settingModel.Settings.Cycle))
				// 通知：休息结束，开始新一轮工作
				notice.SendNotification("番茄钟", "休息结束，开始新一轮工作！")
				if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
					m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
					m.taskManager.Save()
				}
				return tea.Batch(
					m.list.NewStatusMessage(statusMessageStyle("休息结束，开始新一轮工作！")),
					tick(),
				)
			}
		}
		// 根据设置选择时间显示方式
		var timeDisplay string
		if m.settingModel.Settings.TimeDisplayMode == "normal" {
			timeDisplay = fmt.Sprintf("%02d:%02d", m.timeModel.TimerRemaining/60, m.timeModel.TimerRemaining%60)
		} else {
			timeDisplay = common.TimeToAnsiArt(fmt.Sprintf("%02d:%02d", m.timeModel.TimerRemaining/60, m.timeModel.TimerRemaining%60))
		}
		statusMsg := fmt.Sprintf("剩余时间: %s", timeDisplay)
		statusCmd := m.list.NewStatusMessage(statusMessageStyle(statusMsg))
		if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
			m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
			m.taskManager.Save()
		}

		// 广播状态到房间
		if m.roomUI.IsInRoom() {
			m.roomUI.BroadcastState(m)
		}

		// 只有在计时器仍在运行时才返回tick命令
		if m.timeModel.TimerIsRunning && m.timeModel.TimerRemaining > 0 {
			return tea.Batch(statusCmd, tick())
		}
		return statusCmd
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
		// 广播状态变化到房间
		if m.roomUI.IsInRoom() {
			m.roomUI.BroadcastState(m)
		}
		if m.timeModel.TimerIsRunning && m.timeModel.TimerRemaining > -2 {
			return tick()
		}
		return nil
	case key.Matches(keyMsg, m.timeViewKeys.Reset):
		m.timeModel.TimerIsRunning = false
		m.timeModel.IsWorkSession = true
		m.timeModel.TimerRemaining = int(m.settingModel.Settings.Pomodoro) * 60
		if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
			m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
			m.taskManager.Save()
		}
		// 广播状态变化到房间
		if m.roomUI.IsInRoom() {
			m.roomUI.BroadcastState(m)
		}
		return nil
	}
	return nil
}
