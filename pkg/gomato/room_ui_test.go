package gomato

import (
	"gomato/pkg/p2p"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoomEnterInputModel(t *testing.T) {
	// 创建模拟的房间管理器
	node, err := p2p.NewNode("")
	require.NoError(t, err)
	defer node.Close()

	roomManager := node.GetRoomMgr()
	model := NewRoomEnterInputModel(roomManager)

	// 测试初始化
	require.NotNil(t, model.roomInput)
	require.NotNil(t, model.roomManager)
	require.NotNil(t, model.keys)
	require.Equal(t, "输入房间密钥...", model.roomInput.Placeholder)
}

func TestRoomViewModel(t *testing.T) {
	// 创建模拟的房间管理器
	node, err := p2p.NewNode("")
	require.NoError(t, err)
	defer node.Close()

	roomManager := node.GetRoomMgr()
	model := NewRoomViewModel(roomManager)

	// 测试初始化
	require.False(t, model.showMembers)
	require.NotNil(t, model.roomManager)
	require.NotNil(t, model.keys)

	// 测试初始状态
	require.False(t, model.IsInRoom())

	// 测试加入房间
	err = roomManager.JoinRoom("test-room")
	require.NoError(t, err)
	require.True(t, model.IsInRoom())
}

func TestRoomViewModelStateText(t *testing.T) {
	node, err := p2p.NewNode("")
	require.NoError(t, err)
	defer node.Close()

	roomManager := node.GetRoomMgr()
	model := NewRoomViewModel(roomManager)

	// 测试状态文本转换
	require.Equal(t, "工作中", model.getStateText(p2p.StateWork))
	require.Equal(t, "休息中", model.getStateText(p2p.StateRest))
	require.Equal(t, "空闲", model.getStateText(p2p.StateIdle))
	require.Equal(t, "未知", model.getStateText(p2p.MemberState("unknown"))) // 未知状态
}

func TestRoomViewModelTimerText(t *testing.T) {
	node, err := p2p.NewNode("")
	require.NoError(t, err)
	defer node.Close()

	roomManager := node.GetRoomMgr()
	model := NewRoomViewModel(roomManager)

	// 测试计时器文本转换
	timer := p2p.TimerInfo{
		Duration:  1500, // 25分钟
		Remaining: 125,  // 2分5秒
		IsRunning: true,
		IsWork:    true,
	}

	require.Equal(t, "02:05", model.getTimerText(timer))

	// 测试计时器未运行的情况
	timer.IsRunning = false
	require.Equal(t, "", model.getTimerText(timer))
}
