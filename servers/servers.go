package servers

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gamemann/Pterodactyl-Packet-Watch/config"
	"github.com/gamemann/Pterodactyl-Packet-Watch/events"
	"github.com/gamemann/Pterodactyl-Packet-Watch/pterodactyl"
	"github.com/gamemann/Pterodactyl-Packet-Watch/query"
)

var tickers []TickerHolder

// Timer function.
func ServerWatch(srv *config.Server, pckt *config.Packet, timer *time.Ticker, laststats *[]uint32, avglatency *uint32, maxlatency *uint32, minlatency *uint32, detects *uint, nextscan *int64, conn *net.UDPConn, cfg *config.Config, destroy *chan bool) {
	data, err := hex.DecodeString(strings.ReplaceAll(pckt.Request, " ", ""))

	if err != nil {
		fmt.Println("Failed to parse data => " + pckt.Request)
		fmt.Println(err)

		return
	}

	for {
		select {
		case <-timer.C:
			// If the UDP connection or server is nil, break the timer.
			if conn == nil || srv == nil {
				*destroy <- true

				break
			}

			// Check if server is enabled.
			if !srv.Enable {
				continue
			}

			// Check if container status is 'on'.
			if !pterodactyl.CheckStatus(cfg, srv.UID) {
				continue
			}

			// Send request.
			query.SendRequest(conn, data)

			if cfg.DebugLevel > 2 {
				fmt.Printf("[D2] Packet %s:%d:%s (%s) sent. Request data => % x\n", srv.IP, srv.Port, srv.UID, srv.Name, data)
			}

			// Send the request and retrieve latency.
			var latency uint32

			start := time.Now()
			resp := query.CheckResponse(conn, pckt.Timeout)
			stop := time.Since(start)

			if !resp {
				if cfg.DebugLevel > 1 {
					fmt.Printf("[D2] Request timed out for %s:%d:%s (%s). Timeout => %d.\n", srv.IP, srv.Port, srv.UID, srv.Name, pckt.Timeout)
				}

				// Set latency to timeout (to milliseconds).
				latency = uint32(pckt.Timeout * 1000)
			} else {
				latency = uint32(stop.Milliseconds())
			}

			// Add into stats.
			*laststats = append(*laststats, latency)

			if cfg.DebugLevel > 1 {
				fmt.Printf("[D2] %s:%d:%s (%s). Stats => %v.\n", srv.IP, srv.Port, srv.UID, srv.Name, *laststats)
			}

			if uint(len(*laststats)) < pckt.Count/2 {
				continue
			}

			// Check if we need to remove the oldest.
			if uint(len(*laststats)) > pckt.Count {
				RemoveStat(laststats, 0)
			}

			// Calculate latencies.
			var sum uint64
			for _, stat := range *laststats {
				sum += uint64(stat)

				if stat > *maxlatency {
					*maxlatency = stat
				}

				if stat < *minlatency {
					*minlatency = stat
				}
			}

			*avglatency = uint32(sum / uint64(len(*laststats)))

			if cfg.DebugLevel > 1 {
				fmt.Printf("[D2] %s:%d:%s (%s). Average Latency => %d.\n", srv.IP, srv.Port, srv.UID, srv.Name, *avglatency)
			}

			// Check average latency.
			if *avglatency > pckt.Threshold {
				if cfg.DebugLevel > 1 {
					fmt.Printf("[D2] %s:%d:%s (%s). Detects => %d.\n", srv.IP, srv.Port, srv.UID, srv.Name, *detects)
				}

				// Check if we should report this.
				if *detects < pckt.MaxDetects && time.Now().Unix() > *nextscan {
					// Increment detect count.
					*detects++

					// Update scan time.
					*nextscan = time.Now().Unix() + int64(pckt.Cooldown)

					// Debug.
					if cfg.DebugLevel > 0 {
						fmt.Printf("[D1] Reporting %s:%d:%s (%s). Average latency => %dms. Max latency => %dms. Min latency => %dms. Detects => %d.\n", srv.IP, srv.Port, srv.UID, srv.Name, *avglatency, *maxlatency, *minlatency, *detects)
					}

					events.OnDetect(cfg, srv, pckt, *avglatency, *maxlatency, *minlatency, *detects, *laststats)
				}
			} else {
				// Reset everything.
				*detects = 0
				*nextscan = 0
				*maxlatency = 0
				*minlatency = 0
			}

		case <-*destroy:
			// Close UDP connection and check.
			err := conn.Close()

			if err != nil {
				fmt.Println("[ERR] Failed to close UDP connection.")
				fmt.Println(err)
			}

			// Stop timer/ticker.
			timer.Stop()

			// Stop function.
			return
		}
	}
}

