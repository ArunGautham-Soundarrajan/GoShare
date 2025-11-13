package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ArunGautham-Soundarrajan/goshare/p2p"
)

func main() {
	switch os.Args[1] {
	case "send":
		tcpHost := p2p.NewHost(":8080", "./README.md")
		log.Fatal(tcpHost.StartSever())
	case "receive":
		tcpClient := p2p.NewTCPClient("test")
		log.Fatal(tcpClient.DialAndConnect())
	default:
		fmt.Println("expected either 'send' or 'receive'")
		os.Exit(1)
	}
}
