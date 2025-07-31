package p2p

import (
	"context"
	"fmt"
	"gomato/pkg/logging"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

type Node struct {
	host      host.Host
	ctx       context.Context
	cancel    context.CancelFunc
	roomMgr   *RoomManager
	discovery *DiscoveryService
}

func NewNode(keyPath string) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if keyPath == "" {
		keyPath = "node_priv.key" // 默认密钥路径
	}
	// 加载或创建私钥
	privKey, err := loadOrCreatePrivateKey(keyPath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// 正确的 libp2p 配置方式
	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.Security(noise.ID, noise.New), // 只使用 Noise 安全传输
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		// 特别注意：不要混用 DefaultSecurity 和显式 Security 配置
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	node := &Node{
		host:   h,
		ctx:    ctx,
		cancel: cancel,
	}

	node.roomMgr = NewRoomManager(node)

	// 创建发现服务
	node.discovery = NewDiscoveryService(
		node.host,
		func(peerID peer.ID) {
			logging.Log(fmt.Sprintf("🎉 发现新节点: %s\n", peerID.String()))
		},
		func(peerID peer.ID) {
			logging.Log(fmt.Sprintf("👋 节点断开: %s\n", peerID.String()))
		},
	)

	// 启动发现服务
	go func() {
		time.Sleep(100 * time.Millisecond) // 确保主机完全初始化
		if err := node.discovery.Start(); err != nil {
			logging.Log(fmt.Sprintf("Failed to start discovery: %v", err))
		}
	}()

	return node, nil
}

func (n *Node) HandlePeerFound(pi peer.AddrInfo) {
	// 委托给发现服务处理
	if n.discovery != nil {
		n.discovery.HandlePeerFound(pi)
	}
}

func (n *Node) startDiscovery() error {
	// 这个方法现在委托给 DiscoveryService
	if n.discovery != nil {
		return n.discovery.Start()
	}
	return fmt.Errorf("discovery service not initialized")
}

func (n *Node) Close() error {
	n.cancel()

	// 停止发现服务
	if n.discovery != nil {
		n.discovery.Stop()
	}

	return n.host.Close()
}

func (n *Node) GetHost() host.Host {
	return n.host
}

func (n *Node) GetRoomMgr() *RoomManager {
	return n.roomMgr
}

func (n *Node) GetDiscovery() *DiscoveryService {
	return n.discovery
}

func (n *Node) GetHostID() string {
	return n.host.ID().String()
}
