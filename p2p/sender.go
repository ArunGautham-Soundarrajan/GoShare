package p2p

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ArunGautham-Soundarrajan/goshare/handshake"
)

func performHandshake(c net.Conn, ticket string) error {

	var payload handshake.RequestPayload

	err := handshake.ReadFrame(c, &payload)
	if err != nil {
		return fmt.Errorf("failed to read receiver ticket frame: %w", err)
	}

	err = handshake.VerifyClient(ticket, payload.Ticket)
	if err != nil {
		failureJSON, _ := json.Marshal(handshake.Response{Status: "failure"})
		handshake.WriteFrame(c, failureJSON)
		c.Close()
		return fmt.Errorf("receiver ticket verification failed: %w", err)
	}

	successJSON, _ := json.Marshal(handshake.Response{Status: "success"})
	if err := handshake.WriteFrame(c, successJSON); err != nil {
		c.Close()
		return fmt.Errorf("failed to send success response: %w", err)
	}
	return nil
}
