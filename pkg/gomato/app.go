package gomato

import (
	"fmt"
	"gomato/pkg/common"
	"gomato/pkg/keymap"
	"gomato/pkg/logging"
	"gomato/pkg/p2p"
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

// 自定义退出消息
type quitMsg struct{}

type App struct {
	currentView       viewState
	currentTaskIndex  int
	taskManager       *task.Manager
	taskList          list.Model
	taskListViewKeys  *keymap.ListKeyMap
	delegateKeys      *keymap.DelegateKeyMap
	timeViewKeys      *keymap.TimeViewKeyMap
	timeModel         task.TimeModel
	settingModel      SettingModel
	taskInput         TaskInputModel
	CurrentCycleCount int

	// p2p
	node        *p2p.Node
	roomManager *p2p.RoomManager
	roomUI      RoomUIModel
}

// 依赖注入构造函数
func NewApp(taskManager *task.Manager, settings common.Settings) *App {
	return NewAppWithKeyPath(taskManager, settings, "")
}

// 带密钥路径的构造函数
func NewAppWithKeyPath(taskManager *task.Manager, settings common.Settings, keyPath string) *App {
	delegateKeys := keymap.NewDelegateKeyMap()
	listKeys := keymap.NewListKeyMap()
	timeViewKeys := keymap.NewTimeViewKeyMap()

	if len(taskManager.Tasks) == 0 {
		taskManager.AddItem("欢迎使用Gomato!", "这是一个番茄钟应用，希望能帮助你提高效率。")
	}

	// 用注入的 settings 构造 SettingModel
	settingModel := NewSettingModelWithSettings(settings)
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
		IsWorkSession:  true,
	}
	// 初始化P2P节点
	node, err := p2p.NewNode(keyPath)
	if err != nil {
		logging.Log(fmt.Sprintf("Failed to create P2P node: %v", err))
	}

	// 初始化房间UI
	roomUI := NewRoomUIModel(node.GetRoomMgr())

	return &App{
		currentView:       taskListView,
		currentTaskIndex:  0,
		taskList:          taskList,
		taskInput:         NewTaskInputModel(),
		taskListViewKeys:  listKeys,
		delegateKeys:      delegateKeys,
		timeViewKeys:      timeViewKeys,
		taskManager:       taskManager,
		timeModel:         taskTimeModel,
		settingModel:      settingModel,
		CurrentCycleCount: 0,

		// p2p
		node:        node,
		roomManager: node.GetRoomMgr(),
		roomUI:      roomUI,
	}
}

func (m *App) Init() tea.Cmd {
	return nil
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 处理Ctrl+C退出
		if msg.Type == tea.KeyCtrlC {
			// 清理P2P状态
			if m.roomUI.IsInRoom() {
				m.roomUI.roomManager.LeaveRoom()
			}
			if m.node != nil {
				m.node.Close()
			}
			return m, tea.Quit
		}
	case quitMsg:
		// 处理自定义退出消息
		if m.roomUI.IsInRoom() {
			m.roomUI.roomManager.LeaveRoom()
		}
		if m.node != nil {
			m.node.Close()
		}
		return m, tea.Quit
	case taskCreatedMsg:
		return handleTaskCreated(m, msg)
	case backMsg:
		return handleBack(m)
	case tea.WindowSizeMsg:
		h, v := common.AppStyle.GetFrameSize()
		m.taskList.SetSize(msg.Width-h, msg.Height-v)
		// 为房间UI的成员列表设置大小
		m.roomUI.memberList.SetSize(msg.Width-h, msg.Height-v-6) // 减去状态栏和边框的高度
		m.settingModel, _ = m.settingModel.Update(msg)
		return m, nil
	case tickMsg:
		return m, handleTick(m)
	}

	// 处理房间UI更新
	var roomCmd tea.Cmd
	m.roomUI, roomCmd = m.roomUI.Update(msg)

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

	return m, tea.Batch(roomCmd, cmd)
}

func (m *App) View() string {
	var mainView string
	switch m.currentView {
	case taskListView:
		mainView = common.AppStyle.Render(m.taskList.View())
	case timeView:
		mainView = common.AppStyle.Render(m.timeModel.ViewWithSettings(&m.settingModel.Settings))
	case taskInputView:
		mainView = common.AppStyle.Render(m.taskInput.View())
	case settingView:
		mainView = m.settingModel.View()
	default:
		mainView = ""
	}

	// 如果房间UI可见，则显示房间UI
	roomView := m.roomUI.View()
	if roomView != "" {
		return roomView
	}

	return mainView
}
