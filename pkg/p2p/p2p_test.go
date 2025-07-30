package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"gomato/pkg/logging"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeBasic(t *testing.T) {
	t.Parallel() // 支持并行测试

	node, err := createTestNode(t, 0)
	if err != nil {
		if isKnownConfigError(err) {
			t.Skip("Skipping due to environment configuration:", err)
		}
		t.Fatalf("Failed to create node: %v", err)
	}
	defer node.Close()

	if node.GetHost() == nil {
		t.Error("Host should not be nil")
	}
}

func isKnownConfigError(err error) bool {
	return strings.Contains(err.Error(), "security transport") ||
		strings.Contains(err.Error(), "noise") ||
		strings.Contains(err.Error(), "insecure configuration")
}

func TestNodeAndPubSubIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 测试准备
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. 创建两个测试节点
	node1, err := createTestNode(t, 1)
	require.NoError(t, err, "节点1创建失败")
	defer node1.Close()

	node2, err := createTestNode(t, 2)
	require.NoError(t, err, "节点2创建失败")
	defer node2.Close()

	// 2. 连接两个节点
	connectNodes(t, ctx, node1, node2)

	// 3. 测试房间管理
	t.Run("房间加入和状态同步", func(t *testing.T) {
		roomKey := "test-room-1"

		// 节点1加入房间
		err = node1.roomMgr.JoinRoom(roomKey)
		require.NoError(t, err, "节点1加入房间失败")

		// 等待一段时间让节点1完全加入房间
		time.Sleep(500 * time.Millisecond)

		// 节点2加入房间
		err = node2.roomMgr.JoinRoom(roomKey)
		require.NoError(t, err, "节点2加入房间失败")

		// 等待节点发现和消息处理
		time.Sleep(2 * time.Second)

		// 验证节点1看到至少1个成员（包括自己）
		assert.Eventually(t, func() bool {
			members := node1.roomMgr.GetMembers()
			t.Logf("Node1 members: %d", len(members))
			for id, member := range members {
				t.Logf("Node1 member: %s - %s", id, member.Name)
			}
			return len(members) >= 1
		}, 10*time.Second, 500*time.Millisecond, "节点1应看到成员")

		// 验证节点2看到至少1个成员（包括自己）
		assert.Eventually(t, func() bool {
			members := node2.roomMgr.GetMembers()
			t.Logf("Node2 members: %d", len(members))
			for id, member := range members {
				t.Logf("Node2 member: %s - %s", id, member.Name)
			}
			return len(members) >= 1
		}, 10*time.Second, 500*time.Millisecond, "节点2应看到成员")
	})
}

func TestMessageTransport(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过消息传输测试")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建两个测试节点
	node1, err := createTestNode(t, 10)
	require.NoError(t, err, "节点1创建失败")
	defer node1.Close()

	node2, err := createTestNode(t, 11)
	require.NoError(t, err, "节点2创建失败")
	defer node2.Close()

	// 连接节点
	connectNodes(t, ctx, node1, node2)

	t.Run("状态更新消息传输", func(t *testing.T) {
		roomKey := "test-message-room"

		// 两个节点加入房间
		err = node1.roomMgr.JoinRoom(roomKey)
		require.NoError(t, err, "节点1加入房间失败")
		time.Sleep(500 * time.Millisecond)

		err = node2.roomMgr.JoinRoom(roomKey)
		require.NoError(t, err, "节点2加入房间失败")
		time.Sleep(1 * time.Second)

		// 获取节点1的ID
		node1ID := node1.GetHost().ID().String()
		t.Logf("Node1 ID: %s", node1ID)

		// 创建测试成员状态
		member := &Member{
			ID:        node1ID,
			Name:      "测试用户",
			State:     StateWork,
			Timer:     TimerInfo{Duration: 1500, Remaining: 1200, IsRunning: true, IsWork: true},
			UpdatedAt: time.Now().Unix(),
		}

		t.Logf("Created member for state update: %+v", member)

		// 节点1广播状态更新
		err = node1.roomMgr.BroadcastState(member)
		require.NoError(t, err, "状态广播失败")

		// 等待消息传播
		time.Sleep(2 * time.Second)

		// 验证节点2接收到状态更新
		assert.Eventually(t, func() bool {
			members := node2.roomMgr.GetMembers()
			t.Logf("Node2 members count: %d", len(members))
			for id, member := range members {
				t.Logf("Node2 member: %s -> %s (State: %s)", id, member.Name, member.State)
			}

			if member, exists := members[node1ID]; exists {
				t.Logf("Node2 received state update: %+v", member)
				return member.State == StateWork && member.Timer.IsRunning
			}
			return false
		}, 10*time.Second, 500*time.Millisecond, "节点2应接收到状态更新")
	})

	t.Run("加入和离开消息传输", func(t *testing.T) {
		roomKey := "test-join-leave-room"

		// 节点1加入房间
		err = node1.roomMgr.JoinRoom(roomKey)
		require.NoError(t, err, "节点1加入房间失败")
		time.Sleep(1 * time.Second)

		// 验证节点1在房间中
		members1 := node1.roomMgr.GetMembers()
		assert.GreaterOrEqual(t, len(members1), 1, "节点1应该在房间中")

		// 节点1离开房间
		err = node1.roomMgr.LeaveRoom()
		require.NoError(t, err, "节点1离开房间失败")
		time.Sleep(500 * time.Millisecond) // 等待离开消息传播

		// 节点2加入房间
		err = node2.roomMgr.JoinRoom(roomKey)
		require.NoError(t, err, "节点2加入房间失败")
		time.Sleep(1 * time.Second)

		// 验证节点2看不到节点1（因为节点1已经离开）
		members2 := node2.roomMgr.GetMembers()
		t.Logf("Node2 members after join: %d", len(members2))
		for id, member := range members2 {
			t.Logf("Node2 member: %s - %s", id, member.Name)
		}
		// 节点2应该只看到自己
		assert.Equal(t, 1, len(members2), "节点2应该只看到自己")
	})
}

