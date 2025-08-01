package p2p

import (
	"encoding/json"
	"fmt"
	"gomato/pkg/logging"
	"time"
)

// ValidateMessage 验证消息的有效性
func ValidateMessage(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	switch msg.Type {
	case MsgStateUpdate:
		if msg.Member == nil {
			return fmt.Errorf("state update message must contain member")
		}
		if msg.Member.ID == "" {
			return fmt.Errorf("member ID cannot be empty")
		}
		if msg.Member.Name == "" {
			return fmt.Errorf("member name cannot be empty")
		}
	case MsgJoin:
		if msg.Member == nil {
			return fmt.Errorf("join message must contain member")
		}
		if msg.Member.ID == "" {
			return fmt.Errorf("member ID cannot be empty")
		}
	case MsgLeave:
		if msg.Member == nil {
			return fmt.Errorf("leave message must contain member")
		}
		if msg.Member.ID == "" {
			return fmt.Errorf("member ID cannot be empty")
		}
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return nil
}

// SerializeMessage 序列化消息为JSON字节数组
func SerializeMessage(msg *Message) ([]byte, error) {
	if err := ValidateMessage(msg); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	return data, nil
}

// DeserializeMessage 从JSON字节数组反序列化消息
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %w", err)
	}

	if err := ValidateMessage(&msg); err != nil {
		return nil, fmt.Errorf("invalid deserialized message: %w", err)
	}

	return &msg, nil
}

// CreateStateUpdateMessage 创建状态更新消息
func CreateStateUpdateMessage(member *Member) *Message {
	return &Message{
		Type:   MsgStateUpdate,
		Member: member,
	}
}

// CreateJoinMessage 创建加入房间消息
func CreateJoinMessage(member *Member, roomKey string) *Message {
	return &Message{
		Type:    MsgJoin,
		Member:  member,
		RoomKey: roomKey,
	}
}

// CreateLeaveMessage 创建离开房间消息
func CreateLeaveMessage(member *Member, roomKey string) *Message {
	return &Message{
		Type:    MsgLeave,
		Member:  member,
		RoomKey: roomKey,
	}
}

// CreateMemberFromNode 从节点信息创建成员
func CreateMemberFromNode(nodeID string, name string, state MemberState, timer TimerInfo, taskName string) *Member {
	return &Member{
		ID:        nodeID,
		Name:      name,
		State:     state,
		Timer:     timer,
		TaskName:  taskName,
		UpdatedAt: time.Now().Unix(),
	}
}

// IsMessageFromSelf 检查消息是否来自自己
func IsMessageFromSelf(msg *Message, selfID string) bool {
	if msg == nil || msg.Member == nil {
		return false
	}
	return msg.Member.ID == selfID
}

// LogMessage 记录消息日志
func LogMessage(msg *Message, from string, action string) {
	logging.Log(fmt.Sprintf("[%s] %s message from %s: Type=%s, Member=%s\n",
		action, msg.Type, from, msg.Member.ID, msg.Member.Name))
}
