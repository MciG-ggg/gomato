package timer

import (
	"os"
	"testing"
	"time"

	"gomato/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TimerTestSuite struct {
	suite.Suite
	timer *Timer
}

func (s *TimerTestSuite) SetupTest() {
	cfg := config.DefaultConfig
	s.timer = NewTimer(cfg, false)
}

func (s *TimerTestSuite) TearDownTest() {
	if s.timer != nil {
		s.timer.Stop()
	}
}

func (s *TimerTestSuite) TestNewTimer() {
	// 测试手动开始模式
	timer := NewTimer(config.DefaultConfig, false)
	assert.NotNil(s.T(), timer, "Timer 实例不应为 nil")
	assert.False(s.T(), timer.AutoStart, "手动开始模式下 AutoStart 应为 false")

	// 测试自动开始模式
	timer = NewTimer(config.DefaultConfig, true)
	assert.NotNil(s.T(), timer, "Timer 实例不应为 nil")
	assert.True(s.T(), timer.AutoStart, "自动开始模式下 AutoStart 应为 true")
}

func (s *TimerTestSuite) TestWorkSessionCounting() {
	// 创建一个新的计时器
	timer := NewTimer(config.DefaultConfig, false)

	// 启动计时器（在后台运行）
	go timer.Start()

	// 等待一小段时间确保计时器已经启动
	time.Sleep(100 * time.Millisecond)

	// 触发开始
	timer.TriggerStart()

	// 等待一小段时间确保工作会话已经开始
	time.Sleep(100 * time.Millisecond)

	// 检查工作会话数
	stats := timer.GetStats()
	assert.Equal(s.T(), 1, stats.WorkSessions, "工作会话数应该为1")

	// 停止计时器
	timer.Stop()
}

func (s *TimerTestSuite) TestBreakSessionCounting() {
	// 创建一个新的计时器
	timer := NewTimer(config.DefaultConfig, false)

	// 启动计时器（在后台运行）
	go timer.Start()

	// 等待一小段时间确保计时器已经启动
	time.Sleep(100 * time.Millisecond)

	// 触发开始
	timer.TriggerStart()

	// 等待工作时间结束
	time.Sleep(config.DefaultConfig.WorkDuration + 100*time.Millisecond)

	// 检查休息会话数
	stats := timer.GetStats()
	assert.Equal(s.T(), 1, stats.BreakSessions, "休息会话数应该为1")

	// 停止计时器
	timer.Stop()
}

func (s *TimerTestSuite) TestTotalWorkTime() {
	// 创建一个新的计时器
	timer := NewTimer(config.DefaultConfig, false)

	// 启动计时器（在后台运行）
	go timer.Start()

	// 等待一小段时间确保计时器已经启动
	time.Sleep(100 * time.Millisecond)

	// 触发开始
	timer.TriggerStart()

	// 等待工作时间结束
	time.Sleep(config.DefaultConfig.WorkDuration + 100*time.Millisecond)

	// 检查总工作时间
	stats := timer.GetStats()
	assert.Equal(s.T(), config.DefaultConfig.WorkDuration, stats.TotalWorkTime, "总工作时间应该等于配置的工作时间")

	// 停止计时器
	timer.Stop()
}

func (s *TimerTestSuite) TestPauseAndResume() {
	// 创建一个新的计时器
	timer := NewTimer(config.DefaultConfig, false)

	// 启动计时器（在后台运行）
	go timer.Start()

	// 等待一小段时间确保计时器已经启动
	time.Sleep(100 * time.Millisecond)

	// 触发开始
	timer.TriggerStart()

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)

	// 暂停计时器
	timer.Pause()
	assert.True(s.T(), timer.isPaused, "计时器应该处于暂停状态")

	// 恢复计时器
	timer.Resume()
	assert.False(s.T(), timer.isPaused, "计时器应该不再处于暂停状态")

	// 停止计时器
	timer.Stop()
}