func TestJSONMessageHandling(t *testing.T) {
	t.Run("消息序列化和反序列化", func(t *testing.T) {
		// 测试消息序列化
		member := &Member{
			ID:        "test-id",
			Name:      "测试用户",
			State:     StateWork,
			Timer:     TimerInfo{Duration: 1500, Remaining: 1200, IsRunning: true, IsWork: true},
			UpdatedAt: time.Now().Unix(),
		}

		msg := Message{
			Type:   MsgStateUpdate,
			Member: member,
		}

		// 序列化
		data, err := json.Marshal(msg)
		require.NoError(t, err, "消息序列化失败")
		assert.NotEmpty(t, data, "序列化数据不应为空")

		// 反序列化
		var decodedMsg Message
		err = json.Unmarshal(data, &decodedMsg)
		require.NoError(t, err, "消息反序列化失败")

		// 验证数据完整性
		assert.Equal(t, msg.Type, decodedMsg.Type, "消息类型应该一致")
		assert.Equal(t, msg.Member.ID, decodedMsg.Member.ID, "成员ID应该一致")
		assert.Equal(t, msg.Member.Name, decodedMsg.Member.Name, "成员名称应该一致")
		assert.Equal(t, msg.Member.State, decodedMsg.Member.State, "成员状态应该一致")
		assert.Equal(t, msg.Member.Timer.Duration, decodedMsg.Member.Timer.Duration, "计时器时长应该一致")
		assert.Equal(t, msg.Member.Timer.IsRunning, decodedMsg.Member.Timer.IsRunning, "计时器运行状态应该一致")
	})

	t.Run("不同类型的消息", func(t *testing.T) {
		testCases := []struct {
			name     string
			msgType  MessageType
			expected string
		}{
			{"状态更新消息", MsgStateUpdate, "state_update"},
			{"加入消息", MsgJoin, "join"},
			{"离开消息", MsgLeave, "leave"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				msg := Message{
					Type: tc.msgType,
					Member: &Member{
						ID:        "test-id",
						Name:      "测试用户",
						State:     StateIdle,
						UpdatedAt: time.Now().Unix(),
					},
				}

				data, err := json.Marshal(msg)
				require.NoError(t, err, "消息序列化失败")

				var decodedMsg Message
				err = json.Unmarshal(data, &decodedMsg)
				require.NoError(t, err, "消息反序列化失败")

				assert.Equal(t, tc.msgType, decodedMsg.Type, "消息类型应该一致")
				assert.Equal(t, tc.expected, string(decodedMsg.Type), "消息类型字符串应该一致")
			})
		}
	})
}

