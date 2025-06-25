package menu

import (
	"bufio"
	"fmt"
	"gomato/pkg/app"
	"gomato/pkg/common"
	"gomato/pkg/setting"
	"gomato/pkg/task"
	"gomato/pkg/timer"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Menu 表示应用程序的主菜单
type Menu struct {
	app         *app.App
	taskManager *task.TaskManager
	reader      *bufio.Reader
}

// NewMenu 创建一个新的菜单实例
func NewMenu(app *app.App, taskManager *task.TaskManager) *Menu {
	return &Menu{
		app:         app,
		taskManager: taskManager,
		reader:      bufio.NewReader(os.Stdin),
	}
}

// clearScreen clears the terminal screen
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (m *Menu) Display() {
	clearScreen()
	fmt.Printf("%s%s🍅 番茄钟任务管理器 %s\n\n", common.Bold, common.Cyan, common.Reset)
	fmt.Printf("  %s1.%s 添加新任务\n", common.Yellow, common.Reset)
	fmt.Printf("  %s2.%s 查看所有任务\n", common.Yellow, common.Reset)
	fmt.Printf("  %s3.%s 查看最近任务\n", common.Yellow, common.Reset)
	fmt.Printf("  %s4.%s 开始专注任务\n", common.Yellow, common.Reset)
	fmt.Printf("  %s5.%s 标记任务完成\n", common.Yellow, common.Reset)
	fmt.Printf("  %s6.%s 删除任务\n", common.Yellow, common.Reset)
	fmt.Printf("  %s7.%s 清空所有任务\n", common.Yellow, common.Reset)
	fmt.Printf("  %s8.%s setting 设置\n", common.Yellow, common.Reset)
	fmt.Printf("  %s9.%s 退出\n\n", common.Yellow, common.Reset)
	fmt.Printf("%s请选择操作 (1-9): %s", common.Bold, common.Reset)
}

func (m *Menu) HandleChoice(choice string) bool {
	switch choice {
	case "1":
		m.handleAddTask()
	case "2":
		m.handleListTasks()
	case "3":
		m.handleRecentTasks()
	case "4":
		m.handleStartTask()
	case "5":
		m.handleCompleteTask()
	case "6":
		m.handleDeleteTask()
	case "7":
		m.handleClearAllTasks()
	case "8":
		fmt.Println("\n进入设置菜单...")
		m.modifySettings()
	case "9":
		fmt.Println("\n感谢使用！再见！")
		return false
	default:
		fmt.Println("\n无效的选择，请重试。")
	}
	return true
}

func (m *Menu) handleAddTask() {
	fmt.Print("请输入任务描述: ")
	description, _ := m.reader.ReadString('\n')
	description = strings.TrimSpace(description)
	if description == "" {
		fmt.Println("任务描述不能为空！")
	} else {
		m.taskManager.AddTask(description)
		fmt.Println("任务已添加！")
		fmt.Print("按 Enter 键继续...")
		m.reader.ReadString('\n')
	}
}

func (m *Menu) handleListTasks() {
	tasks := m.taskManager.ListTasks()
	if len(tasks) == 0 {
		fmt.Println("\n暂无任务")
	} else {
		fmt.Println("\n============= 所有任务 =============")
		for _, t := range tasks {
			status := "未完成"
			if t.Completed {
				status = "已完成"
			}
			fmt.Printf("ID: %-4d 描述: %-30s 状态: %s\n", t.ID, t.Description, status)
		}
		fmt.Println("====================================")
	}
	fmt.Print("按 Enter 键继续...")
	m.reader.ReadString('\n')
}

func (m *Menu) handleRecentTasks() {
	recentTasks := m.taskManager.GetRecentTasks(5)
	if len(recentTasks) == 0 {
		fmt.Println("\n暂无最近任务")
	} else {
		fmt.Println("\n============= 最近任务 =============")
		for _, t := range recentTasks {
			status := "未完成"
			if t.Completed {
				status = "已完成"
			}
			fmt.Printf("ID: %-4d 描述: %-30s 状态: %s\n", t.ID, t.Description, status)
		}
		fmt.Println("====================================")
	}
	fmt.Print("按 Enter 键继续...")
	m.reader.ReadString('\n') // 等待用户输入, 读取并丢弃用户的一次输入（直到按下回车)
}

func (m *Menu) handleStartTask() {
	incompleteTasks := m.taskManager.GetIncompleteTasks()
	if len(incompleteTasks) == 0 {
		fmt.Println("\n没有未完成的任务，请先添加任务！")
		return
	}

	fmt.Println("\n=========== 未完成的任务 ===========")
	for _, t := range incompleteTasks {
		fmt.Printf("ID: %-4d 描述: %s\n", t.ID, t.Description)
	}
	fmt.Println("====================================")

	fmt.Print("请输入要开始的任务ID: ")
	idStr, _ := m.reader.ReadString('\n')
	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		fmt.Println("\n无效的任务ID，请输入数字。")
		return
	}

	selectedTask, err := m.taskManager.GetTaskByID(id)
	if err != nil || selectedTask.Completed {
		fmt.Printf("\n任务 %d 不存在或已完成，请选择未完成的任务ID。\n", id)
		return
	}

	m.app.SetTaskName(selectedTask.Description)

	// 在开始任务前询问是否修改工作时间和休息时间
	fmt.Printf("\n%s当前工作时间：%s%v%s\n", common.Cyan, common.White, m.app.GetConfig().WorkDuration, common.Reset)
	fmt.Print("是否修改工作时间？(请输入如 '25m' 或直接按 Enter 跳过): ")
	workTimeInput, _ := m.reader.ReadString('\n')
	workTimeInput = strings.TrimSpace(workTimeInput)
	if workTimeInput != "" {
		if d, err := time.ParseDuration(workTimeInput); err == nil {
			m.app.GetConfig().SetWork(d)
			fmt.Printf("%s工作时间已设置为 %v%s\n", common.Blue, d, common.Reset)
		} else {
			fmt.Printf("%s无效的工作时间格式：%v%s\n", common.Red, err, common.Reset)
		}
	}

	fmt.Printf("\n%s当前休息时间：%s%v%s\n", common.Cyan, common.White, m.app.GetConfig().BreakDuration, common.Reset)
	fmt.Print("是否修改休息时间？(请输入如 '5m' 或直接按 Enter 跳过): ")
	breakTimeInput, _ := m.reader.ReadString('\n')
	breakTimeInput = strings.TrimSpace(breakTimeInput)
	if breakTimeInput != "" {
		if d, err := time.ParseDuration(breakTimeInput); err == nil {
			m.app.GetConfig().SetBreak(d)
			fmt.Printf("%s休息时间已设置为 %v%s\n", common.Blue, d, common.Reset)
		} else {
			fmt.Printf("%s无效的休息时间格式：%v%s\n", common.Red, err, common.Reset)
		}
	}

	// 更新计时器中的任务和配置
	if m.app.GetTask() == nil {
		// 如果没有加载的任务，则创建一个新的任务
		newTask := &timer.Task{
			Name:      selectedTask.Description,
			WorkTime:  m.app.GetConfig().WorkDuration,
			BreakTime: m.app.GetConfig().BreakDuration,
		}
		if err := timer.SaveTask(newTask); err != nil {
			fmt.Printf("%s保存任务失败：%v%s\n", common.Red, err, common.Reset)
			return
		}
		m.app.SetTask(newTask) // 更新 app 中的任务实例
	} else {
		// 如果有已加载的任务，则更新它
		if err := m.app.UpdateTask(selectedTask.Description, m.app.GetConfig().WorkDuration, m.app.GetConfig().BreakDuration); err != nil {
			fmt.Printf("%s更新任务配置失败：%v%s\n", common.Red, err, common.Reset)
			return
		}
	}

	// 创建新的 Timer 实例并启动
	newTimer := timer.NewTimer(*m.app.GetConfig())
	m.app.SetTimer(newTimer)
	m.app.Run()
}

