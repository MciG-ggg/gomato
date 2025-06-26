package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gomato/pkg/common"
)

// Task represents a single task in the task list.
// It implements the list.Item interface.
type Task struct {
	Name   string    `json:"title"`
	Detail string    `json:"description"`
	Timer  TimeModel `json:"timer"`
}

// TimeModel 用于任务计时器（需导出字段以便持久化）
type TimeModel struct {
	TimerDuration  int  `json:"timerDuration"`
	TimerRemaining int  `json:"timerRemaining"`
	TimerIsRunning bool `json:"timerIsRunning"`
}

func (t Task) FilterValue() string { return t.Name }
func (t Task) Title() string       { return t.Name }
func (t Task) Description() string { return t.Detail }

// Manager handles task loading, saving, and manipulation.
type Manager struct {
	Tasks    []Task
	filePath string
}

// NewManager creates a new task manager and loads tasks from the default path.
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(home, ".gomato")
	filePath := filepath.Join(configDir, "tasks.json")

	m := &Manager{
		filePath: filePath,
	}

	if err := m.Load(); err != nil {
		// If file doesn't exist, it's fine, we'll create it on save.
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return m, nil
}

// Load reads tasks from the JSON file.
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.Tasks)
}

// Save writes the current tasks to the JSON file.
func (m *Manager) Save() error {
	// Ensure the directory exists.
	dir := filepath.Dir(m.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(m.Tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0644)
}

// AddItem adds a new task to the list and saves it.
func (m *Manager) AddItem(title, description string) {
	m.Tasks = append(m.Tasks, Task{Name: title, Detail: description})
	m.Save() // Consider handling this error
}

// DeleteItem deletes a task from the list and saves the changes.
func (m *Manager) DeleteItem(index int) {
	if index < 0 || index >= len(m.Tasks) {
		return // Or return an error
	}
	m.Tasks = append(m.Tasks[:index], m.Tasks[index+1:]...)
	m.Save() // Consider handling this error
}

func (t TimeModel) View() string {
	min := t.TimerRemaining / 60
	sec := t.TimerRemaining % 60
	remainStr := fmt.Sprintf("%02d:%02d", min, sec)
	status := "已暂停"
	if t.TimerIsRunning {
		status = "运行中"
	}
	controls := "[空格]开始/暂停  [r]重置  [q]返回"
	return common.TitleStyle.Render("番茄钟计时器") + "\n\n" +
		"剩余时间: " + remainStr + "\n" +
		"状态: " + status + "\n\n" +
		common.StatusMessageStyle(controls)
}
