// TUI界面主程序 - 使用Bubble Tea框架构建现代化的终端用户界面
package main

import (
	"fmt"
	"os"

	"gomato/pkg/gomato"
	"gomato/pkg/logging"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := logging.Init(); err != nil {
		fmt.Println("日志系统初始化失败:", err)
		os.Exit(1)
	}
	app := gomato.NewApp()
	if _, err := tea.NewProgram(app, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("运行程序时出错:", err)
		os.Exit(1)
	}
}
