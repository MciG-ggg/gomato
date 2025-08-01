package gomato

import (
	"fmt"
	"gomato/pkg/common"
	"gomato/pkg/p2p"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// 房间UI状态
type roomUIState int

const (
	roomHidden roomUIState = iota
	roomVisible
	roomInput
)

// 房间UI组件
type RoomUIModel struct {
	state       roomUIState
	roomInput   textinput.Model
	showMembers bool
	roomManager *p2p.RoomManager
	currentRoom *p2p.Room
	lastUpdate  time.Time
}

// 消息类型
type joinRoomMsg struct {
	roomKey string
}

type leaveRoomMsg struct{}

type toggleMembersMsg struct{}

// 创建房间UI模型
func NewRoomUIModel(roomManager *p2p.RoomManager) RoomUIModel {
	ti := textinput.New()
	ti.Placeholder = "输入房间密钥..."
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 30

	return RoomUIModel{
		state:       roomHidden,
		roomInput:   ti,
		showMembers: false,
		roomManager: roomManager,
	}
}

func (m RoomUIModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m RoomUIModel) Update(msg tea.Msg) (RoomUIModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case joinRoomMsg:
		return m.handleJoinRoom(msg)
	case leaveRoomMsg:
		return m.handleLeaveRoom()
	case toggleMembersMsg:
		return m.handleToggleMembers()
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, cmd
}

// 处理加入房间消息
func (m RoomUIModel) handleJoinRoom(msg joinRoomMsg) (RoomUIModel, tea.Cmd) {
	if msg.roomKey == "" {
		return m, nil
	}

	err := m.roomManager.JoinRoom(msg.roomKey)
	if err == nil {
		m.state = roomVisible
		m.roomInput.SetValue("")
		m.showMembers = false
	}
	return m, nil
}

// 处理离开房间消息
func (m RoomUIModel) handleLeaveRoom() (RoomUIModel, tea.Cmd) {
	m.roomManager.LeaveRoom()
	m.state = roomHidden
	m.showMembers = false
	return m, nil
}

// 处理切换成员列表消息
func (m RoomUIModel) handleToggleMembers() (RoomUIModel, tea.Cmd) {
	if m.state == roomVisible {
		m.showMembers = !m.showMembers
	}
	return m, nil
}

// 处理键盘按键
func (m RoomUIModel) handleKeyPress(msg tea.KeyMsg) (RoomUIModel, tea.Cmd) {
	switch m.state {
	case roomInput:
		return m.handleInputKeyPress(msg)
	case roomVisible:
		return m.handleVisibleKeyPress(msg)
	default:
		return m, nil
	}
}

// 处理输入状态的键盘按键
func (m RoomUIModel) handleInputKeyPress(msg tea.KeyMsg) (RoomUIModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		roomKey := strings.TrimSpace(m.roomInput.Value())
		if roomKey != "" {
			return m, func() tea.Msg {
				return joinRoomMsg{roomKey: roomKey}
			}
		}
	case "esc":
		m.state = roomHidden
		m.roomInput.SetValue("")
		return m, nil
	}

	// 更新输入框
	m.roomInput, cmd = m.roomInput.Update(msg)
	return m, cmd
}

// 处理可见状态的键盘按键
func (m RoomUIModel) handleVisibleKeyPress(msg tea.KeyMsg) (RoomUIModel, tea.Cmd) {
	switch msg.String() {
	case "q", "Q", "esc":
		return m, func() tea.Msg { return leaveRoomMsg{} }
	case "m", "M":
		return m, func() tea.Msg { return toggleMembersMsg{} }
	}
	return m, nil
}

func (m RoomUIModel) View() string {
	switch m.state {
	case roomHidden:
		return ""
	case roomInput:
		return m.renderRoomInput()
	case roomVisible:
		return m.renderRoomStatus()
	default:
		return ""
	}
}

// 渲染房间输入界面
func (m RoomUIModel) renderRoomInput() string {
	return common.AppStyle.Render(fmt.Sprintf(`
┌─────────────────────────────────────┐
│ 加入房间                            │
├─────────────────────────────────────┤
│ %s │
│                                     │
│ 按 Enter 加入房间                   │
│ 按 Esc 取消                         │
└─────────────────────────────────────┘
`, m.roomInput.View()))
}

