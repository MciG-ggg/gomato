// taskList.go - TUI界面列表组件和委托处理
// 负责定义列表项的按键映射、委托处理函数等
package main

import (
	"gomato/pkg/keymap"
	"gomato/pkg/task"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// newItemDelegate 创建列表项的委托处理函数
// 负责处理列表项的选择、删除等交互操作
func newItemDelegate(keys *keymap.DelegateKeyMap) list.DefaultDelegate {
	// 创建默认委托
	d := list.NewDefaultDelegate()

	// 设置更新函数，处理用户交互
	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string

		// 获取当前选中项目的标题
		if i, ok := m.SelectedItem().(task.Task); ok {
			title = i.Title()
		} else {
			return nil // 如果没有选中项目，直接返回
		}

		// 根据消息类型处理不同的按键操作
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.Choose):
				// 处理选择操作
				return m.NewStatusMessage(statusMessageStyle("You chose " + title))
			}
		}

		return nil // 没有匹配的操作，返回空命令
	}

	// 设置帮助信息
	help := []key.Binding{keys.Choose, keys.Remove}

	// 设置简短帮助函数
	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	// 设置完整帮助函数
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}
