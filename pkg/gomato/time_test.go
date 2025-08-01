package gomato

import (
	"fmt"
	"gomato/pkg/common"
	"gomato/pkg/logging"
	"gomato/pkg/task"
	"os"
	"strings"
	"testing"
	"time"

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
		taskList:    list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
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

// TestTimerTickFrequency 测试计时器每秒只触发一次
func TestTimerTickFrequency(t *testing.T) {
	// 清空日志文件
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
			TimerRemaining: 5, // 设置5秒
			IsWorkSession:  true,
		},
		settingModel: SettingModel{
			Settings: common.Settings{Pomodoro: 25, ShortBreak: 5, LongBreak: 15, Cycle: 4},
		},
		taskManager: taskMgr,
		taskList:    list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}

	// 模拟2秒的时间，每秒调用一次handleTick
	initialRemaining := m.timeModel.TimerRemaining

	// 第一次tick
	handleTick(m)
	time.Sleep(100 * time.Millisecond) // 短暂等待

	// 第二次tick
	handleTick(m)
	time.Sleep(100 * time.Millisecond) // 短暂等待

	// 检查日志文件
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 计算tick日志的数量
	tickCount := strings.Count(string(data), "[Tick] Timer ticked")

	// 应该只有2次tick（因为我们调用了2次handleTick）
	if tickCount != 2 {
		t.Errorf("期望2次tick，但实际有%d次tick", tickCount)
	}

	// 检查剩余时间是否正确递减
	expectedRemaining := initialRemaining - 2
	if m.timeModel.TimerRemaining != expectedRemaining {
		t.Errorf("期望剩余时间%d，但实际是%d", expectedRemaining, m.timeModel.TimerRemaining)
	}

	// 检查日志中的时间序列是否正确
	lines := strings.Split(string(data), "\n")
	tickLines := []string{}
	for _, line := range lines {
		if strings.Contains(line, "[Tick] Timer ticked") {
			tickLines = append(tickLines, line)
		}
	}

	// 验证时间序列：应该是 4, 3 (从5开始，每次减1)
	if len(tickLines) >= 2 {
		if !strings.Contains(tickLines[0], "remaining: 4") {
			t.Errorf("第一次tick应该显示remaining: 4，但显示: %s", tickLines[0])
		}
		if !strings.Contains(tickLines[1], "remaining: 3") {
			t.Errorf("第二次tick应该显示remaining: 3，但显示: %s", tickLines[1])
		}
	}
}

// TestBubbleTeaTickSimulation 模拟Bubble Tea事件循环中的重复tick问题
func TestBubbleTeaTickSimulation(t *testing.T) {
	// 清空日志文件
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
			TimerRemaining: 3,
			IsWorkSession:  true,
		},
		settingModel: SettingModel{
			Settings: common.Settings{Pomodoro: 25, ShortBreak: 5, LongBreak: 15, Cycle: 4},
		},
		taskManager: taskMgr,
		taskList:    list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}

	// 模拟Bubble Tea事件循环：连续处理多个tick消息
	// 这可能会触发重复的tick命令
	for i := 0; i < 3; i++ {
		cmd := handleTick(m)
		if cmd != nil {
			// 模拟执行命令（这里只是记录，实际应该由Bubble Tea处理）
			t.Logf("第%d次tick返回了命令", i+1)
		}
		time.Sleep(50 * time.Millisecond) // 模拟快速连续的处理
	}

	// 检查日志文件
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 计算tick日志的数量
	tickCount := strings.Count(string(data), "[Tick] Timer ticked")
	t.Logf("总共记录了%d次tick", tickCount)

	// 检查剩余时间
	if m.timeModel.TimerRemaining < 0 {
		t.Errorf("剩余时间不应该为负数: %d", m.timeModel.TimerRemaining)
	}

	// 检查日志中的时间序列
	lines := strings.Split(string(data), "\n")
	tickLines := []string{}
	for _, line := range lines {
		if strings.Contains(line, "[Tick] Timer ticked") {
			tickLines = append(tickLines, line)
		}
	}

	// 输出所有tick日志用于调试
	for i, line := range tickLines {
		t.Logf("Tick %d: %s", i+1, line)
	}

	// 验证时间序列应该是递减的
	for i := 1; i < len(tickLines); i++ {
		prevLine := tickLines[i-1]
		currLine := tickLines[i]

		// 提取剩余时间数字
		prevRemaining := extractRemainingTime(prevLine)
		currRemaining := extractRemainingTime(currLine)

		if prevRemaining != -1 && currRemaining != -1 && prevRemaining <= currRemaining {
			t.Errorf("时间序列应该递减，但 %d <= %d", prevRemaining, currRemaining)
		}
	}
}

