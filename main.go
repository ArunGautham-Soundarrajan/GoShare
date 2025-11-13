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
		tcpHost, err := p2p.NewHost(":8080", "./README.md")
		if err != nil {
			os.Exit(1)
		}
		log.Fatal(tcpHost.StartSever())
	case "receive":
		tcpClient := p2p.NewTCPClient("test")
		log.Fatal(tcpClient.DialAndConnect())
	default:
		fmt.Println("expected either 'send' or 'receive'")
		os.Exit(1)
	}
}
