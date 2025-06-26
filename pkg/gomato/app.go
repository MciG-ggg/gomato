package gomato

import (
	"fmt"
	"gomato/pkg/keymap"
	"gomato/pkg/task"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle           = lipgloss.NewStyle().Padding(1, 2)
	titleStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	statusMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).Render
)

const (
	taskListView = iota
	taskInputView
	timeView
	settingView
)

type viewState int

type App struct {
	currentView      viewState
	currentTaskIndex int
	taskManager      *task.Manager
	list             list.Model
	keys             *keymap.ListKeyMap
	delegateKeys     *keymap.DelegateKeyMap
	timeViewKeys     *keymap.TimeViewKeyMap
	timeModel        task.TimeModel
	settingModel     SettingModel
	taskInput        TaskInputModel
}

func NewApp() *App {
	delegateKeys := keymap.NewDelegateKeyMap()
	listKeys := keymap.NewListKeyMap()
	timeViewKeys := keymap.NewTimeViewKeyMap()
	taskManager, _ := task.NewManager()
	if len(taskManager.Tasks) == 0 {
		taskManager.AddItem("欢迎使用Gomato!", "这是一个番茄钟应用，希望能帮助你提高效率。")
	}
	for i := range taskManager.Tasks {
		taskManager.Tasks[i].Timer = task.TimeModel{
			TimerDuration:  25 * 60,
			TimerRemaining: 25 * 60,
			TimerIsRunning: false,
		}
	}
	items := make([]list.Item, len(taskManager.Tasks))
	for i, t := range taskManager.Tasks {
		items[i] = t
	}
	delegate := newItemDelegate(delegateKeys)
	taskList := list.New(items, delegate, 0, 0)
	taskList.Title = "番茄钟任务列表"
	taskList.Styles.Title = titleStyle
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
	taskTimeModel := task.TimeModel{
		TimerDuration:  25 * 60,
		TimerRemaining: 25 * 60,
		TimerIsRunning: false,
	}
	return &App{
		currentView:      taskListView,
		currentTaskIndex: 0,
		list:             taskList,
		taskInput:        NewTaskInputModel(),
		keys:             listKeys,
		delegateKeys:     delegateKeys,
		timeViewKeys:     timeViewKeys,
		taskManager:      taskManager,
		timeModel:        taskTimeModel,
		settingModel:     NewSettingModel(),
	}
}