// 渲染房间状态界面
func (m RoomUIModel) renderRoomStatus() string {
	room := m.roomManager.GetRoom()
	if room == nil {
		return ""
	}

	members := m.roomManager.GetMembers()
	memberCount := len(members)

	// 构建房间状态栏
	statusBar := fmt.Sprintf("房间: %s | 成员: %d", room.Key, memberCount)

	// 如果显示成员列表
	if m.showMembers {
		return m.renderRoomWithMembers(statusBar, members)
	}

	// 只显示状态栏
	return common.AppStyle.Render(fmt.Sprintf(`
┌─────────────────────────────────────┐
│ %s                                  │
│                                     │
│ 按 m 显示成员列表                   │
│ 按 q 或 Esc 离开房间                │
└─────────────────────────────────────┘
`, statusBar))
}

// 渲染带成员列表的房间界面
func (m RoomUIModel) renderRoomWithMembers(statusBar string, members map[string]*p2p.Member) string {
	var memberLines []string

	// 将成员转换为切片以便排序
	var memberSlice []*p2p.Member
	for _, member := range members {
		memberSlice = append(memberSlice, member)
	}

	// 按名称排序
	sort.Slice(memberSlice, func(i, j int) bool {
		return memberSlice[i].Name < memberSlice[j].Name
	})

	// 构建成员列表
	for _, member := range memberSlice {
		stateText := m.getStateText(member.State)
		timerText := m.getTimerText(member.Timer)
		memberLine := fmt.Sprintf("  %s [%s] %s", member.Name, stateText, timerText)
		memberLines = append(memberLines, memberLine)
	}

	// 如果没有成员，显示提示
	if len(memberLines) == 0 {
		memberLines = append(memberLines, "  暂无其他成员")
	}

	membersText := strings.Join(memberLines, "\n")

	return common.AppStyle.Render(fmt.Sprintf(`
┌─────────────────────────────────────┐
│ %s │
├─────────────────────────────────────┤
│ 房间成员:                           │
%s
│                                     │
│ 按 m 隐藏成员列表                   │
│ 按 q 或 Esc 离开房间                │
└─────────────────────────────────────┘
`, statusBar, membersText))
}

// 获取状态文本
func (m RoomUIModel) getStateText(state p2p.MemberState) string {
	switch state {
	case p2p.StateWork:
		return "工作中"
	case p2p.StateRest:
		return "休息中"
	case p2p.StateIdle:
		return "空闲"
	default:
		return "未知"
	}
}

// 获取计时器文本
func (m RoomUIModel) getTimerText(timer p2p.TimerInfo) string {
	if !timer.IsRunning {
		return ""
	}

	minutes := timer.Remaining / 60
	seconds := timer.Remaining % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// 检查是否在房间中
func (m RoomUIModel) IsInRoom() bool {
	return m.state == roomVisible
}

// 显示房间输入
func (m RoomUIModel) ShowInput() RoomUIModel {
	m.state = roomInput
	m.roomInput.Focus()
	return m
}

// 隐藏房间UI
func (m RoomUIModel) Hide() RoomUIModel {
	m.state = roomHidden
	m.showMembers = false
	return m
}

// 将当前应用状态转换为Member对象
func (m RoomUIModel) CreateMemberFromApp(app *App) *p2p.Member {
	var state p2p.MemberState
	if app.timeModel.TimerIsRunning {
		if app.timeModel.IsWorkSession {
			state = p2p.StateWork
		} else {
			state = p2p.StateRest
		}
	} else {
		state = p2p.StateIdle
	}

	timerInfo := p2p.TimerInfo{
		Duration:  app.timeModel.TimerDuration,
		Remaining: app.timeModel.TimerRemaining,
		IsRunning: app.timeModel.TimerIsRunning,
		IsWork:    app.timeModel.IsWorkSession,
	}

	return &p2p.Member{
		ID:        m.roomManager.GetNode().GetHostID(),
		Name:      "用户", // TODO: 从配置获取用户名
		State:     state,
		Timer:     timerInfo,
		UpdatedAt: time.Now().Unix(),
	}
}

// 广播当前状态到房间
func (m RoomUIModel) BroadcastState(app *App) error {
	if !m.IsInRoom() {
		return nil
	}

	member := m.CreateMemberFromApp(app)
	return m.roomManager.BroadcastState(member)
}
