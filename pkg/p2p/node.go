package p2p

import (
	"context"
	"fmt"
	"gomato/pkg/logging"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

type Node struct {
	host    host.Host
	ctx     context.Context
	cancel  context.CancelFunc
	roomMgr *RoomManager
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

	// 启动发现服务
	go func() {
		time.Sleep(100 * time.Millisecond) // 确保主机完全初始化
		if err := node.startDiscovery(); err != nil {
			logging.Log(fmt.Sprintf("Failed to start discovery: %v", err))
		}
	}()

	return node, nil
}

func (n *Node) HandlePeerFound(pi peer.AddrInfo) {
	logging.Log(fmt.Sprintf("✅ mDNS 发现节点: %s\n", pi.ID.String()))

	// 尝试连接到发现的节点
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := n.host.Connect(ctx, pi); err != nil {
		logging.Log(fmt.Sprintf("❌ 连接节点 %s 失败: %v\n", pi.ID.String(), err))
	} else {
		logging.Log(fmt.Sprintf("🎉 成功连接到节点: %s\n", pi.ID.String()))
	}
}

func (n *Node) startDiscovery() error {
	// 使用mDNS进行局域网发现
	service := mdns.NewMdnsService(n.host, "gomato-p2p", n)

	// 启动 mDNS 服务
	return service.Start()
}

func (n *Node) Close() error {
	n.cancel()
	return n.host.Close()
}

func (n *Node) GetHost() host.Host {
	return n.host
}

func (n *Node) GetRoomMgr() *RoomManager {
	return n.roomMgr
}
