package util

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"gomato/common"
	"gomato/config"
	"gomato/task"
	"gomato/timer"
)

type App struct {
	timer *timer.Timer
	tasks *task.TaskManager
	cfg   *config.TaskConfig
	task  *timer.Task
}

func NewApp() *App {
	// read config from task.json
	// 如果没有任务，则使用默认配置
	// 如果有任务，则使用任务中的配置
	cfg := config.DefaultTaskConfig
	loadedTask, _ := timer.LoadTask()

	if loadedTask != nil {
		cfg.TaskName = loadedTask.Name
		cfg.WorkDuration = loadedTask.WorkTime
		cfg.BreakDuration = loadedTask.BreakTime
	}

	app := &App{
		timer: timer.NewTimer(cfg, false),
		tasks: task.NewTaskManager(),
		cfg:   &cfg,
		task:  loadedTask,
	}

	if loadedTask != nil {
		app.timer.UpdateTask(loadedTask.Name, loadedTask.WorkTime, loadedTask.BreakTime)
	}

	return app
}

func (a *App) GetConfig() *config.TaskConfig {
	return a.cfg
}

func (a *App) SetTask(task *timer.Task) {
	a.task = task
	a.timer.UpdateTask(task.Name, task.WorkTime, task.BreakTime)
	a.cfg.TaskName = task.Name
	a.cfg.WorkDuration = task.WorkTime
	a.cfg.BreakDuration = task.BreakTime
}

func (a *App) GetTask() *timer.Task {
	return a.task
}

func (a *App) UpdateTask(name string, workTime, breakTime time.Duration) error {
	if a.task == nil {
		a.task = &timer.Task{
			Name:      name,
			WorkTime:  workTime,
			BreakTime: breakTime,
		}
	} else {
		a.task.Name = name
		a.task.WorkTime = workTime
		a.task.BreakTime = breakTime
	}

	a.cfg.TaskName = name
	a.cfg.WorkDuration = workTime
	a.cfg.BreakDuration = breakTime

	if err := a.timer.UpdateTask(name, workTime, breakTime); err != nil {
		return fmt.Errorf("更新计时器任务失败：%v", err)
	}

	if err := timer.SaveTask(a.task); err != nil {
		return fmt.Errorf("保存任务失败：%v", err)
	}

	return nil
}

func (a *App) Run() {
	fmt.Printf("%s%s🍅 欢迎使用番茄钟计时器！%s\n", common.Bold, common.Cyan, common.Reset)
	fmt.Printf("%s输入 'help' 查看可用命令和快捷键%s\n", common.Yellow, common.Reset)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	cmdChan := make(chan string)
	go a.handleCommands(cmdChan)

	// 创建新的 Timer 实例
	a.timer = timer.NewTimer(*a.cfg, false)
	go a.timer.Start()

	for {
		select {
		case <-sigChan:
			a.timer.Stop()
			return
		case cmd := <-cmdChan:
			if cmd == "quit" {
				a.timer.Stop()
				return
			}
			a.processCommand(cmd)
		}
	}
}

func (a *App) handleCommands(cmdChan chan<- string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			fmt.Printf("\r\033[K")
			if shortcut, exists := shortcuts[cmd]; exists {
				cmdChan <- shortcut
			} else {
				cmdChan <- cmd
			}
		}
	}
}

func (a *App) processCommand(cmd string) {
	switch {
	case cmd == "y":
		a.timer.StartRest()
	case cmd == "n":
		a.timer.EndRest()
	case cmd == "help":
		printHelp()
	case cmd == "start":
		a.timer.TriggerStart()
		fmt.Printf("%s开始计时！%s\n", common.Green, common.Reset)
	case cmd == "stats":
		a.printStats()
	case cmd == "tasks":
		a.printTasks()
	case cmd == "pause":
		a.timer.Pause()
	case cmd == "resume":
		a.timer.Resume()
	case strings.HasPrefix(cmd, "add "):
		a.addTask(strings.TrimPrefix(cmd, "add "))
	case strings.HasPrefix(cmd, "complete "):
		a.completeTask(strings.TrimPrefix(cmd, "complete "))
	case strings.HasPrefix(cmd, "work "):
		a.setWorkDuration(strings.TrimPrefix(cmd, "work "))
	case strings.HasPrefix(cmd, "break "):
		a.setBreakDuration(strings.TrimPrefix(cmd, "break "))
	case strings.HasPrefix(cmd, "name "):
		name := strings.TrimPrefix(cmd, "name ")
		a.SetTaskName(name)
		fmt.Printf("%s当前任务名称已设置为：%s%s\n", common.Blue, name, common.Reset)
	}
}

func (a *App) addTask(description string) {
	a.tasks.AddTask(description)
	fmt.Printf("%s已添加任务: %s%s\n", common.Green, description, common.Reset)
	a.saveTasks()
}

func (a *App) completeTask(idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Printf("%s无效的任务ID%s\n", common.Red, common.Reset)
		return
	}
	if err := a.tasks.CompleteTask(id); err != nil {
		fmt.Printf("%s%s%s\n", common.Red, err, common.Reset)
	} else {
		fmt.Printf("%s任务 %d 已完成%s\n", common.Green, id, common.Reset)
		a.saveTasks()
	}
}

