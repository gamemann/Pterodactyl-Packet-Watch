package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Packet types.
type Packet struct {
	Name       string `json:"name"`
	Request    []byte `json:"data"`
	Interval   uint   `json:"interval"`
	Threshold  uint32 `json:"threshold"`
	Count      uint   `json:"count"`
	Timeout    uint   `json:"timeout"`
	MaxDetects uint   `json:"maxdetects"`
	Cooldown   uint   `json:"cooldown"`
}

// Server struct used for each server config.
type Server struct {
	Name       string   `json:"name"`
	Enable     bool     `json:"enable"`
	IP         string   `json:"ip"`
	Port       uint     `json:"port"`
	UID        string   `json:"uid"`
	Interval   uint     `json:"interval"`
	Threshold  uint32   `json:"threshold"`
	Count      uint     `json:"count"`
	Timeout    uint     `json:"timeout"`
	MaxDetects uint     `json:"maxdetects"`
	Cooldown   uint     `json:"cooldown"`
	Packets    []Packet `json:"packets"`
	Mentions   string   `json:"mentions"`
	ViaAPI     bool
	Delete     bool
}

// Misc options.
type Misc struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Config struct used for the general config.
type Config struct {
	APIURL        string   `json:"apiurl"`
	Token         string   `json:"token"`
	AppToken      string   `json:"apptoken"`
	AddServers    bool     `json:"addservers"`
	DebugLevel    uint     `json:"debug"`
	ReloadTime    uint     `json:"reloadtime"`
	DefEnable     bool     `json:"defenable"`
	DefThreshold  uint32   `json:"defthreshold"`
	DefCount      uint     `json:"defcount"`
	DefInterval   uint     `json:"definterval"`
	DefTimeout    uint     `json:"deftimeout"`
	DefMaxDetects uint     `json:"defmaxdetects"`
	DefCooldown   uint     `json:"defcooldown"`
	DefMentions   string   `json:"defmentions"`
	Servers       []Server `json:"servers"`
	Misc          []Misc   `json:"misc"`
	ConfLoc       string
}

// Reads a config file based off of the file name (string) and returns a Config struct.
func (cfg *Config) ReadConfig(filename string) bool {
	file, err := os.Open(filename)

	if err != nil {
		fmt.Println("[ERR] Cannot open config file.")
		fmt.Println(err)

		return false
	}

	defer file.Close()

	stat, _ := file.Stat()

	data := make([]byte, stat.Size())

	_, err = file.Read(data)

	if err != nil {
		fmt.Println("[ERR] Cannot read config file.")
		fmt.Println(err)

		return false
	}

	err = json.Unmarshal([]byte(data), cfg)

	if err != nil {
		fmt.Println("[ERR] Cannot parse JSON Data.")
		fmt.Println(err)

		return false
	}

	return true
}

// Sets config's default values.
func (cfg *Config) SetDefaults() {
	// Set config defaults.
	cfg.AddServers = false
	cfg.DebugLevel = 0
	cfg.ReloadTime = 500

	cfg.DefEnable = true
	cfg.DefThreshold = 60
	cfg.DefCount = 10
	cfg.DefInterval = 5
	cfg.DefTimeout = 5
	cfg.DefMaxDetects = 2
	cfg.DefCooldown = 120
}
