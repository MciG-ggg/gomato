package notice

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/gen2brain/beeep"
)

const noticeMaintainTime = 10000

// SendNotification 发送桌面通知，自动适配操作系统
func SendNotification(title, message string) {
	switch runtime.GOOS {
	case "linux":
		exec.Command("notify-send", "-t", fmt.Sprintf("%d", noticeMaintainTime), title, message).Run()
	case "windows":
		beeep.Notify(title, message, "")
	}
}
