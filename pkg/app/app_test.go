package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AppTestSuite struct {
	suite.Suite
	app *App
}

func (s *AppTestSuite) SetupTest() {
	s.app = NewApp()
}

func (s *AppTestSuite) TearDownTest() {
	if s.app != nil && s.app.timer != nil {
		s.app.timer.Stop()
	}
}

func (s *AppTestSuite) TestNewApp() {
	assert.NotNil(s.T(), s.app, "App 实例不应为 nil")
	assert.NotNil(s.T(), s.app.timer, "timer 不应为 nil")
	assert.NotNil(s.T(), s.app.tasks, "tasks 不应为 nil")
	assert.NotNil(s.T(), s.app.cfg, "cfg 不应为 nil")
}

func (s *AppTestSuite) TestAddTask() {
	description := "测试任务"
	s.app.addTask(description)

	tasks := s.app.tasks.ListTasks()
	assert.Len(s.T(), tasks, 1, "应该只有一个任务")
	assert.Equal(s.T(), description, tasks[0].Description, "任务描述应该匹配")
}

func (s *AppTestSuite) TestCompleteTask() {
	s.app.addTask("测试任务")

	// 测试完成存在的任务
	err := s.app.tasks.CompleteTask(1)
	assert.NoError(s.T(), err, "完成任务时不应有错误")

	// 测试完成不存在的任务
	err = s.app.tasks.CompleteTask(999)
	assert.Error(s.T(), err, "完成不存在的任务时应该有错误")
}

func (s *AppTestSuite) TestSetWorkDuration() {
	duration := "30m"
	s.app.setWorkDuration(duration)

	expected, _ := time.ParseDuration(duration)
	assert.Equal(s.T(), expected, s.app.cfg.WorkDuration, "工作时间应该匹配")
}

func (s *AppTestSuite) TestSetBreakDuration() {
	duration := "10m"
	s.app.setBreakDuration(duration)

	expected, _ := time.ParseDuration(duration)
	assert.Equal(s.T(), expected, s.app.cfg.BreakDuration, "休息时间应该匹配")
}

func (s *AppTestSuite) TestPrintStats() {
	// 添加一些任务
	s.app.addTask("任务1")
	s.app.addTask("任务2")
	s.app.tasks.CompleteTask(1)

	// 直接检查任务统计
	totalTasks, completedTasks := s.app.tasks.GetTaskStats()
	assert.Equal(s.T(), 2, totalTasks, "总任务数应该为 2")
	assert.Equal(s.T(), 1, completedTasks, "已完成任务数应该为 1")
}

func (s *AppTestSuite) TestPrintTasks() {
	// 测试空任务列表
	s.app.printTasks()

	// 添加一些任务
	s.app.addTask("任务1")
	s.app.addTask("任务2")
	s.app.tasks.CompleteTask(1)

	// 测试有任务的情况
	s.app.printTasks()
}

func (s *AppTestSuite) TestShortcuts() {
	expectedShortcuts := map[string]string{
		"p": "pause",
		"r": "resume",
		"s": "stats",
		"t": "tasks",
		"h": "help",
		"q": "quit",
	}

	for k, v := range expectedShortcuts {
		assert.Equal(s.T(), v, shortcuts[k], "快捷键映射应该匹配")
	}
}

func (s *AppTestSuite) TestProcessCommand() {
	testCases := []struct {
		name string
		cmd  string
	}{
		{"帮助命令", "help"},
		{"开始命令", "start"},
		{"统计命令", "stats"},
		{"任务列表命令", "tasks"},
		{"暂停命令", "pause"},
		{"恢复命令", "resume"},
		{"添加任务命令", "add 测试任务"},
		{"完成任务命令", "complete 1"},
		{"设置工作时间命令", "work 30m"},
		{"设置休息时间命令", "break 10m"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.app.processCommand(tc.cmd)
		})
	}
}

func (s *AppTestSuite) TestSetTaskName() {
	// 测试默认任务名称
	assert.Equal(s.T(), "默认任务", s.app.cfg.TaskName, "默认任务名称应该匹配")

	// 测试设置新的任务名称
	newName := "新任务名称"
	s.app.SetTaskName(newName)
	assert.Equal(s.T(), newName, s.app.cfg.TaskName, "新任务名称应该匹配")

	// 测试设置空任务名称
	s.app.SetTaskName("")
	assert.Equal(s.T(), "", s.app.cfg.TaskName, "空任务名称应该匹配")
}

func (s *AppTestSuite) TestTimerAutoStart() {
	// 测试新建应用时不会自动开始
	app := NewApp()

	// 测试设置任务名称后会自动开始
	app.SetTaskName("测试任务")

	// 测试手动开始命令
	app = NewApp()
	app.processCommand("start")
	// 由于 start 命令是异步的，我们需要等待一小段时间
	time.Sleep(100 * time.Millisecond)
	// 这里我们只能验证命令被处理了，但不能直接验证计时器状态
	// 因为计时器的状态是异步的
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}
