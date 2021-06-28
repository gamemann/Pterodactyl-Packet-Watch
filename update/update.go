package update

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GFLClan/Pterodactyl-PacketWatch/config"
	"github.com/GFLClan/Pterodactyl-PacketWatch/pterodactyl"
	"github.com/GFLClan/Pterodactyl-PacketWatch/servers"
)

type Tuple struct {
	IP   string
	Port int
	UID  string
}

var updateticker *time.Ticker

func AddNewServers(newcfg *config.Config, cfg *config.Config) {
	// Loop through all new servers.
	for _, newsrv := range newcfg.Servers {
		if cfg.DebugLevel > 3 {
			fmt.Printf("[D4] Looking for %s:%d:%s (%s) inside of old configuration.\n", newsrv.IP, newsrv.Port, newsrv.UID, newsrv.Name)
		}

		toadd := true

		// Now loop through old servers.
		for j, oldsrv := range cfg.Servers {
			// Create new server tuple.
			var nt Tuple
			nt.IP = newsrv.IP
			nt.Port = int(newsrv.Port)
			nt.UID = newsrv.UID

			// Create old server tuple.
			var ot Tuple
			ot.IP = oldsrv.IP
			ot.Port = int(oldsrv.Port)
			ot.UID = oldsrv.UID

			if cfg.DebugLevel > 4 {
				fmt.Printf("[D5] Comparing %s:%d:%s == %s:%d:%s.\n", nt.IP, nt.Port, nt.UID, ot.IP, ot.Port, ot.UID)
			}

			// Now compare.
			if nt == ot {
				// We don't have to insert this server into the slice.
				toadd = false

				if cfg.DebugLevel > 2 {
					fmt.Printf("[D3] Found matching server %s:%d:%s (%s) on Add Server check. Applying new configuration. Enabled: %t => %t. Threshold: %d => %d. Count: %d => %d. Interval: %d => %d. Timeout: %d => %d. Max Detections: %d => %d. Cooldown: %d => %d.\n", newsrv.IP, newsrv.Port, newsrv.UID, newsrv.Name, oldsrv.Enable, newsrv.Enable, oldsrv.Threshold, newsrv.Threshold, oldsrv.Count, newsrv.Count, oldsrv.Interval, newsrv.Interval, oldsrv.Timeout, newsrv.Timeout, oldsrv.MaxDetects, newsrv.MaxDetects, oldsrv.Cooldown, newsrv.Cooldown)
				}

				// Update specific configuration.
				cfg.Servers[j].Enable = newsrv.Enable
				cfg.Servers[j].Threshold = newsrv.Threshold
				cfg.Servers[j].Count = newsrv.Count
				cfg.Servers[j].Interval = newsrv.Interval
				cfg.Servers[j].Timeout = newsrv.Timeout
				cfg.Servers[j].MaxDetects = newsrv.MaxDetects
				cfg.Servers[j].Cooldown = newsrv.Cooldown
				cfg.Servers[j].Mentions = newsrv.Mentions
			}
		}

		// If we're not inside of the current configuration, add the server.
		if toadd {
			if cfg.DebugLevel > 1 {
				fmt.Printf("[D3] Adding server %s:%d:%s (%s). Enabled => %t. Threshold => %d. Count => %d. Interval => %d. Timeout => %d. Max Detections => %d. Cooldown  => %d.\n", newsrv.IP, newsrv.Port, newsrv.UID, newsrv.Name, newsrv.Enable, newsrv.Threshold, newsrv.Count, newsrv.Interval, newsrv.Timeout, newsrv.MaxDetects, newsrv.Cooldown)
			}

			cfg.Servers = append(cfg.Servers, newsrv)
		}
	}
}

