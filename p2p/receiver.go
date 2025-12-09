package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ArunGautham-Soundarrajan/goshare/handshake"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/schollz/progressbar/v3"
)

type TCPClient struct {
	ticket   string
	client   host.Host // Extract and store these once the conn is established
	fileName string
	fileSize int64
	address  *peer.AddrInfo
}

// Constructor functions to create a new TCP Client
// Takes in a ticket which is required to receive the file
func NewTCPClient(ticket string) *TCPClient {
	host, err := libp2p.New()
	if err != nil {
		return nil
	}

	return &TCPClient{
		ticket: ticket,
		client: host,
	}
}

// Starting point for the TCP Client,
// Identifies the address to connect using the ticket
// Performs handshake, gets the file details
// Finally stores the file
func (t *TCPClient) DialAndConnect() error {
	// determine the address
	var err error
	t.address, err = peer.AddrInfoFromString(t.ticket)
	if err != nil {
		return fmt.Errorf("failed to parse senders address %w", err)
	}

	err = t.client.Connect(context.TODO(), *t.address)
	if err != nil {
		return fmt.Errorf("error Dialing to the server %w", err)
	}

	stream, err := t.client.NewStream(context.TODO(), t.address.ID, "/goshare")
	if err != nil {
		return fmt.Errorf("failed to create a steam %w", err)
	}

	err = t.ClientHandshake(stream)
	if err != nil {
		return err
	}

	// receive the file info
	if err = t.ReceiveFileInfo(stream); err != nil {
		return err
	}

	if t.fileName != "" {
		err = t.ReceiveFile(stream)
		if err != nil {
			return err
		}
	}
	return nil
}

// Function to perform the handshake with the server
func (t *TCPClient) ClientHandshake(s network.Stream) error {
	// Payload containing the ticket
	payload := handshake.RequestPayload{
		Type:    "Handshake",
		Version: "1.0",
		Ticket:  t.ticket,
	}
	data, _ := json.Marshal(payload)
	err := handshake.WriteFrame(s, data)
	if err != nil {
		return fmt.Errorf("failed to send handshake request: %w", err)
	}

	// Response from the server (ACK)
	var resp handshake.Response
	err = handshake.ReadFrame(s, &resp)
	if err != nil {
		return fmt.Errorf("failed to read handshake response: %w", err)
	}

	if resp.Status != "success" {
		return fmt.Errorf("server rejected handshake")
	}

	return nil
}

// function to receive the file info from the server
func (t *TCPClient) ReceiveFileInfo(s network.Stream) error {
	var fileInfo handshake.FileInfoPayload

	if err := handshake.ReadFrame(s, &fileInfo); err != nil {
		return fmt.Errorf("error while reading file info %w", err)
	}

	if fileInfo.Type == "FILE_INFO" {
		t.fileName = fileInfo.FileName
		t.fileSize = fileInfo.Size

		fmt.Println("Receiving ", fileInfo.FileName)
	}

	return nil
}

func (t *TCPClient) ReceiveFile(s network.Stream) error {
	file, err := os.Create(t.fileName)
	if err != nil {
		return fmt.Errorf("error creating the file %w", err)
	}
	defer file.Close()

	// progressbar
	bar := progressbar.DefaultBytes(t.fileSize, "Receiving file")
	_, err = io.Copy(io.MultiWriter(file, bar), s)
	if err != nil {
		return fmt.Errorf("error downloading the file %w", err)
	}
	return nil
}
