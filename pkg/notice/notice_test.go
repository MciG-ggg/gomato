package notice

import "testing"

func TestSendNotification(t *testing.T) {
	// 这里只测试不会 panic，实际效果需人工观察桌面通知和声音
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SendNotification panic: %v", r)
		}
	}()
	SendNotification("测试通知", "这是一条测试消息")
}

func TestPlaySound(t *testing.T) {
	// 测试声音播放功能不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("playSound panic: %v", r)
		}
	}()

	// 通过反射调用私有函数进行测试
	// 注意：这里我们通过 SendNotification 间接测试声音功能
	SendNotification("声音测试", "请确认是否听到声音")
}