func HandleServers(cfg *config.Config, update bool) {
	stats := make(map[Tuple]Stats)

	// Retrieve current server stats before removing tickers
	for _, ticker := range tickers {
		for sid, srv := range cfg.Servers {
			// Create tuple.
			var srvt Tuple
			srvt.IP = srv.IP
			srvt.Port = int(srv.Port)
			srvt.UID = srv.UID

			for pid, _ := range cfg.Servers[sid].Packets {
				srvt.PcktID = pid

				if srvt == ticker.Srv {
					if cfg.DebugLevel > 3 {
						fmt.Println("[D4] HandleServers :: Found match on " + srvt.IP + ":" + strconv.Itoa(srvt.Port) + ":" + srvt.UID + ".")
					}

					// Fill in stats.
					stats[srvt] = Stats{
						LastStats:  ticker.Stats.LastStats,
						AvgLatency: ticker.Stats.AvgLatency,
						MaxLatency: ticker.Stats.MaxLatency,
						MinLatency: ticker.Stats.MinLatency,
						Detects:    ticker.Stats.Detects,
						NextScan:   ticker.Stats.NextScan,
					}

				}
			}
		}

		// Destroy ticker.
		*ticker.Destroyer <- true
	}

	// Remove servers that should be deleted.
	for i, srv := range cfg.Servers {
		if srv.Delete {
			if cfg.DebugLevel > 1 {
				fmt.Println("[D2] Found server that should be deleted UID => " + srv.UID + ". Name => " + srv.Name + ". IP => " + srv.IP + ". Port => " + strconv.Itoa(int(srv.Port)) + ".")
			}

			RemoveServer(cfg, i)
		}
	}

	tickers = []TickerHolder{}

	// Loop through each container from the config.
	for i, srv := range cfg.Servers {
		// If we're not enabled, ignore.
		if !srv.Enable {
			continue
		}

		// Set defaults.
		if srv.Threshold < 1 {
			cfg.Servers[i].Threshold = cfg.DefThreshold
		}

		if srv.Count < 1 {
			cfg.Servers[i].Count = cfg.DefCount
		}

		if srv.Interval < 1 {
			cfg.Servers[i].Interval = cfg.DefInterval
		}

		if srv.Timeout < 1 {
			cfg.Servers[i].Timeout = cfg.DefTimeout
		}

		if srv.MaxDetects < 1 {
			cfg.Servers[i].MaxDetects = cfg.DefMaxDetects
		}

		if srv.Cooldown < 1 {
			cfg.Servers[i].Cooldown = cfg.DefCooldown
		}

		if len(srv.Mentions) < 1 {
			cfg.Servers[i].Mentions = cfg.DefMentions
		}

		// Create tuple.
		var srvt Tuple
		srvt.IP = srv.IP
		srvt.Port = int(srv.Port)
		srvt.UID = srv.UID

		for pid, pckt := range cfg.Servers[i].Packets {
			// Set defaults.
			if pckt.Threshold < 1 {
				cfg.Servers[i].Packets[pid].Threshold = cfg.Servers[i].Threshold
			}

			if pckt.Count < 1 {
				cfg.Servers[i].Packets[pid].Count = cfg.Servers[i].Count
			}

			if pckt.Interval < 1 {
				cfg.Servers[i].Packets[pid].Interval = cfg.Servers[i].Interval
			}

			if pckt.Timeout < 1 {
				cfg.Servers[i].Packets[pid].Timeout = cfg.Servers[i].Timeout
			}

			if pckt.MaxDetects < 1 {
				cfg.Servers[i].Packets[pid].MaxDetects = cfg.Servers[i].MaxDetects
			}

			if pckt.Cooldown < 1 {
				cfg.Servers[i].Packets[pid].Cooldown = cfg.Servers[i].Cooldown
			}

			// Reassign.
			pckt = cfg.Servers[i].Packets[pid]

			// Specify packet-specific variables.
			var laststats []uint32
			var avglatency uint32 = 0
			var maxlatency uint32 = 0
			var minlatency uint32 = 0
			var detects uint = 0
			var nextscan int64 = 0

			// Replace stats with old ticker's stats.
			if stat, ok := stats[srvt]; ok {
				laststats = *stat.LastStats
				avglatency = *stat.AvgLatency
				maxlatency = *stat.MaxLatency
				minlatency = *stat.MinLatency
				detects = *stat.Detects
				nextscan = *stat.NextScan
			}

			if cfg.DebugLevel > 0 && !update {
				fmt.Printf("[D1] Adding packet %s:%d:%s:%d (%s). Threshold => %d. Count => %d. Interval => %d. Timeout => %d. Request data => % x.\n", srv.IP, srv.Port, srv.UID, pid, srv.Name, pckt.Threshold, pckt.Count, pckt.Interval, pckt.Timeout, pckt.Request)
			}

			// Let's create the connection now.
			conn, err := query.CreateConnection(srv.IP, int(srv.Port))

			if err != nil {
				fmt.Println("Error creating UDP connection for " + srv.IP + ":" + strconv.Itoa(int(srv.Port)) + " ( " + srv.Name + ").")
				fmt.Println(err)

				continue
			}

			if cfg.DebugLevel > 3 {
				fmt.Printf("[D4] Creating packet timer for %s:%d:%s:%d (%s).\n", srv.IP, srv.Port, srv.UID, pid, srv.Name)
			}

			// Create destroyer channel.
			destroyer := make(chan bool)

			// Create repeating timer.
			ticker := time.NewTicker(time.Duration(pckt.Interval) * time.Second)
			go ServerWatch(&cfg.Servers[i], &cfg.Servers[i].Packets[pid], ticker, &laststats, &avglatency, &maxlatency, &minlatency, &detects, &nextscan, conn, cfg, &destroyer)

			// Add ticker to global list.
			var newticker TickerHolder
			newticker.Srv = srvt
			newticker.Ticker = ticker
			newticker.Conn = conn
			newticker.Destroyer = &destroyer
			newticker.Stats.LastStats = &laststats
			newticker.Stats.AvgLatency = &avglatency
			newticker.Stats.MaxLatency = &maxlatency
			newticker.Stats.MinLatency = &minlatency
			newticker.Stats.Detects = &detects
			newticker.Stats.NextScan = &nextscan

			tickers = append(tickers, newticker)
		}
	}
}
