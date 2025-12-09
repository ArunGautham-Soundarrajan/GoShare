# GoShare — P2P File Sharing

## Project overview

GoShare is a planned peer-to-peer (P2P) file sharing system written in Go. The goal of this project is to design and implement a simple, secure, and efficient P2P protocol and reference implementation that enables peers to discover one another and exchange files directly without relying on a single centralized server. The implementation is minimal and intended for learning / experimentation.

## Build

From the repository root:

- Build:
  go build -o goshare

- Or run directly:
  go run main.go ...

## Usage

The program expects exactly two arguments: a command and file/ticket.

Commands:
- send <filename>
  - Starts a host (server) and waits for clinet to send the file.
- receive <peer-addr>
  - Dial/connect to a remote peer at the given address.

If the command is missing or not one of `send`/`receive`, the program prints:
expected either 'send' or 'receive'

Examples:
- Start a sender (listening on random TCP port):
  ```
  ./goshare send ./example.txt
  ```
  or
  ```
  go run main.go send ./example.txt
  ```

- Connect as a receiver to the sender:
  ```
  ./goshare receive /ip4/127.0.0.1/tcp/50114/p2p/12D3KooWAJLW3
  ```
  or
  ```
  go run main.go receive /ip4/127.0.0.1/tcp/50114/p2p/12D3KooWAJLW3
  ```
