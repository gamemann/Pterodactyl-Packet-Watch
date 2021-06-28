package servers

import (
	"net"
	"time"

	"github.com/GFLClan/Pterodactyl-PacketWatch/config"
)

type Tuple struct {
	IP     string
	Port   int
	UID    string
	PcktID int
}

type Stats struct {
	LastStats  *[]uint32
	AvgLatency *uint32
	MaxLatency *uint32
	MinLatency *uint32
	Detects    *uint
}

type TickerHolder struct {
	Srv       Tuple
	Ticker    *time.Ticker
	Conn      *net.UDPConn
	Destroyer *chan bool
	Stats     Stats
}

func RemoveTicker(t *[]TickerHolder, idx int) {
	copy((*t)[idx:], (*t)[idx+1:])
	*t = (*t)[:len(*t)-1]
}

func RemoveServer(cfg *config.Config, idx int) {
	copy(cfg.Servers[idx:], cfg.Servers[idx+1:])
	cfg.Servers = cfg.Servers[:len(cfg.Servers)-1]
}

func RemoveStat(t *[]uint32, idx int) {
	copy((*t)[idx:], (*t)[idx+1:])
	*t = (*t)[:len(*t)-1]
}
