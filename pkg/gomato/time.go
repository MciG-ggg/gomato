package gomato

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
