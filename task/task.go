package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Completed   bool      `json:"completed"`
}

type TaskManager struct {
	tasks []*Task
}

func NewTaskManager() *TaskManager {
	tm := &TaskManager{
		tasks: make([]*Task, 0),
	}
	tm.Load() // 加载保存的任务
	return tm
}

func (tm *TaskManager) AddTask(description string) *Task {
	task := &Task{
		ID:          len(tm.tasks) + 1,
		Description: description,
		StartTime:   time.Now(),
		Completed:   false,
	}
	tm.tasks = append(tm.tasks, task)
	tm.Save() // 保存任务
	return task
}

func (tm *TaskManager) CompleteTask(id int) error {
	for _, task := range tm.tasks {
		if task.ID == id {
			task.Completed = true
			task.EndTime = time.Now()
			tm.Save() // 保存任务
			return nil
		}
	}
	return fmt.Errorf("任务 %d 不存在", id)
}

func (tm *TaskManager) ListTasks() []*Task {
	return tm.tasks
}

func (tm *TaskManager) GetTaskStats() (int, int) {
	total := len(tm.tasks)
	completed := 0
	for _, task := range tm.tasks {
		if task.Completed {
			completed++
		}
	}
	return total, completed
}

// GetRecentTasks 返回最近添加的n个任务
func (tm *TaskManager) GetRecentTasks(n int) []*Task {
	if n <= 0 || n > len(tm.tasks) {
		n = len(tm.tasks)
	}

	// 创建一个新的切片来存储最近的任务
	recentTasks := make([]*Task, n)
	copy(recentTasks, tm.tasks[len(tm.tasks)-n:])
	return recentTasks
}

// GetIncompleteTasks 返回所有未完成的任务
func (tm *TaskManager) GetIncompleteTasks() []*Task {
	var incompleteTasks []*Task
	for _, task := range tm.tasks {
		if !task.Completed {
			incompleteTasks = append(incompleteTasks, task)
		}
	}
	return incompleteTasks
}

// GetTaskByID 根据ID获取任务
func (tm *TaskManager) GetTaskByID(id int) (*Task, error) {
	for _, task := range tm.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return nil, fmt.Errorf("任务 %d 不存在", id)
}

func (tm *TaskManager) Save() error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败：%v", err)
	}

	// 在用户主目录下创建 .gomato 目录
	dataDir := filepath.Join(homeDir, ".gomato")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败：%v", err)
	}

	// 将任务数据转换为JSON
	data, err := json.MarshalIndent(tm.tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("转换任务数据失败：%v", err)
	}

	// 写入文件
	taskFile := filepath.Join(dataDir, "tasks.json")
	if err := os.WriteFile(taskFile, data, 0644); err != nil {
		return fmt.Errorf("写入任务文件失败：%v", err)
	}

	return nil
}

func (tm *TaskManager) Load() error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败：%v", err)
	}

	taskFile := filepath.Join(homeDir, ".gomato", "tasks.json")

	// 检查文件是否存在
	if _, err := os.Stat(taskFile); os.IsNotExist(err) {
		return nil // 如果文件不存在，返回空任务列表
	}

	// 读取文件
	data, err := os.ReadFile(taskFile)
	if err != nil {
		return fmt.Errorf("读取任务文件失败：%v", err)
	}

	// 解析JSON
	if err := json.Unmarshal(data, &tm.tasks); err != nil {
		return fmt.Errorf("解析任务数据失败：%v", err)
	}

	return nil
}

// DeleteTask 删除指定ID的任务
func (tm *TaskManager) DeleteTask(id int) error {
	for i, task := range tm.tasks {
		if task.ID == id {
			// 删除任务
			tm.tasks = append(tm.tasks[:i], tm.tasks[i+1:]...)

			// 重新编号剩余任务
			for j := i; j < len(tm.tasks); j++ {
				tm.tasks[j].ID = j + 1
			}

			// 保存更改
			if err := tm.Save(); err != nil {
				return fmt.Errorf("保存任务失败：%v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("任务 %d 不存在", id)
}

// DeleteAllTasks 删除所有任务
func (tm *TaskManager) DeleteAllTasks() {
	tm.tasks = make([]*Task, 0)

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// 删除任务文件
	taskFile := filepath.Join(homeDir, ".gomato", "tasks.json")
	os.Remove(taskFile)
}
