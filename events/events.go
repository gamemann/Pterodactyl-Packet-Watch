package events

import (
	"github.com/GFLClan/Pterodactyl-PacketWatch/config"
	"github.com/GFLClan/Pterodactyl-PacketWatch/misc"
)

func OnDetect(cfg *config.Config, srv *config.Server, pckt *config.Packet, avglatency uint32, maxlatency uint32, minlatency uint32, detects uint) {
	// Handle Misc options.
	misc.HandleMisc(cfg, srv, pckt, avglatency, maxlatency, minlatency, detects)
}
