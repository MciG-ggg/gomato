package main

import (
	"bufio"
	"fmt"
	"gomato/setting"
	"gomato/task"
	"gomato/util"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// clearConsole 根据系统不同,清除终端屏幕
func clearConsole() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	case "linux", "darwin":
		cmd = exec.Command("clear")
	default:
		fmt.Println("不支持的操作系统:", runtime.GOOS)
		return
	}

	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("清屏失败:", err)
	}
}

func main() {
	// Load settings
	err := setting.Load()
	if err != nil {
		fmt.Println("加载设置失败:", err)
		return
	}

	// 创建应用实例
	// TODO: should load setting
	app := util.NewApp()
	taskManager := task.NewTaskManager()
	menu := NewMenu(app, taskManager)

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