func (m *Menu) handleCompleteTask() {
	incompleteTasks := m.taskManager.GetIncompleteTasks()
	if len(incompleteTasks) == 0 {
		fmt.Println("\n没有未完成的任务可标记完成。")
		fmt.Print("按 Enter 键继续...")
		m.reader.ReadString('\n') // 等待用户输入, 读取并丢弃用户的一次输入（直到按下回车)
		return
	}

	fmt.Println("\n=========== 未完成的任务 ===========")
	for _, t := range incompleteTasks {
		fmt.Printf("ID: %-4d 描述: %s\n", t.ID, t.Description)
	}
	fmt.Println("====================================")

	fmt.Print("请输入要标记完成的任务ID: ")
	idStr, _ := m.reader.ReadString('\n')
	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		fmt.Println("\n无效的任务ID，请输入数字。")
		return
	}

	err = m.taskManager.CompleteTask(id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("\n任务 %d 已标记为完成！\n", id)
	}
}

func (m *Menu) handleDeleteTask() {
	incompleteTasks := m.taskManager.GetIncompleteTasks()
	if len(incompleteTasks) == 0 {
		fmt.Println("\n没有可删除的任务。")
		fmt.Print("按 Enter 键继续...")
		m.reader.ReadString('\n') // 等待用户输入, 读取并丢弃用户的一次输入（直到按下回车)
		return
	}

	fmt.Println("\n=========== 未完成的任务 ===========")
	for _, t := range incompleteTasks {
		fmt.Printf("ID: %-4d 描述: %s\n", t.ID, t.Description)
	}
	fmt.Println("====================================")

	fmt.Print("请输入要删除的任务ID: ")
	idStr, _ := m.reader.ReadString('\n')
	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		fmt.Println("\n无效的任务ID，请输入数字。")
		return
	}

	err = m.taskManager.DeleteTask(id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("\n任务 %d 已删除！\n", id)
	}
}

