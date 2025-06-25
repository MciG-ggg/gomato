// 注释掉的原始命令行界面代码
// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"gomato/pkg/app"
// 	"gomato/pkg/menu"
// 	"gomato/pkg/setting"
// 	"gomato/pkg/task"
// 	"os"
// 	"os/exec"
// 	"runtime"
// 	"strings"
// )

// // clearConsole 根据系统不同,清除终端屏幕
// func clearConsole() {
// 	var cmd *exec.Cmd

// 	switch runtime.GOOS {
// 	case "windows":
// 		cmd = exec.Command("cmd", "/c", "cls")
// 	case "linux", "darwin":
// 		cmd = exec.Command("clear")
// 	default:
// 		fmt.Println("不支持的操作系统:", runtime.GOOS)
// 		return
// 	}

// 	cmd.Stdout = os.Stdout
// 	err := cmd.Run()
// 	if err != nil {
// 		fmt.Println("清屏失败:", err)
// 	}
// }

// func main() {
// 	// Load settings
// 	err := setting.Load()
// 	if err != nil {
// 		fmt.Println("加载设置失败:", err)
// 		return
// 	}

// 	// 创建应用实例
// 	a := app.NewApp()
// 	tm := task.NewTaskManager()
// 	menu := menu.NewMenu(a, tm)

// 	reader := bufio.NewReader(os.Stdin)

// 	// 启动主菜单
// 	for {
// 		clearConsole()
// 		menu.Display()

// 		input, _ := reader.ReadString('\n')
// 		choice := strings.TrimSpace(input)

//			if !menu.HandleChoice(choice) {
//				break
//			}
//		}
//	}

// TUI界面主程序 - 使用Bubble Tea框架构建现代化的终端用户界面
package main

import (
	"fmt"
	"os"

	"gomato/pkg/keymap"
	"gomato/pkg/task"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 全局样式定义 - 使用lipgloss库定义终端UI样式
var (
	// 应用整体样式 - 添加内边距，美化界面布局
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	// 标题样式 - 白色文字，绿色背景，突出显示标题
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	// 状态消息样式 - 自适应颜色，支持浅色/深色主题
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

const (
	// viewState 定义当前视图状态 - 用于切换不同的UI视图
	taskListView = iota // 0: 默认视图
	taskInputView
	timeView // 2: 时间视图（番茄钟计时器）
)

// viewState 定义视图状态的类型
type viewState int

// model 定义应用程序的数据模型 - Bubble Tea的核心数据结构
type model struct {
	// general
	currentView viewState
	taskManager *task.Manager

	// views
	list         list.Model             // 列表组件，显示任务列表
	taskInput    taskInputModel         // 任务输入视图
	delegateKeys *keymap.DelegateKeyMap // 列表项的委托按键映射

	// keys
	keys         *keymap.ListKeyMap     // 列表操作的按键映射
	timeViewKeys *keymap.TimeViewKeyMap // 番茄钟视图按键映射
}

// newModel 创建并返回初始化的模型实例
// 设置初始数据、样式和按键绑定
func newModel() model {
	var (
		delegateKeys = keymap.NewDelegateKeyMap() // 创建委托按键映射
		listKeys     = keymap.NewListKeyMap()     // 创建列表按键映射
		timeViewKeys = keymap.NewTimeViewKeyMap() // 创建番茄钟按键映射
	)

	// 创建任务管理器
	taskManager, err := task.NewManager()
	if err != nil {
		fmt.Println("创建任务管理器失败:", err)
		os.Exit(1)
	}

	// 如果没有任务，创建一个欢迎任务
	if len(taskManager.Tasks) == 0 {
		taskManager.AddItem("欢迎使用Gomato!", "这是一个番茄钟应用，希望能帮助你提高效率。")
	}

	// 将 task.Task 转换为 list.Item
	items := make([]list.Item, len(taskManager.Tasks))
	for i, t := range taskManager.Tasks {
		items[i] = t
	}

	// 设置列表组件
	delegate := newItemDelegate(delegateKeys) // 创建项目委托（在taskList.go中定义）
	groceryList := list.New(items, delegate, 0, 0)
	groceryList.Title = "番茄钟任务列表"         // 设置列表标题
	groceryList.Styles.Title = titleStyle // 应用标题样式
	groceryList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.ToggleSpinner,
			listKeys.InsertItem,
			listKeys.ToggleTitleBar,
			listKeys.ToggleStatusBar,
			listKeys.TogglePagination,
			listKeys.ToggleHelpMenu,
		}
	}

	return model{
		currentView:  taskListView, // 设置初始视图为任务列表视图
		list:         groceryList,
		taskInput:    NewTaskInputModel(),
		keys:         listKeys,
		delegateKeys: delegateKeys,
		timeViewKeys: timeViewKeys,
		taskManager:  taskManager,
	}
}

