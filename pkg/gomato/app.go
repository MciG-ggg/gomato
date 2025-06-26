package gomato

import (
	"gomato/pkg/common"
	"gomato/pkg/keymap"
	"gomato/pkg/task"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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
	settingModel := NewSettingModel()
	for i := range taskManager.Tasks {
		taskManager.Tasks[i].Timer = task.TimeModel{
			TimerDuration:  int(settingModel.Settings.Pomodoro) * 60,
			TimerRemaining: int(settingModel.Settings.Pomodoro) * 60,
			TimerIsRunning: false,
		}
	}
	taskList := NewTaskList(listKeys, delegateKeys, taskManager)
	taskTimeModel := task.TimeModel{
		TimerDuration:  int(settingModel.Settings.Pomodoro) * 60,
		TimerRemaining: int(settingModel.Settings.Pomodoro) * 60,
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
		settingModel:     settingModel,
	}
}

func (m *App) Init() tea.Cmd {
	return nil
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case taskCreatedMsg:
		return handleTaskCreated(m, msg)
	case backMsg:
		return handleBack(m)
	case tea.WindowSizeMsg:
		h, v := common.AppStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.settingModel, _ = m.settingModel.Update(msg)
		return m, nil
	case tickMsg:
		return m, handleTick(m)
	}
	var cmd tea.Cmd
	switch m.currentView {
	case taskListView:
		cmd = updateTaskListView(m, msg)
	case timeView:
		cmd = updateTimeView(m, msg)
	case taskInputView:
		m.taskInput, cmd = m.taskInput.Update(msg)
	case settingView:
		m.settingModel, cmd = m.settingModel.Update(msg)
	}

	return m, cmd
}

func (m *App) View() string {
	switch m.currentView {
	case taskListView:
		return common.AppStyle.Render(m.list.View())
	case timeView:
		return common.AppStyle.Render(m.timeModel.View())
	case taskInputView:
		return common.AppStyle.Render(m.taskInput.View())
	case settingView:
		return m.settingModel.View()
	default:
		return ""
	}
}
