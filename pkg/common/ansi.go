package common

import (
	"strings"
)

// ASCII 艺术数字和符号
const (
	// 数字 0
	ANSI_0 = `  ___  
 / _ \ 
| | | |
| |_| |
 \___/ 
       `

	// 数字 1
	ANSI_1 = ` _ 
/ |
| |
| |
|_|
   `

	// 数字 2
	ANSI_2 = ` ____  
|___ \ 
  __) |
 / __/ 
|_____|
       `

	// 数字 3
	ANSI_3 = ` _____ 
|___ / 
  |_ \ 
 ___) |
|____/ 
       `

	// 数字 4
	ANSI_4 = ` _  _   
| || |  
| || |_ 
|__   _|
   |_|  
        `

	// 数字 5
	ANSI_5 = ` ____  
| ___| 
|___ \ 
 ___) |
|____/ 
       `

	// 数字 6
	ANSI_6 = `  __   
 / /_  
| '_ \ 
| (_) |
 \___/ 
       `

	// 数字 7
	ANSI_7 = ` _____ 
|___  |
   / / 
  / /  
 /_/   
       `

	// 数字 8
	ANSI_8 = `  ___  
 ( _ ) 
 / _ \ 
| (_) |
 \___/ 
       `

	// 数字 9
	ANSI_9 = `  ___  
 / _ \ 
| (_) |
 \__, |
   /_/ 
       `

	// 冒号 :
	ANSI_COLON = `   
 _ 
(_)
 _ 
(_)
   `
)

// TimeToAnsiArt 将时间格式 (MM:SS) 转换为 ANSI 艺术显示
func TimeToAnsiArt(timeStr string) string {
	const linesPerDigit = 8
	var digits []string
	for _, char := range timeStr {
		if char >= '0' && char <= '9' {
			switch char {
			case '0':
				digits = append(digits, ANSI_0)
			case '1':
				digits = append(digits, ANSI_1)
			case '2':
				digits = append(digits, ANSI_2)
			case '3':
				digits = append(digits, ANSI_3)
			case '4':
				digits = append(digits, ANSI_4)
			case '5':
				digits = append(digits, ANSI_5)
			case '6':
				digits = append(digits, ANSI_6)
			case '7':
				digits = append(digits, ANSI_7)
			case '8':
				digits = append(digits, ANSI_8)
			case '9':
				digits = append(digits, ANSI_9)
			}
		} else if char == ':' {
			digits = append(digits, ANSI_COLON)
		}
	}

	if len(digits) == 0 {
		return timeStr
	}

	// 每个数字分割成8行，不足补空行
	digitLines := make([][]string, len(digits))
	for i, digit := range digits {
		lines := strings.Split(digit, "\n")
		for len(lines) < linesPerDigit {
			lines = append(lines, "          ") // 10个空格
		}
		digitLines[i] = lines
	}

	var result strings.Builder
	for i := 0; i < linesPerDigit; i++ {
		if i > 0 {
			result.WriteString("\n")
		}
		for j, lines := range digitLines {
			result.WriteString(lines[i])
			if j < len(digitLines)-1 {
				result.WriteString("  ") // 数字之间的间距
			}
		}
	}
	return result.String()
}
