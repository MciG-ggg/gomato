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
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	menuView viewState = iota
	timeView
)

type displayMode int

const (
	normal displayMode = iota
	ansiArt
)

var modeNames = []string{
	"normal",
	"ansiArt",
}

type model struct {
	// General state
	currentView viewState

	// Time view state
	// TODO: should load from the json or timer struct(?)
	mode displayMode

	// Menu view state
	menuItems    []string
	selectedItem int
	menuCursor   int
}

func initialModel() model {
	return model{
		currentView: menuView,
		mode:        normal,
		menuItems: []string{
			"添加新任务",
			"查看所有任务",
			"查看最近任务",
			"开始专注任务",
			"标记任务完成",
			"删除任务",
			"清空所有任务",
			"设置",
			"退出",
		},
		selectedItem: -1, // -1 means nothing is selected
		menuCursor:   0,
	}
}

func (m model) Init() tea.Cmd {
	// Initialize the model, if needed
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.currentView == menuView {
				m.currentView = timeView
			} else {
				m.currentView = menuView
			}
		}
		switch m.currentView {
		case menuView:
			switch msg.String() {
			case "up", "k":
				if m.menuCursor > 0 {
					m.menuCursor--
				}
			case "down", "j":
				if m.menuCursor < len(m.menuItems)-1 {
					m.menuCursor++
				}
			case "enter":
				// TODO: deal with entering the selected menu item
			}
		case timeView:
			m.mode = normal // Reset mode when switching to time view
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case menuView:
		var sb strings.Builder
		sb.WriteString("=== 番茄钟任务管理器 ===\n")
		for i, item := range m.menuItems {
			if i == m.menuCursor {
				sb.WriteString("> " + item + "\n")
			} else {
				sb.WriteString("  " + item + "\n")
			}
		}
		return sb.String()
	case timeView:
		return "Time View: Mode is " + modeNames[m.mode] + "\nPress 'tab' to switch back to menu."
	default:
		return "Unknown view"
	}
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("出错:", err)
		os.Exit(1)
	}
}