func DelOldServers(newcfg *config.Config, cfg *config.Config) {
	// Loop through all old servers.
	for i, oldsrv := range cfg.Servers {
		if cfg.DebugLevel > 3 {
			fmt.Printf("[D4] Looking for %s:%d:%s (%s) inside of new configuration.\n", oldsrv.IP, oldsrv.Port, oldsrv.UID, oldsrv.Name)
		}

		todel := true

		// Now loop through new servers.
		for _, newsrv := range newcfg.Servers {
			// Create old server tuple.
			var ot Tuple
			ot.IP = oldsrv.IP
			ot.Port = int(oldsrv.Port)
			ot.UID = oldsrv.UID

			// Create new server tuple.
			var nt Tuple
			nt.IP = newsrv.IP
			nt.Port = int(newsrv.Port)
			nt.UID = newsrv.UID

			if cfg.DebugLevel > 4 {
				fmt.Printf("[D5] Comparing %s:%d:%s == %s:%d:%s\n", ot.IP, ot.Port, ot.UID, nt.IP, nt.Port, nt.UID)
			}

			// Now compare.
			if nt == ot {
				todel = false
			}
		}

		// If we're not inside of the new configuration, delete the server.
		if todel {
			if cfg.DebugLevel > 1 {
				fmt.Println("[D2] Deleting server from update %s:%d:%s (%s).\n", oldsrv.IP, oldsrv.Port, oldsrv.UID, oldsrv.Name)
			}

			// Set Delete to true so we'll delete the server, close the connection, etc. on the next scan.
			cfg.Servers[i].Delete = true
		}
	}
}

func ReloadServers(timer *time.Ticker, cfg *config.Config) {
	destroy := make(chan struct{})

	for {
		select {
		case <-timer.C:
			// First, we'll want to read the new config.
			newcfg := config.Config{}

			// Set default values.
			newcfg.SetDefaults()

			success := newcfg.ReadConfig(cfg.ConfLoc)

			if !success {
				continue
			}

			if newcfg.AddServers {
				cont := pterodactyl.AddServers(&newcfg)

				if !cont {
					fmt.Println("[ERR] Not updating server list due to error.")

					continue
				}
			}

			// Assign new values.
			cfg.APIURL = newcfg.APIURL
			cfg.Token = newcfg.Token
			cfg.DebugLevel = newcfg.DebugLevel
			cfg.AddServers = newcfg.AddServers

			cfg.DefEnable = newcfg.DefEnable
			cfg.DefThreshold = newcfg.DefThreshold
			cfg.DefCount = newcfg.DefCount
			cfg.DefInterval = newcfg.DefInterval
			cfg.DefTimeout = newcfg.DefTimeout
			cfg.DefMaxDetects = newcfg.DefMaxDetects
			cfg.DefCooldown = newcfg.DefCooldown

			// If reload time is different, recreate reload timer.
			if cfg.ReloadTime != newcfg.ReloadTime {
				if updateticker != nil {
					updateticker.Stop()
				}

				if cfg.DebugLevel > 2 {
					fmt.Println("[D3] Recreating update timer due to updated reload time (" + strconv.Itoa(int(cfg.ReloadTime)) + " => " + strconv.Itoa(int(newcfg.ReloadTime)) + ").")
				}

				// Create repeating timer.
				updateticker = time.NewTicker(time.Duration(newcfg.ReloadTime) * time.Second)
				go ReloadServers(updateticker, cfg)
			}

			cfg.ReloadTime = newcfg.ReloadTime

			// Level 2 debug message.
			if cfg.DebugLevel > 1 {
				fmt.Println("[D2] Updating servers.")
			}

			// Add new servers.
			AddNewServers(&newcfg, cfg)

			// Remove servers that are not a part of new configuration.
			DelOldServers(&newcfg, cfg)

			// Now rehandle servers.
			servers.HandleServers(cfg, true)

		case <-destroy:
			timer.Stop()

			return
		}
	}
}

func Init(cfg *config.Config) {
	if cfg.ReloadTime < 1 {
		return
	}

	if cfg.DebugLevel > 0 {
		fmt.Println("[D1] Setting up reload timer for every " + strconv.Itoa(int(cfg.ReloadTime)) + " seconds.")
	}

	// Create repeating timer.
	updateticker = time.NewTicker(time.Duration(cfg.ReloadTime) * time.Second)
	go ReloadServers(updateticker, cfg)
}