// Init 初始化模型，返回初始命令
// Bubble Tea生命周期方法，在程序启动时调用
func (m model) Init() tea.Cmd {
	return nil // 不需要初始命令
}

// Update 处理消息并更新模型
// Bubble Tea的核心方法，处理所有用户输入和系统事件
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case taskCreatedMsg:
		m.taskManager.AddItem(msg.title, msg.description)
		newTask := m.taskManager.Tasks[len(m.taskManager.Tasks)-1]
		insertCmd := m.list.InsertItem(len(m.list.Items()), newTask)
		statusCmd := m.list.NewStatusMessage(statusMessageStyle("添加了新任务: " + newTask.Title()))
		m.currentView = taskListView
		return m, tea.Batch(insertCmd, statusCmd)
	case backMsg:
		m.currentView = taskListView
		m.taskInput = NewTaskInputModel() // Reset the form
		return m, nil

	case tea.WindowSizeMsg:
		// 处理窗口大小变化事件
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch m.currentView {
		case taskListView:
			// 处理任务列表视图的按键消息
			// 如果正在过滤模式，不处理其他按键操作
			if m.list.FilterState() == list.Filtering {
				break
			}

			// 处理各种按键操作
			switch {
			case key.Matches(msg, m.keys.ToggleSpinner):
				// 切换加载动画显示
				cmd := m.list.ToggleSpinner()
				return m, cmd

			case key.Matches(msg, m.keys.ToggleTitleBar):
				// 切换标题栏显示，同时控制过滤功能
				v := !m.list.ShowTitle()
				m.list.SetShowTitle(v)
				m.list.SetShowFilter(v)
				m.list.SetFilteringEnabled(v)
				return m, nil

			case key.Matches(msg, m.keys.ToggleStatusBar):
				// 切换状态栏显示
				m.list.SetShowStatusBar(!m.list.ShowStatusBar())
				return m, nil

			case key.Matches(msg, m.keys.TogglePagination):
				// 切换分页显示
				m.list.SetShowPagination(!m.list.ShowPagination())
				return m, nil

			case key.Matches(msg, m.keys.ToggleHelpMenu):
				// 切换帮助菜单显示
				m.list.SetShowHelp(!m.list.ShowHelp())
				return m, nil

			case key.Matches(msg, m.keys.InsertItem):
				m.currentView = taskInputView
				return m, nil

			case key.Matches(msg, m.delegateKeys.Remove):
				// 处理删除任务操作
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
				// 处理选择任务操作
				m.currentView = timeView // 切换到时间视图
				return m, m.list.NewStatusMessage(statusMessageStyle("任务已选择，请继续操作"))
			}
		case timeView:
			// 处理时间视图的按键消息
			switch {
			case key.Matches(msg, m.timeViewKeys.Back):
				m.currentView = taskListView
				return m, nil
			case key.Matches(msg, m.timeViewKeys.StartPause):
				// 你的开始/暂停逻辑
				return m, nil
			case key.Matches(msg, m.timeViewKeys.Reset):
				// 你的重置逻辑
				return m, nil
			}
		case taskInputView:
			var cmd tea.Cmd
			m.taskInput, cmd = m.taskInput.Update(msg)
			return m, cmd
		}
	}

	// 更新列表模型（这会同时调用委托的update函数）
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View 渲染模型到字符串
// Bubble Tea生命周期方法，负责将数据模型转换为显示字符串
func (m model) View() string {
	switch m.currentView {
	case taskListView:
		return appStyle.Render(m.list.View()) // 应用样式并返回渲染结果
	case timeView:
		// 时间视图渲染逻辑
		return appStyle.Render(
			titleStyle.Render("番茄钟计时器") + "\n\n" +
				"（此处显示番茄钟计时界面，功能开发中...）\n\n" +
				statusMessageStyle("按 q 返回任务列表"),
		)
	case taskInputView:
		return m.taskInput.View()
	default:
		return "" // 未知视图状态，返回空字符串
	}
}

// main 程序入口点
// 创建并启动Bubble Tea程序
func main() {
	// 创建并运行Bubble Tea程序
	// tea.WithAltScreen() 启用备用屏幕模式，提供更好的显示效果
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("运行程序时出错:", err)
		os.Exit(1)
	}
}
