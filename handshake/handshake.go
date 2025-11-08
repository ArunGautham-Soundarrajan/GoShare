package handshake

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// MaxFrameSize defines the maximum allowed size (in bytes) for a single incoming
// payload frame. This limit prevents Denial-of-Service (DoS) attacks by
// pre-allocating excessively large buffers and protects against running out
// of memory (OOM).
// The current limit is set to 10 MB.
const MaxFrameSize = 10 * 1024 * 1024

func readFrame(r io.Reader, v interface{}) error {
	lenBuf := make([]byte, 4)
	// Read the 4-byte length prefix
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return fmt.Errorf("failed to read frame length: %w", err)
	}

	length := binary.BigEndian.Uint32(lenBuf)

	// Check for maximum size limit (Security/Stability)
	if length > MaxFrameSize {
		return fmt.Errorf("frame size %d exceeds max limit %d", length, MaxFrameSize)
	}

	// Read the payload
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return fmt.Errorf("failed to read payload of length %d: %w", length, err)
	}

	//Unmarshal the JSON
	return json.Unmarshal(payload, v)
}

// verifyClient checks if the clientTicket matches the expectedServerTicket.
// Returns an error if the tickets do not match, or nil if they do.
func verifyClient(expectedServerTicket string, clientTicket string) error {
	if clientTicket != expectedServerTicket {
		return fmt.Errorf("client ticket validation failed: ticket mismatch")
	}
	// Return nil for success
	return nil
}

func writeFrame(w io.Writer, data []byte) error {

	// calculate the length of the payload
	length := uint32(len(data))

	// return error if the length of the payload exceeds Max framesize
	if length > MaxFrameSize {
		return fmt.Errorf("response data size %d exceeds max limit %d", length, MaxFrameSize)
	}

	// Encode the lenght of the payload to 4-byte length prefix
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, length)

	// Write the length prefix
	_, err := w.Write(lenBuf)
	if err != nil {
		return fmt.Errorf("failed to write frame length %w", err)
	}

	// Write the payload data
	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write frame payload %w", err)
	}

	return nil
}
