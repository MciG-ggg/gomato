package keymap

import "github.com/charmbracelet/bubbles/key"

// 任务列表视图的按键映射
// ListKeyMap 用于任务列表视图
// 字段名首字母大写以便外部包访问
type ListKeyMap struct {
	Setting          key.Binding
	ToggleTitleBar   key.Binding
	ToggleStatusBar  key.Binding
	TogglePagination key.Binding
	ToggleHelpMenu   key.Binding
	InsertItem       key.Binding
	ChooseTask       key.Binding
	JoinRoom         key.Binding
	LeaveRoom        key.Binding
	ShowMembers      key.Binding
}

func NewListKeyMap() *ListKeyMap {
	return &ListKeyMap{
		InsertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
		Setting: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "setting"),
		),
		ToggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		ToggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		TogglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		ChooseTask: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose a task"),
		),
		ToggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
		JoinRoom: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "join room"),
		),
		LeaveRoom: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "leave room"),
		),
		ShowMembers: key.NewBinding(
			key.WithKeys("ctrl+m"),
			key.WithHelp("ctrl+m", "show members"),
		),
	}
}

// 列表项委托的按键映射
// DelegateKeyMap 用于列表项的选择和删除操作
type DelegateKeyMap struct {
	Choose key.Binding // 选择项目的按键绑定
	Remove key.Binding // 删除项目的按键绑定
}

func NewDelegateKeyMap() *DelegateKeyMap {
	return &DelegateKeyMap{
		Choose: key.NewBinding(
			key.WithKeys("enter"),           // 使用回车键选择
			key.WithHelp("enter", "choose"), // 帮助文本
		),
		Remove: key.NewBinding(
			key.WithKeys("x", "backspace"), // 使用x或退格键删除
			key.WithHelp("x", "delete"),    // 帮助文本
		),
	}
}

// ShortHelp 实现help.KeyMap接口
func (d DelegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.Choose,
		d.Remove,
	}
}

// FullHelp 实现help.KeyMap接口
func (d DelegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.Choose,
			d.Remove,
		},
	}
}

// 番茄钟视图的按键映射
// TimeViewKeyMap 用于番茄钟视图
type TimeViewKeyMap struct {
	Back       key.Binding
	StartPause key.Binding
	Reset      key.Binding
}

func NewTimeViewKeyMap() *TimeViewKeyMap {
	return &TimeViewKeyMap{
		Back: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "返回任务列表"),
		),
		StartPause: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "开始/暂停"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "重置计时"),
		),
	}
}
