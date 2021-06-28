package misc

import (
	"strconv"
	"strings"

	"github.com/gamemann/Pterodactyl-Packet-Watch/config"
)

func FormatContents(app string, formatstr *string, avglatency uint32, maxlatency uint32, minlatency uint32, detects uint, srv *config.Server, pckt *config.Packet, mentionstr string) {
	*formatstr = strings.ReplaceAll(*formatstr, "{IP}", srv.IP)
	*formatstr = strings.ReplaceAll(*formatstr, "{PORT}", strconv.FormatUint(uint64(srv.Port), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{UID}", srv.UID)
	*formatstr = strings.ReplaceAll(*formatstr, "{NAME}", srv.Name)
	*formatstr = strings.ReplaceAll(*formatstr, "{PCKTNAME}", pckt.Name)
	*formatstr = strings.ReplaceAll(*formatstr, "{THRESHOLD}", strconv.FormatUint(uint64(pckt.Threshold), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{COUNT}", strconv.FormatUint(uint64(pckt.Count), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{INTERVAL}", strconv.FormatUint(uint64(pckt.Interval), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{TIMEOUT}", strconv.FormatUint(uint64(pckt.Timeout), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{MAXDETECTS}", strconv.FormatUint(uint64(pckt.MaxDetects), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{COOLDOWN}", strconv.FormatUint(uint64(pckt.Cooldown), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{AVG}", strconv.FormatUint(uint64(avglatency), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{MAX}", strconv.FormatUint(uint64(maxlatency), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{MIN}", strconv.FormatUint(uint64(minlatency), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{DETECTS}", strconv.FormatUint(uint64(detects), 10))
	*formatstr = strings.ReplaceAll(*formatstr, "{MENTIONS}", mentionstr)
}
