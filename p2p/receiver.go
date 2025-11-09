package p2p

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ArunGautham-Soundarrajan/goshare/handshake"
)

func ClientHandshake(c net.Conn, ticket string) error {

	payload := handshake.RequestPayload{
		Type:    "Handshake",
		Version: "1.0",
		Ticket:  ticket,
	}
	data, _ := json.Marshal(payload)
	err := handshake.WriteFrame(c, data)
	if err != nil {
		return fmt.Errorf("failed to send handshake request: %w", err)
	}

	var resp handshake.Response
	err = handshake.ReadFrame(c, &resp)
	if err != nil {
		return fmt.Errorf("failed to read handshake response: %w", err)
	}

	if resp.Status != "success" {
		return fmt.Errorf("server rejected handshake")
	}

	return nil
}

func StartClient() error {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		return fmt.Errorf("error Dialing to the server %w", err)
	}

	ClientHandshake(conn, "test")

	return nil
}