func (m *App) Init() tea.Cmd {
	return nil
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case taskCreatedMsg:
		m.taskManager.AddItem(msg.title, msg.description)
		m.taskManager.Tasks[len(m.taskManager.Tasks)-1].Timer = task.TimeModel{
			TimerDuration:  25 * 60,
			TimerRemaining: 25 * 60,
			TimerIsRunning: false,
		}
		newTask := m.taskManager.Tasks[len(m.taskManager.Tasks)-1]
		insertCmd := m.list.InsertItem(len(m.list.Items()), newTask)
		statusCmd := m.list.NewStatusMessage(statusMessageStyle("添加了新任务: " + newTask.Title()))
		m.currentView = taskListView
		return m, tea.Batch(insertCmd, statusCmd)
	case backMsg:
		m.currentView = taskListView
		m.taskInput = NewTaskInputModel()
		return m, nil
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tickMsg:
		if m.timeModel.TimerIsRunning && m.timeModel.TimerRemaining > 0 {
			m.timeModel.TimerRemaining--
			if m.timeModel.TimerRemaining == 0 {
				m.timeModel.TimerIsRunning = false
				if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
					m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
					m.taskManager.Save()
				}
				return m, m.list.NewStatusMessage(statusMessageStyle("计时结束！"))
			}
			statusMsg := fmt.Sprintf("剩余时间: %02d:%02d", m.timeModel.TimerRemaining/60, m.timeModel.TimerRemaining%60)
			statusCmd := m.list.NewStatusMessage(statusMessageStyle(statusMsg))
			if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
				m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
				m.taskManager.Save()
			}
			return m, tea.Batch(statusCmd, tick())
		}
	case tea.KeyMsg:
		switch m.currentView {
		case taskListView:
			if m.list.FilterState() == list.Filtering {
				break
			}
			switch {
			case key.Matches(msg, m.keys.Setting):
				m.currentView = settingView
				return m, nil
			case key.Matches(msg, m.keys.ToggleTitleBar):
				v := !m.list.ShowTitle()
				m.list.SetShowTitle(v)
				m.list.SetShowFilter(v)
				m.list.SetFilteringEnabled(v)
				return m, nil
			case key.Matches(msg, m.keys.ToggleStatusBar):
				m.list.SetShowStatusBar(!m.list.ShowStatusBar())
				return m, nil
			case key.Matches(msg, m.keys.TogglePagination):
				m.list.SetShowPagination(!m.list.ShowPagination())
				return m, nil
			case key.Matches(msg, m.keys.ToggleHelpMenu):
				m.list.SetShowHelp(!m.list.ShowHelp())
				return m, nil
			case key.Matches(msg, m.keys.InsertItem):
				m.currentView = taskInputView
				return m, nil
			case key.Matches(msg, m.delegateKeys.Remove):
				index := m.list.Index()
				if index >= 0 && index < len(m.taskManager.Tasks) {
					deletedTaskTitle := m.taskManager.Tasks[index].Title()
					m.taskManager.DeleteItem(index)
					m.list.RemoveItem(index)
					if len(m.list.Items()) == 0 {
						m.delegateKeys.Remove.SetEnabled(false)
					}
					statusCmd := m.list.NewStatusMessage(statusMessageStyle("删除了任务: " + deletedTaskTitle))
					return m, statusCmd
				}
			case key.Matches(msg, m.keys.ChooseTask):
				m.currentTaskIndex = m.list.Index()
				if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
					m.timeModel = m.taskManager.Tasks[m.currentTaskIndex].Timer
				}
				m.currentView = timeView
				return m, m.list.NewStatusMessage(statusMessageStyle("任务已选择，请继续操作"))
			}
		case timeView:
			switch {
			case key.Matches(msg, m.timeViewKeys.Back):
				if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.taskManager.Tasks) {
					m.taskManager.Tasks[m.currentTaskIndex].Timer = m.timeModel
					m.taskManager.Save()
				}
				m.currentView = taskListView
				return m, nil
			case key.Matches(msg, m.timeViewKeys.StartPause):
				m.timeModel.TimerIsRunning = !m.timeModel.TimerIsRunning
				if m.timeModel.TimerIsRunning && m.timeModel.TimerRemaining > 0 {
					return m, tick()
				}
				return m, nil
			case key.Matches(msg, m.timeViewKeys.Reset):
				m.timeModel.TimerIsRunning = false
				m.timeModel.TimerRemaining = 25 * 60
				m.currentView = taskListView
				return m, nil
			}
		case taskInputView:
			var cmd tea.Cmd
			m.taskInput, cmd = m.taskInput.Update(msg)
			return m, cmd
		case settingView:
			var cmd tea.Cmd
			m.settingModel, cmd = m.settingModel.Update(msg)
			return m, cmd
		}
	}
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *App) View() string {
	switch m.currentView {
	case taskListView:
		return appStyle.Render(m.list.View())
	case timeView:
		min := m.timeModel.TimerRemaining / 60
		sec := m.timeModel.TimerRemaining % 60
		remainStr := fmt.Sprintf("%02d:%02d", min, sec)
		status := "已暂停"
		if m.timeModel.TimerIsRunning {
			status = "运行中"
		}
		controls := "[空格]开始/暂停  [r]重置  [q]返回"
		return appStyle.Render(
			titleStyle.Render("番茄钟计时器") + "\n\n" +
				"剩余时间: " + remainStr + "\n" +
				"状态: " + status + "\n\n" +
				statusMessageStyle(controls),
		)
	case taskInputView:
		return m.taskInput.View()
	case settingView:
		return m.settingModel.View()
	default:
		return ""
	}
}
