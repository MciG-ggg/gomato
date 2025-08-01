package gomato

import (
	"gomato/pkg/p2p"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoomUIModel(t *testing.T) {
	// 创建模拟的房间管理器
	node, err := p2p.NewNode("")
	require.NoError(t, err)
	defer node.Close()

	roomManager := node.GetRoomMgr()
	model := NewRoomUIModel(roomManager)

	// 测试初始化
	require.NotNil(t, model.roomInput)
	require.NotNil(t, model.roomManager)
	require.Equal(t, "输入房间密钥...", model.roomInput.Placeholder)
	require.Equal(t, roomHidden, model.state)
	require.NotNil(t, model.joinOrderMap)
	require.Equal(t, 1, model.nextJoinOrder)
}

func TestRoomUIModelInRoom(t *testing.T) {
	// 创建模拟的房间管理器
	node, err := p2p.NewNode("")
	require.NoError(t, err)
	defer node.Close()

	roomManager := node.GetRoomMgr()
	model := NewRoomUIModel(roomManager)

	// 测试初始状态
	require.False(t, model.IsInRoom())

	// 测试加入房间
	err = roomManager.JoinRoom("test-room")
	require.NoError(t, err)

	// 更新模型状态以反映房间加入
	model.state = roomVisible
	require.True(t, model.IsInRoom())
}

func TestRoomUIModelStateText(t *testing.T) {
	// 测试状态文本转换
	require.Equal(t, "工作中", getStateText(p2p.StateWork))
	require.Equal(t, "休息中", getStateText(p2p.StateRest))
	require.Equal(t, "空闲", getStateText(p2p.StateIdle))
	require.Equal(t, "未知", getStateText(p2p.MemberState("unknown"))) // 未知状态
}

func TestRoomUIModelTimerText(t *testing.T) {
	// 测试计时器文本转换
	timer := p2p.TimerInfo{
		Duration:  1500, // 25分钟
		Remaining: 125,  // 2分5秒
		IsRunning: true,
		IsWork:    true,
	}

	require.Equal(t, "02:05", getTimerText(timer))

	// 测试计时器未运行的情况
	timer.IsRunning = false
	require.Equal(t, "", getTimerText(timer))
}

func TestMemberItem(t *testing.T) {
	// 测试MemberItem结构体
	member := &p2p.Member{
		ID:    "test-id",
		Name:  "测试用户",
		State: p2p.StateWork,
		Timer: p2p.TimerInfo{
			Duration:  1500,
			Remaining: 125,
			IsRunning: true,
			IsWork:    true,
		},
		TaskName: "测试任务",
	}

	item := MemberItem{
		Member:    member,
		JoinOrder: 1,
		IsSelf:    false,
	}

	// 测试Title方法（包含任务名）
	expectedTitle := "测试用户 [工作中] 02:05 - 测试任务"
	require.Equal(t, expectedTitle, item.Title())

	// 测试Description方法
	expectedDesc := "ID: test-id"
	require.Equal(t, expectedDesc, item.Description())

	// 测试FilterValue方法
	require.Equal(t, "测试用户", item.FilterValue())

	// 测试没有任务名的情况
	member.TaskName = ""
	expectedTitleNoTask := "测试用户 [工作中] 02:05"
	require.Equal(t, expectedTitleNoTask, item.Title())

	// 测试IsSelf字段
	require.False(t, item.IsSelf)

	// 测试当前用户的情况
	item.IsSelf = true
	require.True(t, item.IsSelf)
}

func TestRoomUIModelKeyMap(t *testing.T) {
	// 创建模拟的房间管理器
	node, err := p2p.NewNode("")
	require.NoError(t, err)
	defer node.Close()

	roomManager := node.GetRoomMgr()
	model := NewRoomUIModel(roomManager)

	// 测试快捷键映射是否正确创建
	require.NotNil(t, model.roomKeys)
	require.NotNil(t, model.roomKeys.LeaveRoom)

	// 测试快捷键绑定
	require.Equal(t, "q/esc", model.roomKeys.LeaveRoom.Help().Key)
}
