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

	err = ClientHandshake(conn, "test")
	if err != nil {
		return err
	}

	err = ReceiveFile(conn, "READMEEEE.md")
	if err != nil {
		return err
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
