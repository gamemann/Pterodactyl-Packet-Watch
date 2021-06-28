package query

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Creates a UDP connection using the host and port.
func CreateConnection(host string, port int) (*net.UDPConn, error) {
	var UDPC *net.UDPConn

	// Combine host and port.
	fullHost := host + ":" + strconv.Itoa(port)

	UDPAddr, err := net.ResolveUDPAddr("udp", fullHost)

	if err != nil {
		return UDPC, err
	}

	// Attempt to open a UDP connection.
	UDPC, err = net.DialUDP("udp", nil, UDPAddr)

	if err != nil {
		fmt.Println(err)

		return UDPC, errors.New("NoConnection")
	}

	return UDPC, nil
}

// Sends an A2S_INFO request to the host and port specified in the arguments.
func SendRequest(conn *net.UDPConn, data []byte) {
	conn.Write(data)
}

// Checks for A2S_INFO response. Returns true if it receives a response. Returns false otherwise.
func CheckResponse(conn *net.UDPConn, timeout uint) bool {
	buffer := make([]byte, 1024)

	// Set read timeout.
	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(timeout)))

	_, _, err := conn.ReadFromUDP(buffer)

	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}
