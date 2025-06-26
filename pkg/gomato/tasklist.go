package gomato

import (
	"gomato/pkg/common"
	"gomato/pkg/keymap"
	"gomato/pkg/task"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

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
		}
	}
	return taskList
}

func handleTaskCreated(m *App, msg taskCreatedMsg) (tea.Model, tea.Cmd) {
	m.taskManager.AddItem(msg.title, msg.description)
	m.taskManager.Tasks[len(m.taskManager.Tasks)-1].Timer = task.TimeModel{
		TimerDuration:  25 * 60,
		TimerRemaining: 25 * 60,
		TimerIsRunning: false,
		IsWorkSession:  true,
	}
	newTask := m.taskManager.Tasks[len(m.taskManager.Tasks)-1]
	insertCmd := m.list.InsertItem(len(m.list.Items()), newTask)
	statusCmd := m.list.NewStatusMessage(statusMessageStyle("添加了新任务: " + newTask.Title()))
	m.currentView = taskListView
	return m, tea.Batch(insertCmd, statusCmd)
}

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

func updateTaskListView(m *App, msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch keyMsg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch {
		case key.Matches(keyMsg, m.keys.Setting):
			common.LoadSettings()
			m.settingModel.ReloadInputsFromSettings()
			m.currentView = settingView
			return nil
		case key.Matches(keyMsg, m.keys.ToggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return nil
		case key.Matches(keyMsg, m.keys.ToggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return nil
		case key.Matches(keyMsg, m.keys.TogglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return nil
		case key.Matches(keyMsg, m.keys.ToggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return nil
		case key.Matches(keyMsg, m.keys.InsertItem):
			m.currentView = taskInputView
			return nil
		case key.Matches(keyMsg, m.delegateKeys.Remove):
			index := m.list.Index()
			if index >= 0 && index < len(m.taskManager.Tasks) {
				deletedTaskTitle := m.taskManager.Tasks[index].Title()
				m.taskManager.DeleteItem(index)
				m.list.RemoveItem(index)
				if len(m.list.Items()) == 0 {
					m.delegateKeys.Remove.SetEnabled(false)
				}
				statusCmd := m.list.NewStatusMessage(statusMessageStyle("删除了任务: " + deletedTaskTitle))
				return statusCmd
			}
		case key.Matches(keyMsg, m.keys.ChooseTask):
			m.currentTaskIndex = m.list.Index()
			if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
				m.timeModel = m.taskManager.Tasks[m.currentTaskIndex].Timer
			}
			m.currentView = timeView
			m.timeModel.TimerIsRunning = true
			return tea.Batch(
				m.list.NewStatusMessage(statusMessageStyle("任务已选择，计时已开始！")),
				tick(),
			)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}
