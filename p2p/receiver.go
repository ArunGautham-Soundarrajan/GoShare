package p2p

import (
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
	fileName string
	fileSize int64
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
	// TODO: Once ticketing system is ready, derive address from that
	t.address = ":8080"

	conn, err := net.Dial("tcp", t.address)
	if err != nil {
		return fmt.Errorf("error Dialing to the server %w", err)
	}

	defer conn.Close()

	err = t.ClientHandshake(conn)
	if err != nil {
		return err
	}

	// receive the file info
	if err = t.ReceiveFileInfo(conn); err != nil {
		return err
	}

	if t.fileName != "" {
		err = ReceiveFile(conn, t.fileName)
		if err != nil {
			return err
		}
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

// function to receive the file info from the server
func (t *TCPClient) ReceiveFileInfo(c net.Conn) error {
	var fileInfo handshake.FileInfoPayload

	if err := handshake.ReadFrame(c, &fileInfo); err != nil {
		return fmt.Errorf("error while reading file info %w", err)
	}

	if fileInfo.Type == "FILE_INFO" {
		t.fileName = fileInfo.FileName
		t.fileSize = fileInfo.Size

		fmt.Println(fileInfo)
	}

	return nil
}

func ReceiveFile(c net.Conn, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating the file %w", err)
	}
	defer file.Close()

	for {

		chunk, err := handshake.ReadRawFrame(c)

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading data frame: %w", err)
		}

		_, err = file.Write(chunk)
		if err != nil {
			return fmt.Errorf("failed to write data to file: %w", err)
		}

	}
	fmt.Println("Successfully received file")
	return nil
}
