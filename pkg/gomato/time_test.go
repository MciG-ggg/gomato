package gomato

import (
	"gomato/pkg/common"
	"gomato/pkg/logging"
	"gomato/pkg/task"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
)

func TestHandleTick_Log(t *testing.T) {
	// 清空日志文件（先删再初始化）
	home, _ := os.UserHomeDir()
	logPath := home + "/.gomato/gomato.log"
	os.Remove(logPath)

	// 初始化日志
	err := logging.Init()
	if err != nil {
		t.Fatalf("日志初始化失败: %v", err)
	}
	defer logging.Close()

	// 构造 App
	taskMgr := &task.Manager{Tasks: []task.Task{}}
	m := &App{
		timeModel: task.TimeModel{
			TimerIsRunning: true,
			TimerRemaining: 2,
			IsWorkSession:  true,
		},
		settingModel: SettingModel{
			Settings: common.Settings{Pomodoro: 25, ShortBreak: 5, LongBreak: 15, Cycle: 4},
		},
		taskManager: taskMgr,
		list:        list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}

	handleTick(m)

	// 检查日志内容
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	if !strings.Contains(string(data), "[Tick] Timer ticked") {
		t.Errorf("日志未包含Tick信息: %s", string(data))
	}
}