func (s *TimerTestSuite) TestBreakChoice() {
	// 创建一个自定义配置，使用更短的测试时间
	testConfig := config.Config{
		WorkDuration:  2 * time.Second,
		BreakDuration: 1 * time.Second,
		SoundEnabled:  false,
		TaskName:      "测试任务",
	}

	// 创建一个新的计时器
	timer := NewTimer(testConfig, false)

	// 启动计时器（在后台运行）
	go timer.Start()

	// 等待一小段时间确保计时器已经启动
	time.Sleep(100 * time.Millisecond)

	// 触发开始
	timer.TriggerStart()

	// 等待工作时间结束
	time.Sleep(testConfig.WorkDuration + 100*time.Millisecond)

	// 检查初始休息会话数
	initialStats := timer.GetStats()
	initialBreakSessions := initialStats.BreakSessions

	// 模拟用户选择休息
	// 注意：这里我们不能直接测试用户输入，但我们可以验证休息会话数是否正确增加
	time.Sleep(testConfig.BreakDuration + 100*time.Millisecond)

	// 检查休息会话数是否增加
	finalStats := timer.GetStats()
	assert.Equal(s.T(), initialBreakSessions+1, finalStats.BreakSessions, "休息会话数应该增加1")

	// 停止计时器
	timer.Stop()
}

func (s *TimerTestSuite) TestWorkAndRestCycles() {
	// 为测试配置短时长，确保至少触发一次秒级计时器
	testConfig := config.Config{
		WorkDuration:  2 * time.Second, // 改为2秒，确保至少有一个完整的tick
		BreakDuration: 2 * time.Second, // 改为2秒，确保至少有一个完整的tick
		SoundEnabled:  false,
		TaskName:      "循环测试任务",
	}
	timer := NewTimer(testConfig, false)

	// 显式设置计时器的配置，以确保使用测试值，覆盖任何加载的任务配置。
	timer.config.WorkDuration = testConfig.WorkDuration
	timer.config.BreakDuration = testConfig.BreakDuration
	timer.config.SoundEnabled = testConfig.SoundEnabled
	timer.config.TaskName = testConfig.TaskName

	// 模拟 os.Stdin 以控制用户输入
	oldStdin := os.Stdin

	// 创建一个管道来模拟 stdin
	r, w, err := os.Pipe()
	s.Require().NoError(err)

	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin // 确保测试后恢复 stdin
		w.Close()
		r.Close()
	}()

	// 在 Goroutine 中写入模拟输入
	go func() {
		defer w.Close()
		// 分三次写入输入，每次写入之间添加更长的延迟
		for i := 0; i < 3; i++ {
			_, err := w.WriteString("y\n")
			s.Require().NoError(err)
			time.Sleep(1000 * time.Millisecond) // 增加输入间隔到1000毫秒
		}
		// 在第三个周期结束后立即写入 'n' 来停止计时器
		time.Sleep(testConfig.WorkDuration + testConfig.BreakDuration + 500*time.Millisecond)
		_, err := w.WriteString("n\n")
		s.Require().NoError(err)
	}()

	// 在 Goroutine 中运行计时器
	go timer.Start()

	// 给计时器一些时间来启动
	time.Sleep(500 * time.Millisecond)

	// 触发计时器开始（因为 AutoStart 为 false）
	timer.TriggerStart()

	// 计算 3 个完整工作-休息周期的近似时间
	expectedTotalCycleDuration := 3 * (testConfig.WorkDuration + testConfig.BreakDuration)
	sleepDuration := expectedTotalCycleDuration + 3000*time.Millisecond // 增加缓冲时间到3秒

	// 等待周期完成
	time.Sleep(sleepDuration)

	// 显式停止计时器
	timer.Stop()
	// 给 Goroutine 一些时间来完成处理停止信号
	time.Sleep(500 * time.Millisecond) // 减少停止后的等待时间到500毫秒

	// 获取最终统计数据
	stats := timer.GetStats()

	// 断言 4 个完成的周期
	s.Equal(4, stats.WorkSessions, "工作会话数应该为4")
	s.Equal(3, stats.BreakSessions, "休息会话数应该为3")

	// 验证总工作时间。由于可能存在微小差异，使用 InDelta 进行 time.Duration 比较。
	// 总工作时间应该大约是工作时长的 4 倍。
	s.InDelta(4*testConfig.WorkDuration.Seconds(), stats.TotalWorkTime.Seconds(), 0.1, "总工作时间应该约等于配置的4倍工作时间")
}

func TestTimerSuite(t *testing.T) {
	suite.Run(t, new(TimerTestSuite))
}
