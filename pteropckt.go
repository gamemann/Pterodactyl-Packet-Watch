package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gamemann/Pterodactyl-Packet-Watch/config"
	"github.com/gamemann/Pterodactyl-Packet-Watch/pterodactyl"
	"github.com/gamemann/Pterodactyl-Packet-Watch/servers"
	"github.com/gamemann/Pterodactyl-Packet-Watch/update"
)

func main() {
	// Look for 'cfg' flag in command line arguments (default path: /etc/pteropckt/pteropcktconf).
	configFile := flag.String("cfg", "/etc/pteropckt/pteropckt.conf", "The path to the Pterowatch config file.")
	flag.Parse()

	// Create config struct.
	cfg := config.Config{}

	// Set config defaults.
	cfg.SetDefaults()

	// Attempt to read config.
	cfg.ReadConfig(*configFile)

	// Level 1 debug.
	if cfg.DebugLevel > 0 {
		fmt.Printf("[D1] Found config with API URL => %s. Token => %s. App Token => %s. Auto Add Servers => %t. Debug level => %d. Reload time => %d.\n", cfg.APIURL, cfg.Token, cfg.AppToken, cfg.AddServers, cfg.DebugLevel, cfg.ReloadTime)
	}

	// Level 2 debug.
	if cfg.DebugLevel > 1 {
		fmt.Printf("[D2] Config default server values. Enable => %t. Threshold => %d. Count => %d. Interval => %d. Timeout => %d. Max Detects => %d. Cooldown => %d. Mentions => %s.\n", cfg.DefEnable, cfg.DefThreshold, cfg.DefCount, cfg.DefInterval, cfg.DefTimeout, cfg.DefMaxDetects, cfg.DefCooldown, cfg.DefMentions)
	}

	// Check if we want to automatically add servers.
	if cfg.AddServers {
		pterodactyl.AddServers(&cfg)
	}

	// Handle all servers (create timers, etc.).
	servers.HandleServers(&cfg, false)

	// Set config file for use later (e.g. updating/reloading).
	cfg.ConfLoc = *configFile

	// Initialize updater/reloader.
	update.Init(&cfg)

	// Signal.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
}
