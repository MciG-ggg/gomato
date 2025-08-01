package p2p

// 房间成员状态
type MemberState string

const (
	StateIdle MemberState = "idle"
	StateWork MemberState = "work"
	StateRest MemberState = "rest"
)

// 成员信息
type Member struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	State     MemberState `json:"state"`
	Timer     TimerInfo   `json:"timer"`
	TaskName  string      `json:"taskName"` // 当前任务名称
	UpdatedAt int64       `json:"updatedAt"`
}

// 计时器信息
type TimerInfo struct {
	Duration  int  `json:"duration"`
	Remaining int  `json:"remaining"`
	IsRunning bool `json:"isRunning"`
	IsWork    bool `json:"isWork"`
}

// 房间信息
type Room struct {
	Key     string             `json:"key"`
	Members map[string]*Member `json:"members"`
	Topic   string             `json:"topic"`
}

// 消息类型
type MessageType string

const (
	MsgStateUpdate MessageType = "state_update"
	MsgJoin        MessageType = "join"
	MsgLeave       MessageType = "leave"
)

// 网络消息
type Message struct {
	Type    MessageType `json:"type"`
	Member  *Member     `json:"member,omitempty"`
	RoomKey string      `json:"roomKey,omitempty"`
}
