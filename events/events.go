package events

import (
	"github.com/GFLClan/Pterodactyl-PacketWatch/config"
	"github.com/GFLClan/Pterodactyl-PacketWatch/misc"
)

func OnServerDown(cfg *config.Config, srv *config.Server, fails int, restarts int) {
	// Handle Misc options.
	misc.HandleMisc(cfg, srv, fails, restarts)
}
