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

func ServerHandshake(c net.Conn, ticket string) error {

	var payload handshake.RequestPayload

	err := handshake.ReadFrame(c, &payload)
	if err != nil {
		return fmt.Errorf("failed to read receiver ticket frame: %w", err)
	}

	err = handshake.VerifyTicket(ticket, payload.Ticket)
	if err != nil {
		failureJSON, _ := json.Marshal(handshake.Response{Status: "failure"})
		if err := handshake.WriteFrame(c, failureJSON); err != nil {
			c.Close()
			return fmt.Errorf("failed to send response: %w", err)
		}
		c.Close()
		return fmt.Errorf("receiver ticket verification failed: %w", err)
	}

	successJSON, _ := json.Marshal(handshake.Response{Status: "success"})
	if err := handshake.WriteFrame(c, successJSON); err != nil {
		c.Close()
		return fmt.Errorf("failed to send success response: %w", err)
	}

	err = StreamFile(c, "/Users/arun_s/Workspace/go_projects/GoShare/README.md")
	if err != nil {
		return err
	}

	return nil
}

func StartSever() error {
	listener, err := net.Listen("tcp", ":8080")
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
		go ServerHandshake(conn, "test")
	}
}

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
