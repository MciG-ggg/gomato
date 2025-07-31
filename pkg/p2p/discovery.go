package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type DiscoveryMessage struct {
	Type        string    `json:"type"`
	AppName     string    `json:"app_name"`
	AppID       string    `json:"app_id"`
	ListenPort  int       `json:"listen_port"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}
