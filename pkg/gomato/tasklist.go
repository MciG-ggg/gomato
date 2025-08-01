package gomato

import (
	"gomato/pkg/common"
	"gomato/pkg/keymap"
	"gomato/pkg/task"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// 创建任务列表视图
func NewTaskList(listKeys *keymap.ListKeyMap, delegateKeys *keymap.DelegateKeyMap, taskManager *task.Manager) list.Model {
	items := make([]list.Item, len(taskManager.Tasks))
	for i, t := range taskManager.Tasks {
		items[i] = t
	}
	delegate := newItemDelegate(delegateKeys)
	taskList := list.New(items, delegate, 0, 0)
	taskList.Title = "番茄钟任务列表"
	taskList.Styles.Title = common.TitleStyle
	taskList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.Setting,
			listKeys.InsertItem,
			listKeys.ToggleTitleBar,
			listKeys.ToggleStatusBar,
			listKeys.TogglePagination,
			listKeys.ToggleHelpMenu,
			listKeys.JoinRoom,
			listKeys.LeaveRoom,
			listKeys.ShowMembers,
		}
	}
	return taskList
}

// 处理任务创建消息
func handleTaskCreated(m *App, msg taskCreatedMsg) (tea.Model, tea.Cmd) {
	m.taskManager.AddItem(msg.title, msg.description)
	m.taskManager.Tasks[len(m.taskManager.Tasks)-1].Timer = task.TimeModel{
		TimerDuration:  25 * 60,
		TimerRemaining: 25 * 60,
		TimerIsRunning: false,
		IsWorkSession:  true,
	}
	newTask := m.taskManager.Tasks[len(m.taskManager.Tasks)-1]
	insertCmd := m.taskList.InsertItem(len(m.taskList.Items()), newTask)
	statusCmd := m.taskList.NewStatusMessage(statusMessageStyle("添加了新任务: " + newTask.Title()))
	m.currentView = taskListView
	return m, tea.Batch(insertCmd, statusCmd)
}

// 创建一个新的列表项委托
// 该委托处理任务列表项的交互逻辑
// 包括选择任务、删除任务等操作
// 并返回一个新的委托实例
func newItemDelegate(keys *keymap.DelegateKeyMap) list.ItemDelegate {
	d := list.NewDefaultDelegate()
	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string
		if i, ok := m.SelectedItem().(task.Task); ok {
			title = i.Title()
		} else {
			return nil
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.Choose):
				return m.NewStatusMessage("You chose " + title)
			}
		}
		return nil
	}
	help := []key.Binding{keys.Choose, keys.Remove}
	d.ShortHelpFunc = func() []key.Binding {
		return help
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}
	return d
}

// 更新任务列表视图
func updateTaskListView(m *App, msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch keyMsg := msg.(type) {
	case tea.KeyMsg:
		if m.taskList.FilterState() == list.Filtering {
			break
		}
		switch {
		case key.Matches(keyMsg, m.taskListViewKeys.Setting):
			common.LoadSettings()
			m.settingModel.ReloadInputsFromSettings()
			m.currentView = settingView
			return nil
		case key.Matches(keyMsg, m.taskListViewKeys.ToggleTitleBar):
			v := !m.taskList.ShowTitle()
			m.taskList.SetShowTitle(v)
			m.taskList.SetShowFilter(v)
			m.taskList.SetFilteringEnabled(v)
			return nil
		case key.Matches(keyMsg, m.taskListViewKeys.ToggleStatusBar):
			m.taskList.SetShowStatusBar(!m.taskList.ShowStatusBar())
			return nil
		case key.Matches(keyMsg, m.taskListViewKeys.TogglePagination):
			m.taskList.SetShowPagination(!m.taskList.ShowPagination())
			return nil
		case key.Matches(keyMsg, m.taskListViewKeys.ToggleHelpMenu):
			m.taskList.SetShowHelp(!m.taskList.ShowHelp())
			return nil
		case key.Matches(keyMsg, m.taskListViewKeys.InsertItem):
			m.currentView = taskInputView
			return nil
		case key.Matches(keyMsg, m.delegateKeys.Remove):
			index := m.taskList.Index()
			if index >= 0 && index < len(m.taskManager.Tasks) {
				deletedTaskTitle := m.taskManager.Tasks[index].Title()
				m.taskManager.DeleteItem(index)
				m.taskList.RemoveItem(index)
				if len(m.taskList.Items()) == 0 {
					m.delegateKeys.Remove.SetEnabled(false)
				}
				statusCmd := m.taskList.NewStatusMessage(statusMessageStyle("删除了任务: " + deletedTaskTitle))
				return statusCmd
			}
		case key.Matches(keyMsg, m.taskListViewKeys.ChooseTask):
			// 选择任务并开始计时
			m.currentTaskIndex = m.taskList.Index()
			if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
				m.timeModel = m.taskManager.Tasks[m.currentTaskIndex].Timer
			}
			m.currentView = timeView
			m.timeModel.TimerIsRunning = true
			return tea.Batch(
				m.taskList.NewStatusMessage(statusMessageStyle("任务已选择，计时已开始！")),
				tick(),
			)
		case key.Matches(keyMsg, m.taskListViewKeys.JoinRoom):
			m.roomUI = m.roomUI.ShowInput()
			return nil
		case key.Matches(keyMsg, m.taskListViewKeys.LeaveRoom):
			m.roomUI = m.roomUI.Hide()
			return nil
		}
	}

	var cmd tea.Cmd
	m.taskList, cmd = m.taskList.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}
