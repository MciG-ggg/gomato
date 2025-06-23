package common

import "fmt"

// ANSI 颜色代码
const (
	Reset      = "\033[0m"
	Bold       = "\033[1m"
	Red        = "\033[31m"
	Green      = "\033[32m"
	Yellow     = "\033[33m"
	Blue       = "\033[34m"
	Magenta    = "\033[35m"
	Cyan       = "\033[36m"
	White      = "\033[37m"
	RedBold    = "\033[1;31m"
	GreenBold  = "\033[1;32m"
	YellowBold = "\033[1;33m"
	BlueBold   = "\033[1;34m"
)

// ClearLine 清除当前行并移动光标到行首
func ClearLine() {
	fmt.Print("\r\033[K")
}
