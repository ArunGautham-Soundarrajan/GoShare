package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/ArunGautham-Soundarrajan/goshare/p2p"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	switch os.Args[1] {
	case "send":
		tcpHost, err := p2p.NewHost(os.Args[2])
		if err != nil {
			os.Exit(1)
		}
		if err := tcpHost.StartSever(ctx); err != nil {
			fmt.Errorf("server exited with error %w", err)
		}
	case "receive":
		tcpClient := p2p.NewTCPClient(os.Args[2])
		if err := tcpClient.DialAndConnect(ctx); err != nil {
			fmt.Errorf("stopped receiving file with %w", err)
		}
	default:
		fmt.Println("expected either 'send' or 'receive'")
		os.Exit(1)
	}
}
