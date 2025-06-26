package gomato

import (
	"gomato/pkg/keymap"
	"gomato/pkg/task"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newItemDelegate(keys *keymap.DelegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string
		if i, ok := m.SelectedItem().(task.Task); ok {
			title = i.Title()
		} else {
			return nil
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.Choose):
				return m.NewStatusMessage("You chose " + title)
			}
		}
		return nil
	}
	help := []key.Binding{keys.Choose, keys.Remove}
	d.ShortHelpFunc = func() []key.Binding {
		return help
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}
	return d
}