func (a *App) setWorkDuration(duration string) {
	if d, err := time.ParseDuration(duration); err == nil {
		a.cfg.SetWork(d)
		if a.task != nil {
			a.task.WorkTime = d
			a.timer.UpdateTask(a.task.Name, a.task.WorkTime, a.task.BreakTime)
		} else {
			a.timer.UpdateTask(a.cfg.TaskName, a.cfg.WorkDuration, a.cfg.BreakDuration)
		}

		fmt.Printf("%s工作时间已设置为 %v%s\n", common.Blue, d, common.Reset)
	} else {
		fmt.Printf("%s无效的时间格式%s\n", common.Red, common.Reset)
	}
}

func (a *App) setBreakDuration(duration string) {
	if d, err := time.ParseDuration(duration); err == nil {
		a.cfg.SetBreak(d)
		if a.task != nil {
			a.task.BreakTime = d
			a.timer.UpdateTask(a.task.Name, a.task.WorkTime, a.task.BreakTime)
		} else {
			a.timer.UpdateTask(a.cfg.TaskName, a.cfg.WorkDuration, a.cfg.BreakDuration)
		}

		fmt.Printf("%s休息时间已设置为 %v%s\n", common.Blue, d, common.Reset)
	} else {
		fmt.Printf("%s无效的时间格式%s\n", common.Red, common.Reset)
	}
}

func (a *App) printStats() {
	stats := a.timer.GetStats()
	totalTasks, completedTasks := a.tasks.GetTaskStats()

	fmt.Printf("\n%s统计信息：%s\n", common.Bold, common.Reset)
	fmt.Printf("%s当前任务：%s%s%s\n", common.Cyan, common.White, a.cfg.TaskName, common.Reset)
	fmt.Printf("%s工作时长：%s%v%s\n", common.Cyan, common.White, a.cfg.WorkDuration, common.Reset)
	fmt.Printf("%s休息时长：%s%v%s\n", common.Cyan, common.White, a.cfg.BreakDuration, common.Reset)
	fmt.Printf("%s工作会话数：%s%d%s\n", common.Cyan, common.White, stats.WorkSessions, common.Reset)
	fmt.Printf("%s休息会话数：%s%d%s\n", common.Cyan, common.White, stats.BreakSessions, common.Reset)
	fmt.Printf("%s总工作时间：%s%v%s\n", common.Cyan, common.White, stats.TotalWorkTime, common.Reset)
	fmt.Printf("%s任务完成率：%s%d/%d%s\n", common.Cyan, common.White, completedTasks, totalTasks, common.Reset)
}

func (a *App) printTasks() {
	tasks := a.tasks.ListTasks()
	if len(tasks) == 0 {
		fmt.Printf("\n%s当前没有任务%s\n", common.Yellow, common.Reset)
		return
	}

	fmt.Printf("\n%s任务列表：%s\n", common.Bold, common.Reset)
	for _, task := range tasks {
		status := "进行中"
		statusColor := common.Yellow
		if task.Completed {
			status = "已完成"
			statusColor = common.Green
		}
		fmt.Printf("%s[%d]%s %s - %s%s%s\n",
			common.Blue, task.ID, common.Reset,
			task.Description,
			statusColor, status, common.Reset)
	}
}

func (a *App) quit() {
	fmt.Printf("\n%s正在退出程序...%s\n", common.Yellow, common.Reset)
	a.timer.Stop()
}

func (a *App) SetTaskName(name string) {
	a.cfg.TaskName = name
	currentWorkTime := a.cfg.WorkDuration
	currentBreakTime := a.cfg.BreakDuration

	if a.task != nil {
		a.task.Name = name
		currentWorkTime = a.task.WorkTime
		currentBreakTime = a.task.BreakTime
	}

	a.timer.UpdateTask(name, currentWorkTime, currentBreakTime)
}

func (a *App) saveTasks() {
	if err := a.tasks.Save(); err != nil {
		fmt.Printf("%s保存任务失败：%v%s\n", common.Red, err, common.Reset)
	}
}

func (a *App) SetTimer(t *timer.Timer) {
	a.timer = t
}

var shortcuts = map[string]string{
	"p": "pause",
	"r": "resume",
	"s": "stats",
	"t": "tasks",
	"h": "help",
	"q": "quit",
	"g": "start",
}

func printHelp() {
	fmt.Printf("\n%s可用命令：%s\n", common.Bold, common.Reset)
	fmt.Printf("%s  help              - %s显示帮助信息%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  start             - %s开始计时%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  stats             - %s显示统计信息%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  tasks             - %s显示任务列表%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  add <描述>        - %s添加新任务%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  complete <ID>     - %s完成任务%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  work <时间>       - %s设置工作时间（例如：25m）%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  break <时间>      - %s设置休息时间（例如：5m）%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  name <名称>       - %s设置当前任务名称%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  pause             - %s暂停当前计时%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  resume            - %s恢复计时%s\n", common.Blue, common.White, common.Reset)

	fmt.Printf("\n%s快捷键：%s\n", common.Bold, common.Reset)
	fmt.Printf("%s  g                 - %s开始计时%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  p                 - %s暂停计时%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  r                 - %s恢复计时%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  s                 - %s显示统计%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  t                 - %s显示任务%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  h                 - %s显示帮助%s\n", common.Blue, common.White, common.Reset)
	fmt.Printf("%s  q                 - %s退出程序%s\n", common.Blue, common.White, common.Reset)
}
