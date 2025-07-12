package notice

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/gen2brain/beeep"
)

const noticeMaintainTime = 10000

// playSound 播放系统声音，自动适配操作系统
func playSound() {
	switch runtime.GOOS {
	case "linux":
		// 使用 paplay 播放系统声音（PulseAudio）
		exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga").Run()
		// 如果 paplay 不可用，尝试使用 aplay
		if exec.Command("which", "paplay").Run() != nil {
			exec.Command("aplay", "/usr/share/sounds/sound-icons/complete.wav").Run()
		}
	case "darwin":
		// macOS 使用 afplay 播放系统声音
		exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()
	case "windows":
		// Windows 使用 PowerShell 播放系统声音
		exec.Command("powershell", "-c", "[console]::beep(800,200)").Run()
	}
}

// SendNotification 发送桌面通知，自动适配操作系统
func SendNotification(title, message string) {
	// 播放声音
	go playSound()

	// 短暂延迟确保声音和通知同步
	time.Sleep(100 * time.Millisecond)

	switch runtime.GOOS {
	case "linux":
		exec.Command("notify-send", "-t", fmt.Sprintf("%d", noticeMaintainTime), title, message).Run()
	case "windows":
		beeep.Notify(title, message, "")
	case "darwin":
		// macOS 使用 osascript 发送通知
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
		exec.Command("osascript", "-e", script).Run()
	}
}
