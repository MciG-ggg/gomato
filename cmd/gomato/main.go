// TUI界面主程序 - 使用Bubble Tea框架构建现代化的终端用户界面
package main

import (
	"fmt"
	"os"

	"gomato/pkg/common"
	"gomato/pkg/gomato"
	"gomato/pkg/logging"
	"gomato/pkg/task"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/dig"
)

func main() {
	container := dig.New()

	// 先初始化 logging
	if err := logging.Init(); err != nil {
		fmt.Println("日志系统初始化失败:", err)
		os.Exit(1)
	}

	// 注册依赖并统一处理错误
	provides := []interface{}{
		func() (common.Settings, error) { return common.LoadSettings() },
		func() (*task.Manager, error) { return task.NewManager() },
		gomato.NewApp,
	}
	for _, p := range provides {
		if err := container.Provide(p); err != nil {
			fmt.Println("依赖注入注册失败:", err)
			os.Exit(1)
		}
	}

	// 启动主程序
	if err := container.Invoke(func(app *gomato.App) {
		if _, err := tea.NewProgram(app, tea.WithAltScreen()).Run(); err != nil {
			fmt.Println("运行程序时出错:", err)
			os.Exit(1)
		}
	}); err != nil {
		fmt.Println("依赖注入调用失败:", err)
		os.Exit(1)
	}
}
