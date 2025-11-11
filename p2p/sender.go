package p2p

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/ArunGautham-Soundarrajan/goshare/handshake"
)

type Peer struct {
	Conn net.Conn
	Addr string
}

type TCPHost struct {
	ListenAddr string
	ticket     string
	filepath   string
	peers      map[string]*Peer // clientaddr : net.conn
	mu         sync.RWMutex
}

// Constructor for new TCP host
func NewHost(listenAddr string, filepath string) *TCPHost {
	return &TCPHost{
		ListenAddr: listenAddr,
		filepath:   filepath,
		peers:      make(map[string]*Peer),
	}
}

// Get the file name, and generte a ticket which would let the client
// Identify the server and dial to it to request files
func (t *TCPHost) GenerateTicket() error {

	// Placeholder while we implement Ticket logic
	t.ticket = "test"
	return nil
}

// Start the Server and listen for incoming connections
// Upon receiving connections, handle the connections concurrently
func (t *TCPHost) StartSever() error {

	// Generate the ticket
	t.GenerateTicket()

	// Start listening for peers
	listener, err := net.Listen("tcp", t.ListenAddr)
	fmt.Println("Lisenting on localhost:8080")
	if err != nil {
		fmt.Println(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go t.handleConnection(conn)
	}
}

// Core logic to handle incoming connections
// Perform a handshake to verify the client
// Append to the list of peers
// Transfer the file
func (t *TCPHost) handleConnection(c net.Conn) error {

	defer c.Close()

	// Perform the handshake with the client
	err := t.SeverHandshake(c)
	if err != nil {
		return err
	}

	// Add the peer to the list of peers
	t.mu.Lock()
	t.peers[c.RemoteAddr().String()] = &Peer{
		Conn: c,
		Addr: c.RemoteAddr().String(),
	}
	t.mu.Unlock()

	// Stream the file
	err = StreamFile(c, t.filepath)
	if err != nil {
		return err
	}

	return nil
}

// Perform handshake with the client
// This involves, verifying if the ticket is valid and acknowleding it
func (t *TCPHost) SeverHandshake(c net.Conn) error {

	var payload handshake.RequestPayload

	err := handshake.ReadFrame(c, &payload)
	if err != nil {
		return fmt.Errorf("failed to read receiver ticket frame: %w", err)
	}

	err = handshake.VerifyTicket(t.ticket, payload.Ticket)
	if err != nil {
		failureJSON, _ := json.Marshal(handshake.Response{Status: "failure"})
		if err := handshake.WriteFrame(c, failureJSON); err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}
		return fmt.Errorf("receiver ticket verification failed: %w", err)
	}

	successJSON, _ := json.Marshal(handshake.Response{Status: "success"})
	if err := handshake.WriteFrame(c, successJSON); err != nil {
		return fmt.Errorf("failed to send success response: %w", err)
	}
	return nil
}

// This functions streams the file to the client
// TODO: Refactor and make it concurrent
func StreamFile(c net.Conn, filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening the file %w", err)
	}
	defer file.Close()

	fileBuffer := make([]byte, 1024*1024) //1MB chunks

	for {
		n, readErr := file.Read(fileBuffer)
		if readErr == io.EOF {
			break // File finished
		}
		if readErr != nil {
			return fmt.Errorf("error reading file: %w", readErr)
		}
		if n > 0 {
			chunk := fileBuffer[:n]

			encodedChunk := base64.StdEncoding.EncodeToString(chunk)

			dataFrame := handshake.FileData{
				Type: "CHUNK",
				Data: encodedChunk,
			}

			data, _ := json.Marshal(dataFrame)
			err = handshake.WriteFrame(c, data)
			if err != nil {
				return fmt.Errorf("failed to stream chunk: %w", err)
			}
		}
	}

	eofFrame, _ := json.Marshal(struct{ Type string }{"EOF"})
	return handshake.WriteFrame(c, eofFrame)
}