func (m *Menu) handleClearAllTasks() {
	fmt.Print("确定要清空所有任务吗？(y/n): ")
	confirm, _ := m.reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" || confirm == "yes" {
		m.taskManager.DeleteAllTasks()
		fmt.Println("\n所有任务已清空！")
	} else {
		fmt.Println("\n已取消清空操作。")
	}
	fmt.Print("按 Enter 键继续...")
	m.reader.ReadString('\n') // 等待用户输入, 读取并丢弃用户的一次输入（直到按下回车)
}

func (m *Menu) modifySettings() {
	for {
		fmt.Println("\n==========================")
		fmt.Println("=== 设置菜单 ===")
		fmt.Println("==========================")
		fmt.Println("1. 修改时间显示格式")
		fmt.Println("2. 查看当前设置")
		fmt.Println("3. 重置为默认设置")
		fmt.Println("4. 返回主菜单")
		fmt.Println("==========================")
		fmt.Print("请选择操作 (1-6): ")

		choice, _ := m.reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			m.modifyTimeDisplayFormat()
		case "2":
			m.showCurrentSettings()
		case "3":
			m.resetToDefaultSettings()
		case "4":
			return
		default:
			fmt.Println("\n无效的选择，请重试。")
		}
	}
}

func (m *Menu) modifyTimeDisplayFormat() {
	fmt.Println("\n时间显示格式选项：")
	fmt.Println("1. normal")
	fmt.Println("2. ansi")

	fmt.Print("请选择时间显示格式 (1-2): ")
	choice, _ := m.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	newSetting := setting.GetAppSetting() // 获取当前设置
	switch choice {
	case "1":
		fmt.Printf("%s时间显示格式已设置为：normal%s\n", common.Green, common.Reset)
		newSetting.TimeDisplayFormat = "normal"
	case "2":
		fmt.Printf("%s时间显示格式已设置为：ansi%s\n", common.Green, common.Reset)
		newSetting.TimeDisplayFormat = "ansi"
	default:
		fmt.Printf("%s无效的选择，保持当前格式。%s\n", common.Yellow, common.Reset)
		return

	}
	// 更新全局设置
	setting.SetAppSetting(newSetting)
	fmt.Print("按 Enter 键继续...")
	m.reader.ReadString('\n')
}

func (m *Menu) showCurrentSettings() {
	fmt.Println("\n============= 当前设置 =============")
	fmt.Printf("时间显示格式：%s\n", setting.GetAppSetting().TimeDisplayFormat)
	fmt.Println("====================================")

	fmt.Print("按 Enter 键继续...")
	m.reader.ReadString('\n')
}

func (m *Menu) resetToDefaultSettings() {
	fmt.Print("确定要重置为默认设置吗？(y/n): ")
	confirm, _ := m.reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" || confirm == "yes" {
		// 重置为默认配置
		if err := setting.ResetToDefaultSettings(); err != nil {
			fmt.Printf("%s重置设置失败：%v%s\n", common.Red, err, common.Reset)
		} else {
			fmt.Printf("%s设置已重置为默认值！%s\n", common.Green, common.Reset)
			fmt.Printf("时间显示格式：%s\n", setting.GetAppSetting().TimeDisplayFormat)
		}
	} else {
		fmt.Printf("%s已取消重置操作。%s\n", common.Yellow, common.Reset)
	}

	fmt.Print("按 Enter 键继续...")
	m.reader.ReadString('\n')
}
