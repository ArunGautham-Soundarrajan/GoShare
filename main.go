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
		tcpHost, err := p2p.NewHost(os.Args[2])
		if err != nil {
			os.Exit(1)
		}
		log.Fatal(tcpHost.StartSever())
	case "receive":
		tcpClient := p2p.NewTCPClient(os.Args[2])
		log.Fatal(tcpClient.DialAndConnect())
	default:
		fmt.Println("expected either 'send' or 'receive'")
		os.Exit(1)
	}
}
