package main

import (
	"bufio"
	"flag"
	"gomato/task"
	"gomato/util"
	"os"
	"os/exec"
	"strings"
)

// clearConsole 清除终端屏幕
func clearConsole() {
	cmd := exec.Command("clear") // For Linux/macOS
	// cmd := exec.Command("cmd", "/c", "cls") // For Windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	// 解析命令行参数
	name := flag.String("name", "", "设置任务名称")
	flag.Parse()

	// 创建应用实例
	app := util.NewApp()
	taskManager := task.NewTaskManager()
	menu := NewMenu(app, taskManager)

	// 只有在提供了任务名称时才设置
	if *name != "" {
		app.SetTaskName(*name)
	}

	reader := bufio.NewReader(os.Stdin)

	// 启动主菜单
	for {
		clearConsole()
		menu.Display()

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		if !menu.HandleChoice(choice) {
			break
		}
	}
}
