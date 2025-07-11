package notice

import "testing"

func TestSendNotification(t *testing.T) {
	// 这里只测试不会 panic，实际效果需人工观察桌面通知
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SendNotification panic: %v", r)
		}
	}()
	SendNotification("测试通知", "这是一条测试消息")
}
