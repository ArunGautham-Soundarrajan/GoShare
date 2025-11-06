package main

import (
	"bufio"
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

func handleConnection(c net.Conn) {
	defer c.Close()
	buffer := bufio.NewReader(c)

	fmt.Println("Connection received from ", c.RemoteAddr())

	msg, err := buffer.ReadString('\n')
	if err != nil {
		fmt.Println("Connection closed or err", err)
		return
	}

	if msg != "HELLO\n" {
		fmt.Println("Invaild Handshake from", c.RemoteAddr(), "msg :", msg)
		c.Write([]byte("INVALID HANDSHAKE\n"))
		return
	}

	c.Write([]byte("WELCOME\n"))
	fmt.Println("Handshake successful with", c.RemoteAddr())

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