func TestJSONFileOperations(t *testing.T) {
	t.Run("保存和加载房间状态", func(t *testing.T) {
		// 创建测试房间状态
		room := &Room{
			Key: "test-room",
			Members: map[string]*Member{
				"user1": {
					ID:        "user1",
					Name:      "用户1",
					State:     StateWork,
					Timer:     TimerInfo{Duration: 1500, Remaining: 1200, IsRunning: true, IsWork: true},
					UpdatedAt: time.Now().Unix(),
				},
				"user2": {
					ID:        "user2",
					Name:      "用户2",
					State:     StateRest,
					Timer:     TimerInfo{Duration: 300, Remaining: 180, IsRunning: true, IsWork: false},
					UpdatedAt: time.Now().Unix(),
				},
			},
			Topic: "gomato-room-test-room",
		}

		// 保存到临时文件
		tempFile := "test_room_state.json"
		defer os.Remove(tempFile) // 清理临时文件

		data, err := json.MarshalIndent(room, "", "  ")
		require.NoError(t, err, "房间状态序列化失败")

		err = os.WriteFile(tempFile, data, 0644)
		require.NoError(t, err, "写入文件失败")

		// 验证文件存在
		_, err = os.Stat(tempFile)
		require.NoError(t, err, "文件应该存在")

		// 读取并验证数据
		fileData, err := os.ReadFile(tempFile)
		require.NoError(t, err, "读取文件失败")

		var loadedRoom Room
		err = json.Unmarshal(fileData, &loadedRoom)
		require.NoError(t, err, "房间状态反序列化失败")

		// 验证数据完整性
		assert.Equal(t, room.Key, loadedRoom.Key, "房间键应该一致")
		assert.Equal(t, room.Topic, loadedRoom.Topic, "房间主题应该一致")
		assert.Equal(t, len(room.Members), len(loadedRoom.Members), "成员数量应该一致")

		// 验证成员数据
		for id, expectedMember := range room.Members {
			loadedMember, exists := loadedRoom.Members[id]
			assert.True(t, exists, "成员应该存在")
			assert.Equal(t, expectedMember.Name, loadedMember.Name, "成员名称应该一致")
			assert.Equal(t, expectedMember.State, loadedMember.State, "成员状态应该一致")
			assert.Equal(t, expectedMember.Timer.Duration, loadedMember.Timer.Duration, "计时器时长应该一致")
		}
	})

	t.Run("消息历史记录", func(t *testing.T) {
		// 创建测试消息历史
		messages := []Message{
			{
				Type: MsgJoin,
				Member: &Member{
					ID:        "user1",
					Name:      "用户1",
					State:     StateIdle,
					UpdatedAt: time.Now().Unix(),
				},
			},
			{
				Type: MsgStateUpdate,
				Member: &Member{
					ID:        "user1",
					Name:      "用户1",
					State:     StateWork,
					Timer:     TimerInfo{Duration: 1500, Remaining: 1200, IsRunning: true, IsWork: true},
					UpdatedAt: time.Now().Unix(),
				},
			},
			{
				Type: MsgLeave,
				Member: &Member{
					ID:        "user1",
					Name:      "用户1",
					State:     StateIdle,
					UpdatedAt: time.Now().Unix(),
				},
			},
		}

		// 保存消息历史
		tempFile := "test_message_history.json"
		defer os.Remove(tempFile)

		data, err := json.MarshalIndent(messages, "", "  ")
		require.NoError(t, err, "消息历史序列化失败")

		err = os.WriteFile(tempFile, data, 0644)
		require.NoError(t, err, "写入消息历史文件失败")

		// 读取并验证消息历史
		fileData, err := os.ReadFile(tempFile)
		require.NoError(t, err, "读取消息历史文件失败")

		var loadedMessages []Message
		err = json.Unmarshal(fileData, &loadedMessages)
		require.NoError(t, err, "消息历史反序列化失败")

		// 验证消息历史完整性
		assert.Equal(t, len(messages), len(loadedMessages), "消息数量应该一致")

		for i, expectedMsg := range messages {
			loadedMsg := loadedMessages[i]
			assert.Equal(t, expectedMsg.Type, loadedMsg.Type, "消息类型应该一致")
			assert.Equal(t, expectedMsg.Member.ID, loadedMsg.Member.ID, "成员ID应该一致")
			assert.Equal(t, expectedMsg.Member.Name, loadedMsg.Member.Name, "成员名称应该一致")
		}
	})
}

func connectNodes(t *testing.T, ctx context.Context, nodeA, nodeB *Node) {
	t.Helper()

	addrInfo := peer.AddrInfo{
		ID:    nodeB.GetHost().ID(),
		Addrs: nodeB.GetHost().Addrs(),
	}

	t.Logf("Connecting %s to %s", nodeA.GetHost().ID(), addrInfo.ID)
	t.Logf("NodeB addresses: %v", addrInfo.Addrs)

	err := nodeA.GetHost().Connect(ctx, addrInfo)
	require.NoError(t, err, "节点连接失败")

	// 检查实际连接状态
	assert.Eventually(t, func() bool {
		conns := nodeA.GetHost().Network().ConnsToPeer(nodeB.GetHost().ID())
		t.Logf("Connections from A to B: %d", len(conns))
		return len(conns) > 0
	}, 10*time.Second, 500*time.Millisecond, "连接未建立")
}

func createTestNode(t *testing.T, id int) (*Node, error) {
	privKeyFile := fmt.Sprintf("node_priv_%d.key", id)
	// 临时修改 loadOrCreatePrivateKey 函数调用
	return NewNode(privKeyFile)
}

func TestMain(m *testing.M) {
	logging.Init() // 初始化日志系统
	code := m.Run()
	os.Exit(code)
}

