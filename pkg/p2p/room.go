package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"gomato/pkg/logging"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type RoomManager struct {
	node   *Node
	ps     *pubsub.PubSub
	topic  *pubsub.Topic
	sub    *pubsub.Subscription
	room   *Room
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRoomManager(node *Node) *RoomManager {
	ctx, cancel := context.WithCancel(node.ctx)
	rm := &RoomManager{
		node:   node,
		ctx:    ctx,
		cancel: cancel,
		room:   &Room{Members: make(map[string]*Member)},
	}
	ps, err := pubsub.NewGossipSub(ctx, node.host)
	if err != nil {
		// TODO : do not panic when deploy
		panic("Failed to create PubSub: " + err.Error())
	}
	rm.ps = ps

	return rm
}

func (rm *RoomManager) JoinRoom(roomKey string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 创建或获取房间
	rm.room = &Room{
		Key:     roomKey,
		Members: make(map[string]*Member),
		Topic:   fmt.Sprintf("gomato-room-%s", roomKey),
	}

	// 加入房间主题
	topic, err := rm.ps.Join(rm.room.Topic)
	if err != nil {
		return fmt.Errorf("failed to join room topic: %w", err)
	}
	rm.topic = topic

	// 订阅房间主题
	sub, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to room topic: %w", err)
	}
	rm.sub = sub

	// 启动消息处理协程
	go rm.handleMessages()

	// 等待一小段时间确保订阅生效
	time.Sleep(100 * time.Millisecond)

	// 广播加入消息
	rm.broadcastJoin()

	return nil
}

func (rm *RoomManager) LeaveRoom() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.room == nil {
		return nil
	}

	// 广播离开消息
	rm.broadcastLeave()

	// 取消订阅和关闭主题
	if rm.sub != nil {
		rm.sub.Cancel()
		rm.sub = nil
	}
	if rm.topic != nil {
		rm.topic.Close()
		rm.topic = nil
	}

	// 清理房间数据
	rm.room = nil

	// 不要取消上下文，因为可能还需要重新加入房间
	// rm.cancel() // 取消上下文
	return nil
}

func (rm *RoomManager) handleMessages() {
	for {
		// 检查订阅是否有效
		if rm.sub == nil {
			logging.Log("Subscription is nil, stopping message handler\n")
			return
		}

		msg, err := rm.sub.Next(rm.ctx)
		if err != nil {
			if err == context.Canceled {
				return // 上下文取消，退出处理
			}
			logging.Log(fmt.Sprintf("Error reading message: %v\n", err))
			continue
		}

		// 不要忽略自己的消息，因为我们需要处理自己的加入消息
		// if msg.ReceivedFrom == rm.node.host.ID() {
		// 	continue // 忽略自己的消息
		// }

		var message Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			logging.Log(fmt.Sprintf("Failed to decode message: %v\n", err))
			continue
		}

		logging.Log(fmt.Sprintf("Received message: %+v from %s\n", message, msg.ReceivedFrom))
		rm.handleMessage(&message)
	}
}

func (rm *RoomManager) handleMessage(msg *Message) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 检查房间是否存在
	if rm.room == nil {
		logging.Log("Room is nil, ignoring message\n")
		return
	}

	logging.Log(fmt.Sprintf("Processing message: Type=%s, Member=%+v\n", msg.Type, msg.Member))

	switch msg.Type {
	case MsgStateUpdate:
		if msg.Member != nil {
			// 总是更新成员状态
			rm.room.Members[msg.Member.ID] = msg.Member
		}
	case MsgJoin:
		if msg.Member != nil {
			// 只在成员不存在时添加
			if _, exists := rm.room.Members[msg.Member.ID]; !exists {
				rm.room.Members[msg.Member.ID] = msg.Member
			}
		}
	case MsgLeave:
		if msg.Member != nil {
			delete(rm.room.Members, msg.Member.ID)
		}
	}
}

func (rm *RoomManager) broadcastJoin() {
	member := &Member{
		ID:        rm.node.host.ID().String(),
		Name:      "用户", // TODO: 从配置获取用户名
		State:     StateIdle,
		UpdatedAt: time.Now().Unix(),
	}

	msg := Message{
		Type:   MsgJoin,
		Member: member,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logging.Log(fmt.Sprintf("Failed to marshal join message: %v\n", err))
		return
	}

	err = rm.topic.Publish(rm.ctx, data)
	if err != nil {
		logging.Log(fmt.Sprintf("Failed to publish join message: %v\n", err))
	} else {
		logging.Log(fmt.Sprintf("Broadcasted join message for member: %s\n", member.ID))
	}
}

func (rm *RoomManager) broadcastLeave() {
	member := &Member{
		ID:        rm.node.host.ID().String(),
		Name:      "用户",
		State:     StateIdle,
		UpdatedAt: time.Now().Unix(),
	}

	msg := Message{
		Type:   MsgLeave,
		Member: member,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logging.Log(fmt.Sprintf("Failed to marshal leave message: %v\n", err))
		return
	}

	err = rm.topic.Publish(rm.ctx, data)
	if err != nil {
		logging.Log(fmt.Sprintf("Failed to publish leave message: %v\n", err))
	} else {
		logging.Log(fmt.Sprintf("Broadcasted leave message for member: %s\n", member.ID))
	}
}

func (rm *RoomManager) GetMembers() map[string]*Member {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.room == nil {
		return nil
	}

	members := make(map[string]*Member)
	for id, member := range rm.room.Members {
		members[id] = member
	}
	return members
}

func (rm *RoomManager) BroadcastState(member *Member) error {
	logging.Log(fmt.Sprintf("BroadcastState called with member: %+v\n", member))

	if rm.room == nil || rm.topic == nil {
		logging.Log("BroadcastState failed: room or topic is nil\n")
		return fmt.Errorf("not in a room")
	}

	msg := Message{
		Type:   MsgStateUpdate,
		Member: member,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logging.Log(fmt.Sprintf("BroadcastState failed to marshal: %v\n", err))
		return err
	}

	logging.Log(fmt.Sprintf("Broadcasting state update: %+v\n", msg))

	err = rm.topic.Publish(rm.ctx, data)
	if err != nil {
		logging.Log(fmt.Sprintf("Failed to publish state update: %v\n", err))
		return err
	}

	logging.Log(fmt.Sprintf("Successfully broadcasted state update for member: %s\n", member.ID))
	return nil
}
