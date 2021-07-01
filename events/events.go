package events

import (
	"github.com/gamemann/Pterodactyl-Packet-Watch/config"
	"github.com/gamemann/Pterodactyl-Packet-Watch/misc"
)

func OnDetect(cfg *config.Config, srv *config.Server, pckt *config.Packet, avglatency uint32, maxlatency uint32, minlatency uint32, detects uint, laststats []uint32) {
	// Handle Misc options.
	misc.HandleMisc(cfg, srv, pckt, avglatency, maxlatency, minlatency, detects, laststats)
}
