package p2p

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/ArunGautham-Soundarrajan/goshare/handshake"
)

type TCPClient struct {
	ticket string

	// Extract and store these once the conn is established
	filepath string
	address  string
}

// Constructor functions to create a new TCP Client
// Takes in a ticket which is required to receive the file
func NewTCPClient(ticket string) *TCPClient {
	return &TCPClient{
		ticket: ticket,
	}
}

// Starting point for the TCP Client,
// Identifies the address to connect using the ticket
// Performs handshake, gets the file details
// Finally stores the file
func (t *TCPClient) DialAndConnect() error {

	// determine the address

	conn, err := net.Dial("tcp", t.address)
	if err != nil {
		return fmt.Errorf("error Dialing to the server %w", err)
	}

	err = t.ClientHandshake(conn)
	if err != nil {
		return err
	}

	err = ReceiveFile(conn, t.filepath)
	if err != nil {
		return err
	}

	return nil

}

// Function to perform the handshake with the server
func (t *TCPClient) ClientHandshake(c net.Conn) error {

	// Payload containing the ticket
	payload := handshake.RequestPayload{
		Type:    "Handshake",
		Version: "1.0",
		Ticket:  t.ticket,
	}
	data, _ := json.Marshal(payload)
	err := handshake.WriteFrame(c, data)
	if err != nil {
		return fmt.Errorf("failed to send handshake request: %w", err)
	}

	// Response from the server (ACK)
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

func ReceiveFile(c net.Conn, filePath string) error {

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating the file %w", err)
	}
	defer file.Close()

	var data handshake.FileData

	for {

		err := handshake.ReadFrame(c, &data)

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading data frame: %w", err)
		}
		if data.Type == "EOF" {
			fmt.Println("Transfer finished by EOF signal.")
			break
		}

		chunk, err := base64.StdEncoding.DecodeString(data.Data)
		if err != nil {
			return fmt.Errorf("error decoding base64 chunk : %w", err)
		}

		_, err = file.Write(chunk)
		if err != nil {
			return fmt.Errorf("failed to write data to file: %w", err)
		}

	}

	return nil
}
