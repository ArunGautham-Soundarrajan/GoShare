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
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/schollz/progressbar/v3"
)

type TCPHost struct {
	ticket   string
	Host     host.Host
	file     os.FileInfo
	filepath string
	done     chan struct{}
}

// Constructor for new TCP host
func NewHost(filepath string) (*TCPHost, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file Doesn't exist at the location %w", err)
		}
		return nil, err
	}
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // Listen on random port
		libp2p.EnableNATService(),
		libp2p.NATPortMap(),         // Try UPnP/PMP
		libp2p.EnableHolePunching(), // Crucial for NAT traversal
		libp2p.EnableRelay(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating a host %w", err)
	}

	return &TCPHost{
		Host:     host,
		file:     info,
		filepath: filepath,
		done:     make(chan struct{}),
	}, nil
}

// Get the file name, and generte a ticket which would let the client
// Identify the server and dial to it to request files
func (t *TCPHost) GenerateTicket() error {
	peerInfo := peer.AddrInfo{
		ID:    t.Host.ID(),
		Addrs: t.Host.Addrs(),
	}

	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return err
	}

	// Placeholder while we implement Ticket logic
	t.ticket = addrs[0].String()
	return nil
}

// Start the Server and listen for incoming connections
// Upon receiving connections, handle the connections concurrently
func (t *TCPHost) StartSever(ctx context.Context) error {
	// Generate the ticket

	protocolID := "/goshare"

	t.Host.SetStreamHandler(protocol.ID(protocolID), func(s network.Stream) { t.handleConnection(ctx, s) })

	err := t.GenerateTicket()
	if err != nil {
		return err
	}

	fmt.Println("Server is ready. Share the ticket:", t.ticket)
	<-t.done
	fmt.Println("File transferred. Server shutting down.")

	return nil
}

// Core logic to handle incoming connections
// Perform a handshake to verify the client
// Append to the list of peers
// Transfer the file
func (t *TCPHost) handleConnection(ctx context.Context,
	s network.Stream,
) {
	defer s.Close()

	select {
	case <-ctx.Done():
		fmt.Println("Connection canceled")
		return

	default:
		// Perform the handshake with the client
		err := t.SeverHandshake(s)
		if err != nil {
			fmt.Printf("Handshake failed: %v\n", err)
			return
		}

		// Send the file info
		err = t.SendFileInfo(s)
		if err != nil {
			fmt.Printf("sending fileinfo failed: %v\n", err)
			return
		}

		// Stream the file
		err = t.StreamFile(s)
		if err != nil {
			return
		}

		close(t.done)
	}
}

// Perform handshake with the client
// This involves, verifying if the ticket is valid and acknowleding it
func (t *TCPHost) SeverHandshake(rw io.ReadWriteCloser) error {
	var payload handshake.RequestPayload

	err := handshake.ReadFrame(rw, &payload)
	if err != nil {
		return fmt.Errorf("failed to read receiver ticket frame: %w", err)
	}

	err = handshake.VerifyTicket(t.ticket, payload.Ticket)
	if err != nil {
		failureJSON, _ := json.Marshal(handshake.Response{Status: "failure"})
		if err := handshake.WriteFrame(rw, failureJSON); err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}
		return fmt.Errorf("receiver ticket verification failed: %w", err)
	}

	successJSON, _ := json.Marshal(handshake.Response{Status: "success"})
	if err := handshake.WriteFrame(rw, successJSON); err != nil {
		return fmt.Errorf("failed to send success response: %w", err)
	}
	return nil
}

// Send the fileinfo to the client
func (t *TCPHost) SendFileInfo(w io.Writer) error {
	payload := handshake.FileInfoPayload{
		Type:     "FILE_INFO",
		FileName: t.file.Name(),
		Size:     t.file.Size(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal file info payload: %w", err)
	}

	if err := handshake.WriteFrame(w, data); err != nil {
		return fmt.Errorf("failed to send file info: %w", err)
	}

	return nil
}

// This functions streams the file to the client
// TODO: Refactor and make it concurrent
func (t *TCPHost) StreamFile(w io.Writer) error {
	file, err := os.Open(t.filepath)
	if err != nil {
		return fmt.Errorf("error opening the file %w", err)
	}
	defer file.Close()

	// progressbar
	bar := progressbar.DefaultBytes(t.file.Size(), "Sending")

	_, err = io.Copy(io.MultiWriter(w, bar), file)
	if err != nil {
		return fmt.Errorf("error streaming the file %w", err)
	}
	return nil
}
