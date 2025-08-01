package gomato

import (
	"fmt"
	"gomato/pkg/common"
	"gomato/pkg/keymap"
	"gomato/pkg/p2p"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// MemberItem 实现 list.Item 接口，用于在列表中显示成员信息
type MemberItem struct {
	Member    *p2p.Member
	JoinOrder int  // 加入顺序
	IsSelf    bool // 是否是当前用户自己
}

func (i MemberItem) Title() string {
	stateText := getStateText(i.Member.State)
	timerText := getTimerText(i.Member.Timer)

	// 如果有任务名，显示任务名
	if i.Member.TaskName != "" {
		return fmt.Sprintf("%s [%s] %s - %s", i.Member.Name, stateText, timerText, i.Member.TaskName)
	}

	return fmt.Sprintf("%s [%s] %s", i.Member.Name, stateText, timerText)
}

func (i MemberItem) Description() string {
	return fmt.Sprintf("ID: %s", i.Member.ID)
}

func (i MemberItem) FilterValue() string {
	return i.Member.Name
}

// 房间UI状态
type roomUIState int

const (
	roomHidden roomUIState = iota
	roomVisible
	roomInput
)

// 房间UI组件
type RoomUIModel struct {
	state         roomUIState
	roomInput     textinput.Model
	roomManager   *p2p.RoomManager
	memberList    list.Model
	joinOrderMap  map[string]int     // 记录成员加入顺序
	nextJoinOrder int                // 下一个加入顺序号
	roomKeys      *keymap.RoomKeyMap // 房间UI快捷键
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

	// 创建房间快捷键映射
	roomKeys := keymap.NewRoomKeyMap()

	// 创建空的成员列表
	memberList := list.New([]list.Item{}, newMemberDelegate(), 0, 0)
	memberList.Title = "房间成员"
	memberList.Styles.Title = common.TitleStyle
	memberList.SetShowHelp(true) // 显示帮助菜单

	// 设置快捷键提示，类似tasklist的方式
	memberList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			roomKeys.LeaveRoom,
		}
	}

	return RoomUIModel{
		state:         roomHidden,
		roomInput:     ti,
		roomManager:   roomManager,
		memberList:    memberList,
		joinOrderMap:  make(map[string]int),
		nextJoinOrder: 1,
		roomKeys:      roomKeys,
	}
}

// 创建成员委托，用于处理成员列表的显示
func newMemberDelegate() list.ItemDelegate {
	d := list.NewDefaultDelegate()
	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		// 成员列表不需要选择功能，所以这里不做任何处理
		return nil
	}

	// 不显示快捷键帮助
	d.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{}
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{}
	}
	return d
}

// 更新成员列表
func (m *RoomUIModel) updateMemberList() {
	members := m.roomManager.GetMembers()
	var memberItems []list.Item

	// 获取当前用户的ID
	currentUserID := m.roomManager.GetNode().GetHostID()

	// 将成员转换为MemberItem切片
	var memberSlice []*MemberItem
	for _, member := range members {
		joinOrder, exists := m.joinOrderMap[member.ID]
		if !exists {
			// 新成员，分配加入顺序
			joinOrder = m.nextJoinOrder
			m.joinOrderMap[member.ID] = joinOrder
			m.nextJoinOrder++
		}

		// 检查是否是当前用户
		isSelf := member.ID == currentUserID

		memberSlice = append(memberSlice, &MemberItem{
			Member:    member,
			JoinOrder: joinOrder,
			IsSelf:    isSelf,
		})
	}

	// 按加入顺序排序
	sort.Slice(memberSlice, func(i, j int) bool {
		return memberSlice[i].JoinOrder < memberSlice[j].JoinOrder
	})

	// 转换为list.Item切片并找到当前用户的索引
	var currentUserIndex int = -1
	for i, item := range memberSlice {
		memberItems = append(memberItems, item)
		if item.IsSelf {
			currentUserIndex = i
		}
	}

	// 更新列表
	m.memberList.SetItems(memberItems)

	// 如果找到当前用户，将其设置为选中项（显示为紫色）
	if currentUserIndex >= 0 {
		m.memberList.Select(currentUserIndex)
	}
}

// 清理离开的成员
func (m *RoomUIModel) cleanupLeftMembers() {
	currentMembers := m.roomManager.GetMembers()

	// 找出已离开的成员
	var leftMemberIDs []string
	for memberID := range m.joinOrderMap {
		if _, exists := currentMembers[memberID]; !exists {
			leftMemberIDs = append(leftMemberIDs, memberID)
		}
	}

	// 从加入顺序映射中删除离开的成员
	for _, memberID := range leftMemberIDs {
		delete(m.joinOrderMap, memberID)
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
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	// 更新成员列表
	m.updateMemberList()
	m.cleanupLeftMembers()

	m.memberList, cmd = m.memberList.Update(msg)

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
		// 重置加入顺序
		m.joinOrderMap = make(map[string]int)
		m.nextJoinOrder = 1
	}
	return m, nil
}

// 处理离开房间消息
func (m RoomUIModel) handleLeaveRoom() (RoomUIModel, tea.Cmd) {
	m.roomManager.LeaveRoom()
	m.state = roomHidden
	// 清理加入顺序
	m.joinOrderMap = make(map[string]int)
	m.nextJoinOrder = 1
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
	switch {
	case key.Matches(msg, m.roomKeys.LeaveRoom):
		return m, func() tea.Msg { return leaveRoomMsg{} }
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

	return m.renderRoom(statusBar)
}

// 渲染带成员列表的房间界面
func (m RoomUIModel) renderRoom(statusBar string) string {
	// 使用list.Model来渲染成员列表
	m.memberList.Title = statusBar
	memberListView := m.memberList.View()

	return common.AppStyle.Render(fmt.Sprintf("%s\n", memberListView))
}

// 获取状态文本
func getStateText(state p2p.MemberState) string {
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
func getTimerText(timer p2p.TimerInfo) string {
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

	// 获取当前任务名
	taskName := ""
	if app.currentTaskIndex >= 0 && app.currentTaskIndex < len(app.taskManager.Tasks) {
		taskName = app.taskManager.Tasks[app.currentTaskIndex].Title()
	}

	return &p2p.Member{
		ID:        m.roomManager.GetNode().GetHostID(),
		Name:      "用户", // TODO: 从配置获取用户名
		State:     state,
		Timer:     timerInfo,
		TaskName:  taskName,
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
