package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func main() {
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
		go handleConnection(conn)
	}
}

type Metadata struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Ticket  string `json:"ticket,omitempty"`
}

func handleHandshake(reader *bufio.Reader) (*Metadata, error) {

	var metadata Metadata
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&metadata)
	if err != nil {
		return &metadata, err
	}
	return &metadata, nil
}

func handleConnection(c net.Conn) {
	defer c.Close()
	buffer := bufio.NewReader(c)

	fmt.Println("Connection received from ", c.RemoteAddr())
	metadata, err := handleHandshake(buffer)
	if err != nil {
		fmt.Println("Invalid Handshake")
		return
	}

	fmt.Println("Handshake received. Ticket:", metadata.Ticket)

	for {
		msg, err := buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client Closed Connection")
				return
			}
			fmt.Println("Read error:", err)
		}
		fmt.Println("Received: ", msg)
		c.Write([]byte("ACK: " + msg))
	}

}
