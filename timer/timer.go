package timer

import (
	"fmt"
	"os"
	"time"

	"gomato/common"
	"gomato/config"
)

// TimerDisplay 定义了计时器时间显示的接口
// 你可以实现不同的显示方式（如控制台、GUI等）
type TimerDisplay interface {
	ShowTime(phase string, minutes, seconds int)
	Clear()
}

// ConsoleTimerDisplay 是当前的控制台显示实现
// 默认实现与现有逻辑一致
type ConsoleTimerDisplay struct{}

func (c *ConsoleTimerDisplay) ShowTime(phase string, minutes, seconds int) {
	phaseColor := common.Blue
	if phase == "休息时间" {
		phaseColor = common.Green
	}
	common.ClearLine()
	fmt.Printf("\r%s%s剩余时间: %s%02d:%02d%s",
		phaseColor, phase, common.White, minutes, seconds, common.Reset)
}

func (c *ConsoleTimerDisplay) Clear() {
	common.ClearLine()
}

type Timer struct {
	config    config.TaskConfig
	stats     *Stats
	done      chan struct{}
	pause     chan struct{}
	start     chan struct{}
	AutoStart bool
	isPaused  bool
	task      *Task
	isStopped bool         // 添加标志来跟踪计时器是否已停止
	isRest    bool         // 添加标志来跟踪是否在休息模式
	display   TimerDisplay // 新增：用于显示时间的接口
}

type Stats struct {
	WorkSessions  int
	BreakSessions int
	TotalWorkTime time.Duration
}

func NewTimer(cfg config.TaskConfig, autoStart bool) *Timer {
	// 尝试加载已保存的任务
	task, _ := LoadTask()
	if task != nil {
		// 如果存在保存的任务，使用保存的配置
		cfg.TaskName = task.Name
		cfg.WorkDuration = task.WorkTime
		cfg.BreakDuration = task.BreakTime
	}

	timer := &Timer{
		config:    cfg,
		stats:     &Stats{},
		done:      make(chan struct{}),
		pause:     make(chan struct{}),
		start:     make(chan struct{}),
		AutoStart: autoStart,
		isPaused:  false,
		task:      task,
		isStopped: false,
		isRest:    false,
		display:   &ConsoleTimerDisplay{}, // 默认使用控制台显示
	}

	return timer
}

func (t *Timer) Start() {
	for {
		if !t.AutoStart {
			select {
			case <-t.done:
				return
			case <-t.start:
				// 收到开始信号，继续执行
			}
		}

		select {
		case <-t.done:
			return
		default:
			// 工作时间
			fmt.Printf("\n%s========== 开始专注工作 [%s%s%s] ==========%s\n", common.Bold, common.Red, t.config.TaskName, common.Reset, common.Reset)
			t.stats.WorkSessions++ // 在开始工作时就增加工作会话数
			if !t.timer(t.config.WorkDuration, "工作时间") {
				return // 如果用户在计时器中途停止或暂停，则返回
			}
			fmt.Printf("\n%s工作时间结束！%s\n", common.Blue, common.Reset)
			t.stats.TotalWorkTime += t.config.WorkDuration

			// 询问是否继续休息
			fmt.Printf("\n%s是否开始休息？(y/n)%s ", common.Yellow, common.Reset)
			os.Stdout.Sync() // 确保提示信息被打印

			t.isRest = true // 设置休息模式标志

			// 处理暂停/恢复的逻辑
			select {
			case <-t.done:
				return
			case <-t.pause:
				fmt.Printf("\n%s已暂停，输入 'resume' 继续%s\n", common.Yellow, common.Reset)
				<-t.pause // 等待恢复信号
			default:
				// 如果没有暂停信号，继续下一个循环
			}
		}
	}
}

func (t *Timer) Stop() {
	if !t.isStopped {
		t.isStopped = true
		close(t.done)
	}
}

func (t *Timer) Pause() {
	if !t.isPaused {
		t.pause <- struct{}{}
		t.isPaused = true
	}
}

func (t *Timer) Resume() {
	if t.isPaused {
		t.pause <- struct{}{}
		t.isPaused = false
	}
}

func (t *Timer) TriggerStart() {
	t.start <- struct{}{}
}

func (t *Timer) timer(duration time.Duration, phase string) bool {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	endTime := time.Now().Add(duration)
	if phase == "休息时间" {
	}

	var remaining time.Duration // 新增变量用于记录剩余时间

	for {
		select {
		case <-t.done:
			return false
		case <-t.pause:
			// 暂停时记录剩余时间
			remaining = time.Until(endTime)
			t.display.Clear() // 用接口清除当前行
			fmt.Printf("%s已暂停，输入 'resume' 继续%s\n", common.Yellow, common.Reset)
			<-t.pause         // 等待恢复信号
			t.display.Clear() // 恢复时清除暂停提示行
			// 恢复时用剩余时间重新设置 endTime
			endTime = time.Now().Add(remaining)
		case <-ticker.C:
			remaining = time.Until(endTime)
			if remaining <= 0 {
				t.display.Clear() // 结束时清除倒计时行
				playSound()
				return true
			}
			minutes := int(remaining.Minutes())
			seconds := int(remaining.Seconds()) % 60
			t.display.ShowTime(phase, minutes, seconds)
		}
	}
}

func (t *Timer) GetStats() *Stats {
	return t.stats
}

func playSound() {
	// TODO: 实现声音提醒功能
}

func (t *Timer) UpdateTask(name string, workTime, breakTime time.Duration) error {
	// 如果 task 为 nil，创建一个新的 Task 实例
	if t.task == nil {
		t.task = &Task{
			Name:      name,
			WorkTime:  workTime,
			BreakTime: breakTime,
		}
	} else {
		// 更新任务数据
		t.task.Name = name
		t.task.WorkTime = workTime
		t.task.BreakTime = breakTime
	}

	// 更新配置
	t.config.TaskName = name
	t.config.WorkDuration = workTime
	t.config.BreakDuration = breakTime

	// 保存到JSON文件
	if err := SaveTask(t.task); err != nil {
		return fmt.Errorf("保存任务配置失败：%v", err)
	}

	return nil
}

// StartRest 开始休息时间
func (t *Timer) StartRest() {
	if t.isRest {
		fmt.Printf("\n%s========== 开始休息时间 ==========%s\n", common.Bold, common.Reset)
		t.stats.BreakSessions++ // 增加休息会话数
		if !t.timer(t.config.BreakDuration, "休息时间") {
			return // 如果用户在计时器中途停止或暂停，则返回
		}
		fmt.Printf("\n%s休息时间结束！%s\n", common.Green, common.Reset)
	}
}

// EndRest 结束休息时间
func (t *Timer) EndRest() {
	if t.isRest {
		t.isRest = false
		fmt.Printf("\n%s已跳过休息时间%s\n", common.Yellow, common.Reset)
	}
}