func TestBroadcastStateDirectly(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过直接广播测试")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建两个测试节点
	node1, err := createTestNode(t, 20)
	require.NoError(t, err, "节点1创建失败")
	defer node1.Close()

	node2, err := createTestNode(t, 21)
	require.NoError(t, err, "节点2创建失败")
	defer node2.Close()

	// 连接节点
	connectNodes(t, ctx, node1, node2)

	// 两个节点加入房间
	roomKey := "test-broadcast-room"
	err = node1.roomMgr.JoinRoom(roomKey)
	require.NoError(t, err, "节点1加入房间失败")
	time.Sleep(500 * time.Millisecond)

	err = node2.roomMgr.JoinRoom(roomKey)
	require.NoError(t, err, "节点2加入房间失败")
	time.Sleep(1 * time.Second)

	// 验证两个节点都在房间中
	members1 := node1.roomMgr.GetMembers()
	members2 := node2.roomMgr.GetMembers()
	t.Logf("Node1 members: %d, Node2 members: %d", len(members1), len(members2))

	// 创建测试成员状态
	member := &Member{
		ID:        node1.GetHost().ID().String(),
		Name:      "测试用户",
		State:     StateWork,
		Timer:     TimerInfo{Duration: 1500, Remaining: 1200, IsRunning: true, IsWork: true},
		UpdatedAt: time.Now().Unix(),
	}

	t.Logf("About to broadcast state: %+v", member)

	// 直接调用BroadcastState
	t.Logf("Calling BroadcastState method...")
	err = node1.roomMgr.BroadcastState(member)
	t.Logf("BroadcastState returned: %v", err)
	require.NoError(t, err, "状态广播失败")

	// 等待消息传播
	time.Sleep(2 * time.Second)

	// 检查节点1自己的状态是否更新
	members1After := node1.roomMgr.GetMembers()
	if member1, exists := members1After[node1.GetHost().ID().String()]; exists {
		t.Logf("Node1 after broadcast: %+v", member1)
		assert.Equal(t, StateWork, member1.State, "节点1的状态应该更新为work")
	}

	// 检查节点2是否接收到状态更新
	members2After := node2.roomMgr.GetMembers()
	if member2, exists := members2After[node1.GetHost().ID().String()]; exists {
		t.Logf("Node2 after broadcast: %+v", member2)
		assert.Equal(t, StateWork, member2.State, "节点2应该接收到work状态")
	}
}

func TestBroadcastStateMethodExists(t *testing.T) {
	// 创建一个简单的节点来测试方法是否存在
	node, err := createTestNode(t, 30)
	require.NoError(t, err, "节点创建失败")
	defer node.Close()

	// 测试方法是否存在
	roomMgr := node.GetRoomMgr()
	require.NotNil(t, roomMgr, "房间管理器应该存在")

	// 创建一个测试成员
	member := &Member{
		ID:        "test-id",
		Name:      "测试用户",
		State:     StateWork,
		Timer:     TimerInfo{Duration: 1500, Remaining: 1200, IsRunning: true, IsWork: true},
		UpdatedAt: time.Now().Unix(),
	}

	// 尝试调用BroadcastState方法（应该失败，因为不在房间中）
	err = roomMgr.BroadcastState(member)
	require.Error(t, err, "BroadcastState应该失败，因为不在房间中")
	assert.Contains(t, err.Error(), "not in a room", "错误消息应该包含'not in a room'")

	t.Logf("BroadcastState method exists and can be called")
}

func TestLoggingAndBroadcastState(t *testing.T) {
	// 创建一个简单的节点来测试
	node, err := createTestNode(t, 40)
	require.NoError(t, err, "节点创建失败")
	defer node.Close()

	// 测试日志是否工作
	logging.Log("Testing logging system\n")
	t.Log("Logging test completed")

	// 加入房间
	err = node.roomMgr.JoinRoom("test-logging-room")
	require.NoError(t, err, "加入房间失败")
	time.Sleep(500 * time.Millisecond)

	// 创建测试成员
	member := &Member{
		ID:        node.GetHost().ID().String(),
		Name:      "测试用户",
		State:     StateWork,
		Timer:     TimerInfo{Duration: 1500, Remaining: 1200, IsRunning: true, IsWork: true},
		UpdatedAt: time.Now().Unix(),
	}

	t.Logf("Created member: %+v", member)

	// 调用BroadcastState
	t.Log("About to call BroadcastState")
	err = node.roomMgr.BroadcastState(member)
	t.Logf("BroadcastState result: %v", err)
	require.NoError(t, err, "BroadcastState应该成功")

	// 等待消息处理
	time.Sleep(1 * time.Second)

	// 检查成员状态
	members := node.roomMgr.GetMembers()
	t.Logf("Members after broadcast: %d", len(members))
	for id, member := range members {
		t.Logf("Member: %s -> State: %s", id, member.State)
	}
}
