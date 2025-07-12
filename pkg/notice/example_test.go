package notice

import (
	"time"
)

// ExampleSendNotification 演示如何使用带声音的通知功能
func ExampleSendNotification() {
	// 发送一条带声音的通知
	SendNotification("任务完成", "您的番茄钟任务已完成！")

	// 等待一下让用户听到声音
	time.Sleep(2 * time.Second)

	// 发送另一条通知
	SendNotification("休息提醒", "现在是休息时间，请放松一下")
}

// ExampleSendNotification_multiple 演示发送多条通知
func ExampleSendNotification_multiple() {
	// 发送多条通知，每条都会播放声音
	notifications := []struct {
		title   string
		message string
	}{
		{"开始工作", "番茄钟开始，专注工作！"},
		{"工作提醒", "还有5分钟结束当前番茄钟"},
		{"工作完成", "当前番茄钟已完成，请休息"},
		{"休息结束", "休息时间结束，准备开始下一个番茄钟"},
	}

	for _, notification := range notifications {
		SendNotification(notification.title, notification.message)
		time.Sleep(1 * time.Second) // 间隔1秒发送下一条
	}
}
