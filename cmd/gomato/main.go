package main

import (
	"bufio"
	"gomato/pkg/app"
	"gomato/pkg/menu"
	"gomato/pkg/task"
	"os"
	"strings"
)

func main() {
	a := app.NewApp()
	taskManager := task.NewTaskManager()
	m := menu.NewMenu(a, taskManager)
	reader := bufio.NewReader(os.Stdin)

	for {
		m.Display()
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if !m.HandleChoice(choice) {
			break
		}
	}
}
