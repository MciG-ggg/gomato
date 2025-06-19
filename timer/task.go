package timer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Task struct {
	Name         string        `json:"name"`
	WorkTime     time.Duration `json:"work_time"`
	BreakTime    time.Duration `json:"break_time"`
	SoundEnabled bool          `json:"sound_enabled"`
}

func SaveTask(task *Task) error {
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
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("转换任务数据失败：%v", err)
	}

	// 写入文件
	taskFile := filepath.Join(dataDir, "task.json")
	if err := os.WriteFile(taskFile, data, 0644); err != nil {
		return fmt.Errorf("写入任务文件失败：%v", err)
	}

	return nil
}

func LoadTask() (*Task, error) {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败：%v", err)
	}

	taskFile := filepath.Join(homeDir, ".gomato", "task.json")

	// 检查文件是否存在
	if _, err := os.Stat(taskFile); os.IsNotExist(err) {
		return nil, nil
	}

	// 读取文件
	data, err := os.ReadFile(taskFile)
	if err != nil {
		return nil, fmt.Errorf("读取任务文件失败：%v", err)
	}

	// 解析JSON
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("解析任务数据失败：%v", err)
	}

	return &task, nil
}

// ModifyWorkTime 修改任务的工作时间
func ModifyWorkTime(task *Task, newWorkTime time.Duration) error {
	task.WorkTime = newWorkTime
	return SaveTask(task)
}
