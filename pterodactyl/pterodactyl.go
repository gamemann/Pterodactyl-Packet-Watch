package pterodactyl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gamemann/Pterodactyl-Packet-Watch/config"
)

// Attributes struct from /api/client/servers/xxxx/resources.
type Attributes struct {
	State string `json:"current_state"`
}

// Utilization struct from /api/client/servers/xxxx/resources.
type Utilization struct {
	Attributes Attributes `json:"attributes"`
}

// Retrieves all servers/containers from Pterodactyl API and add them to the config.
func AddServers(cfg *config.Config) bool {
	// Retrieve max page count.
	pagecount := 1
	maxpages := 1
	total := 0
	done := false

	for done != true {
		// Build endpoint.
		urlstr := cfg.APIURL + "/api/application/servers?page=" + strconv.Itoa(pagecount) + "&include=allocations,variables"

		// Setup HTTP GET request.
		client := &http.Client{Timeout: time.Second * 5}
		req, err := http.NewRequest("GET", urlstr, nil)

		if err != nil {
			fmt.Println(err)

			return false
		}

		// Set Application API token.
		req.Header.Set("Authorization", "Bearer "+cfg.AppToken)

		// Accept only JSON.
		req.Header.Set("Accept", "application/json")

		// Perform HTTP request and check for errors.
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println(err)

			return false
		}

		// Read body.
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Println(err)

			return false
		}

		// Create data interface.
		var dataobj interface{}

		// Parse JSON.
		err = json.Unmarshal([]byte(string(body)), &dataobj)

		if err != nil {
			fmt.Println(err)

			return false
		}

		// Look for object item before anything.
		if dataobj.(map[string]interface{})["object"] == nil {
			fmt.Println("[ERR] 'object' item not found when listing all servers.")

			return false
		}

		// Retrieve max page count and total count.
		maxpages = int(dataobj.(map[string]interface{})["meta"].(map[string]interface{})["pagination"].(map[string]interface{})["total_pages"].(float64))
		total = int(dataobj.(map[string]interface{})["meta"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64))

		resp.Body.Close()

		// Loop through each data item (server).
		for z, j := range dataobj.(map[string]interface{})["data"].([]interface{}) {
			item := j.(map[string]interface{})

			// Make sure we have a server object.
			if item["object"] == "server" {
				attr := item["attributes"].(map[string]interface{})

				// Build new server structure.
				var sta config.Server

				// Set UID (in this case, identifier) and default values.
				sta.ViaAPI = true
				sta.UID = attr["identifier"].(string)
				sta.Name = attr["name"].(string)

				sta.Enable = cfg.DefEnable
				sta.Threshold = cfg.DefThreshold
				sta.Count = cfg.DefCount
				sta.Interval = cfg.DefInterval
				sta.Mentions = cfg.DefMentions
				sta.Timeout = cfg.DefTimeout

				// Retrieve default IP/port.
				for _, i := range attr["relationships"].(map[string]interface{})["allocations"].(map[string]interface{})["data"].([]interface{}) {
					if i.(map[string]interface{})["object"].(string) != "allocation" {
						continue
					}

					alloc := i.(map[string]interface{})["attributes"].(map[string]interface{})

					if alloc["assigned"].(bool) {
						sta.IP = alloc["ip"].(string)
						sta.Port = uint(alloc["port"].(float64))
					}
				}

				// Look for overrides.
				if attr["relationships"].(map[string]interface{})["variables"].(map[string]interface{})["data"] != nil {
					for _, i := range attr["relationships"].(map[string]interface{})["variables"].(map[string]interface{})["data"].([]interface{}) {
						if i.(map[string]interface{})["object"].(string) != "server_variable" {
							continue
						}

						vari := i.(map[string]interface{})["attributes"].(map[string]interface{})

						// Check if we have a value.
						if vari["server_value"] == nil {
							continue
						}

						val := vari["server_value"].(string)

						// Override variables should always be at least one byte in length.
						if len(val) < 1 {
							continue
						}

						// Check for IP override.
						if vari["env_variable"].(string) == "PTEROPCKT_IP" {
							sta.IP = val
						}

						// Check for port override.
						if vari["env_variable"].(string) == "PTEROPCKT_PORT" {
							tmp, err := strconv.ParseUint(val, 10, 16)

							if err == nil {
								sta.Port = uint(tmp)
							}
						}

						// Check for threshold override.
						if vari["env_variable"].(string) == "PTEROPCKT_THRESHOLD" {
							tmp, _ := strconv.ParseUint(val, 10, 16)

							if err == nil {
								sta.Threshold = uint32(tmp)
							}
						}

						// Check for count override.
						if vari["env_variable"].(string) == "PTEROPCKT_COUNT" {
							tmp, _ := strconv.ParseUint(val, 10, 16)

							if err == nil {
								sta.Count = uint(tmp)
							}
						}

						// Check for interval override.
						if vari["env_variable"].(string) == "PTEROPCKT_INTERVAL" {
							tmp, _ := strconv.ParseUint(val, 10, 16)

							if err == nil {
								sta.Interval = uint(tmp)
							}
						}

						// Check for timeout override.
						if vari["env_variable"].(string) == "PTEROPCKT_TIMEOUT" {
							tmp, _ := strconv.ParseUint(val, 10, 16)

							if err == nil {
								sta.Timeout = uint(tmp)
							}
						}

						// Check for packets override.
						if vari["env_variable"].(string) == "PTEROPCKT_PACKETS" {
							err := json.Unmarshal([]byte(val), &cfg.Servers[z].Packets)

							if err != nil {
								fmt.Printf("Error parsing packets JSON for server %s (%s:%d:%s).\nError => %v", sta.Name, sta.IP, sta.Port, sta.UID, err)
							}
						}

						// Check for mentions override.
						if vari["env_variable"].(string) == "PTEROPCKT_MENTIONS" {
							sta.Mentions = val
						}

						// Check for disable override.
						if vari["env_variable"].(string) == "PTEROPCKT_DISABLE" {
							disable, _ := strconv.Atoi(val)

							if disable > 0 {
								sta.Enable = false
							} else {
								sta.Enable = true
							}
						}
					}
				}

				// Append to servers slice.
				cfg.Servers = append(cfg.Servers, sta)
			}
		}

		// Check page count.
		if pagecount >= maxpages {
			done = true

			break
		}

		pagecount++
	}

	// Level 2 debug.
	if cfg.DebugLevel > 1 {
		fmt.Println("[D2] Found " + strconv.Itoa(total) + " servers from API (" + strconv.Itoa(maxpages) + " page(s)).")
	}

	return true
}

// Checks the status of a Pterodactyl server. Returns true if on and false if off.
// DOES NOT INCLUDE IN "STARTING" MODE.
func CheckStatus(cfg *config.Config, uid string) bool {
	// Build endpoint.
	urlstr := cfg.APIURL + "/" + "api/client/servers/" + uid + "/resources"

	// Setup HTTP GET request.
	client := &http.Client{Timeout: time.Second * 5}
	req, _ := http.NewRequest("GET", urlstr, nil)

	// Set authorization header.
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	// Set data to JSON.
	req.Header.Set("Content-Type", "application/json")

	// Accept JSON.
	req.Header.Set("Accept", "application/json")

	// Perform HTTP request and check for errors.
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)

		return false
	}

	// Close body at the end.
	defer resp.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)

		return false
	}

	// Create utilization struct.
	var util Utilization

	// Parse JSON.
	json.Unmarshal([]byte(string(body)), &util)

	// Check if the server's state isn't on. If not, return false.
	if util.Attributes.State != "running" {
		return false
	}

	// Otherwise, return true meaning the container is online.
	return true
}
