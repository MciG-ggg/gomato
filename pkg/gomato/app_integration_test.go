package gomato

import (
	"gomato/pkg/common"
	"gomato/pkg/logging"
	"gomato/pkg/p2p"
	"gomato/pkg/task"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAppP2PIntegration(t *testing.T) {
	// 初始化日志
	home, _ := os.UserHomeDir()
	logPath := home + "/.gomato/app_integration.log"
	os.Remove(logPath)

	taskMgr := &task.Manager{Tasks: []task.Task{}}
	settings := common.Settings{
		Pomodoro:        25,
		ShortBreak:      5,
		LongBreak:       15,
		Cycle:           4,
		TimeDisplayMode: "ansi",
		Language:        "zh",
	}

	// 创建两个App实例，分别模拟两个用户
	// 为每个节点使用不同的密钥路径
	app1 := NewAppWithKeyPath(taskMgr, settings, "node1_priv.key")
	require.NotNil(t, app1.node, "app1 P2P节点应初始化")
	app2 := NewAppWithKeyPath(taskMgr, settings, "node2_priv.key")
	require.NotNil(t, app2.node, "app2 P2P节点应初始化")

	// 两个节点加入同一个房间
	err := app1.node.GetRoomMgr().JoinRoom("integration-room")
	require.NoError(t, err, "app1 加入房间失败")
	time.Sleep(1 * time.Second)
	err = app2.node.GetRoomMgr().JoinRoom("integration-room")
	require.NoError(t, err, "app2 加入房间失败")
	time.Sleep(3 * time.Second) // 等待同步

	// 检查成员同步
	members1 := app1.node.GetRoomMgr().GetMembers()
	members2 := app2.node.GetRoomMgr().GetMembers()
	require.GreaterOrEqual(t, len(members1), 1, "app1应看到成员")
	require.GreaterOrEqual(t, len(members2), 1, "app2应看到成员")

	// app1广播状态
	id1 := app1.node.GetHost().ID().String()
	member := members1[id1]
	member.State = p2p.StateWork
	err = app1.node.GetRoomMgr().BroadcastState(member)
	require.NoError(t, err, "app1广播状态失败")
	time.Sleep(1 * time.Second)

	// 检查app2是否收到状态
	members2 = app2.node.GetRoomMgr().GetMembers()
	if m, ok := members2[id1]; ok {
		require.Equal(t, p2p.StateWork, m.State, "app2应收到app1的work状态")
	} else {
		t.Fatalf("app2未找到app1成员")
	}

	// 清理
	app1.node.Close()
	app2.node.Close()
}

func TestMain(m *testing.M) {
	logging.Init() // 初始化日志系统
	code := m.Run()
	os.Exit(code)
}
