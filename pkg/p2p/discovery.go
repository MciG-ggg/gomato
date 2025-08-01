package p2p

import (
	"context"
	"fmt"
	"gomato/pkg/logging"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

// DiscoveryService 节点发现服务
type DiscoveryService struct {
	host        host.Host
	ctx         context.Context
	cancel      context.CancelFunc
	mdns        mdns.Service
	peers       map[peer.ID]*PeerInfo
	peersMu     sync.RWMutex
	onPeerFound func(peer.ID)
	onPeerLost  func(peer.ID)
}

// PeerInfo 节点信息
type PeerInfo struct {
	ID        peer.ID     `json:"id"`
	Addrs     []string    `json:"addrs"`
	Connected bool        `json:"connected"`
	LastSeen  time.Time   `json:"lastSeen"`
	State     MemberState `json:"state,omitempty"`
}

// NewDiscoveryService 创建新的发现服务
func NewDiscoveryService(host host.Host, onPeerFound, onPeerLost func(peer.ID)) *DiscoveryService {
	ctx, cancel := context.WithCancel(context.Background())

	ds := &DiscoveryService{
		host:        host,
		ctx:         ctx,
		cancel:      cancel,
		peers:       make(map[peer.ID]*PeerInfo),
		onPeerFound: onPeerFound,
		onPeerLost:  onPeerLost,
	}

	// 创建mDNS服务
	ds.mdns = mdns.NewMdnsService(host, "gomato-p2p", ds)

	return ds
}

// Start 启动发现服务
func (ds *DiscoveryService) Start() error {
	logging.Log("🚀 启动节点发现服务...\n")

	// 启动mDNS服务
	if err := ds.mdns.Start(); err != nil {
		return fmt.Errorf("failed to start mDNS service: %w", err)
	}

	// 启动连接监控
	go ds.monitorConnections()

	// 启动节点清理
	go ds.cleanupPeers()

	logging.Log("✅ 节点发现服务已启动\n")
	return nil
}

// Stop 停止发现服务
func (ds *DiscoveryService) Stop() error {
	logging.Log("🛑 停止节点发现服务...\n")

	ds.cancel()

	if ds.mdns != nil {
		ds.mdns.Close()
	}

	logging.Log("✅ 节点发现服务已停止\n")
	return nil
}

// HandlePeerFound 处理发现的节点 (实现 mdns.Notifee 接口)
func (ds *DiscoveryService) HandlePeerFound(pi peer.AddrInfo) {
	logging.Log(fmt.Sprintf("🔍 发现节点: %s\n", pi.ID.String()))

	ds.peersMu.Lock()
	defer ds.peersMu.Unlock()

	// 检查是否已存在
	if _, exists := ds.peers[pi.ID]; exists {
		// 更新现有节点信息
		ds.peers[pi.ID].Addrs = peerAddrsToStrings(pi.Addrs)
		ds.peers[pi.ID].LastSeen = time.Now()
		return
	}

	// 创建新节点信息
	peerInfo := &PeerInfo{
		ID:       pi.ID,
		Addrs:    peerAddrsToStrings(pi.Addrs),
		LastSeen: time.Now(),
		State:    StateIdle,
	}

	ds.peers[pi.ID] = peerInfo

	// 尝试连接
	go ds.connectToPeer(pi)

	// 通知回调
	if ds.onPeerFound != nil {
		ds.onPeerFound(pi.ID)
	}
}

// connectToPeer 连接到指定节点
func (ds *DiscoveryService) connectToPeer(pi peer.AddrInfo) {
	logging.Log(fmt.Sprintf("🔗 尝试连接到节点: %s\n", pi.ID.String()))

	ctx, cancel := context.WithTimeout(ds.ctx, 10*time.Second)
	defer cancel()

	if err := ds.host.Connect(ctx, pi); err != nil {
		logging.Log(fmt.Sprintf("❌ 连接节点 %s 失败: %v\n", pi.ID.String(), err))
		return
	}

	// 更新连接状态
	ds.peersMu.Lock()
	if peerInfo, exists := ds.peers[pi.ID]; exists {
		peerInfo.Connected = true
		peerInfo.LastSeen = time.Now()
	}
	ds.peersMu.Unlock()

	logging.Log(fmt.Sprintf("✅ 成功连接到节点: %s\n", pi.ID.String()))
}

// monitorConnections 监控连接状态
func (ds *DiscoveryService) monitorConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.checkConnections()
		}
	}
}

// checkConnections 检查连接状态
func (ds *DiscoveryService) checkConnections() {
	ds.peersMu.Lock()
	defer ds.peersMu.Unlock()

	for peerID, peerInfo := range ds.peers {
		// 检查连接是否仍然有效
		conns := ds.host.Network().ConnsToPeer(peerID)
		wasConnected := peerInfo.Connected
		peerInfo.Connected = len(conns) > 0

		// 如果连接状态发生变化
		if wasConnected && !peerInfo.Connected {
			logging.Log(fmt.Sprintf("❌ 节点 %s 连接断开\n", peerID.String()))
			if ds.onPeerLost != nil {
				ds.onPeerLost(peerID)
			}
		} else if !wasConnected && peerInfo.Connected {
			logging.Log(fmt.Sprintf("✅ 节点 %s 重新连接\n", peerID.String()))
		}

		if peerInfo.Connected {
			peerInfo.LastSeen = time.Now()
		}
	}
}

// cleanupPeers 清理长时间未见的节点
func (ds *DiscoveryService) cleanupPeers() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.peersMu.Lock()
			now := time.Now()
			for peerID, peerInfo := range ds.peers {
				// 如果节点超过5分钟未见到，则移除
				if now.Sub(peerInfo.LastSeen) > 5*time.Minute {
					logging.Log(fmt.Sprintf("🗑️ 清理过期节点: %s\n", peerID.String()))
					delete(ds.peers, peerID)
				}
			}
			ds.peersMu.Unlock()
		}
	}
}

// GetPeers 获取所有已知节点
func (ds *DiscoveryService) GetPeers() map[peer.ID]*PeerInfo {
	ds.peersMu.RLock()
	defer ds.peersMu.RUnlock()

	peers := make(map[peer.ID]*PeerInfo)
	for id, info := range ds.peers {
		peers[id] = info
	}
	return peers
}

// GetConnectedPeers 获取已连接的节点
func (ds *DiscoveryService) GetConnectedPeers() []peer.ID {
	ds.peersMu.RLock()
	defer ds.peersMu.RUnlock()

	var connected []peer.ID
	for id, info := range ds.peers {
		if info.Connected {
			connected = append(connected, id)
		}
	}
	return connected
}

// UpdatePeerState 更新节点状态
func (ds *DiscoveryService) UpdatePeerState(peerID peer.ID, state MemberState) {
	ds.peersMu.Lock()
	defer ds.peersMu.Unlock()

	if peerInfo, exists := ds.peers[peerID]; exists {
		peerInfo.State = state
		peerInfo.LastSeen = time.Now()
	}
}

// GetPeerInfo 获取指定节点信息
func (ds *DiscoveryService) GetPeerInfo(peerID peer.ID) *PeerInfo {
	ds.peersMu.RLock()
	defer ds.peersMu.RUnlock()

	return ds.peers[peerID]
}

// peerAddrsToStrings 将peer地址转换为字符串切片
func peerAddrsToStrings(addrs []multiaddr.Multiaddr) []string {
	var result []string
	for _, addr := range addrs {
		result = append(result, addr.String())
	}
	return result
}