// extractRemainingTime 从日志行中提取剩余时间数字
func extractRemainingTime(line string) int {
	// 简单的字符串解析，查找 "remaining: X" 格式
	parts := strings.Split(line, "remaining: ")
	if len(parts) < 2 {
		return -1
	}

	// 提取数字部分
	numStr := strings.TrimSpace(parts[1])
	var num int
	_, err := fmt.Sscanf(numStr, "%d", &num)
	if err != nil {
		return -1
	}
	return num
}

// TestTimerNoDuplicateTicks 测试修复后的计时器不会重复tick
func TestTimerNoDuplicateTicks(t *testing.T) {
	// 清空日志文件
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
			TimerRemaining: 3,
			IsWorkSession:  true,
		},
		settingModel: SettingModel{
			Settings: common.Settings{Pomodoro: 25, ShortBreak: 5, LongBreak: 15, Cycle: 4},
		},
		taskManager: taskMgr,
		taskList:    list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}

	// 模拟Bubble Tea事件循环：快速连续处理tick消息
	// 这应该不会导致重复的tick
	initialRemaining := m.timeModel.TimerRemaining
	t.Logf("初始剩余时间: %d", initialRemaining)

	for i := 0; i < 3; i++ {
		cmd := handleTick(m)
		if cmd != nil {
			t.Logf("第%d次tick返回了命令", i+1)
		}
		t.Logf("第%d次tick后剩余时间: %d", i+1, m.timeModel.TimerRemaining)
		// 不等待，模拟快速连续的处理
	}

	// 检查日志文件
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 计算tick日志的数量
	tickCount := strings.Count(string(data), "[Tick] Timer ticked")
	t.Logf("总共记录了%d次tick", tickCount)

	// 应该只有3次tick（因为TimerRemaining从3开始，减到0）
	if tickCount != 3 {
		t.Errorf("期望3次tick，但实际有%d次tick", tickCount)
	}

	// 检查剩余时间（应进入短休息阶段）
	expectedRest := int(m.settingModel.Settings.ShortBreak) * 60
	if m.timeModel.TimerRemaining != expectedRest {
		t.Errorf("期望剩余时间为%d（短休息），但实际是%d", expectedRest, m.timeModel.TimerRemaining)
	}
}

func TestHandleTick_WorkToRestSwitch(t *testing.T) {
	home, _ := os.UserHomeDir()
	logPath := home + "/.gomato/gomato.log"
	os.Remove(logPath)
	logging.Init()
	defer logging.Close()

	// 构造 App，剩余1秒，工作阶段
	taskMgr := &task.Manager{Tasks: []task.Task{}}
	m := &App{
		timeModel: task.TimeModel{
			TimerIsRunning: true,
			TimerRemaining: 1,
			IsWorkSession:  true,
		},
		settingModel: SettingModel{
			Settings: common.Settings{Pomodoro: 25, ShortBreak: 5, LongBreak: 15, Cycle: 4},
		},
		taskManager: taskMgr,
		taskList:    list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}
	_ = handleTick(m)
	// 应进入短休息
	if m.timeModel.IsWorkSession {
		t.Errorf("应已切换到休息阶段")
	}
	expectedRest := int(m.settingModel.Settings.ShortBreak) * 60
	if m.timeModel.TimerRemaining != expectedRest {
		t.Errorf("短休息剩余时间应为%d, 实际为%d", expectedRest, m.timeModel.TimerRemaining)
	}
}
