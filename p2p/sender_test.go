package p2p

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"testing"

	"github.com/ArunGautham-Soundarrajan/goshare/handshake"
)

// Helper function to create a temporary file with content for testing
func createTestFile(t *testing.T, content string) (string, int64, func()) {
	tmpFile, err := os.CreateTemp("", "testfile-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Write the content
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Get file info
	info, err := tmpFile.Stat()
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	filePath := tmpFile.Name()
	fileSize := info.Size()

	// Ensure the file is closed and cleaned up after the test
	cleanup := func() {
		tmpFile.Close()
		os.Remove(filePath)
	}

	return filePath, fileSize, cleanup
}

// ====================================================================
// Test SeverHandshake
// ====================================================================

func TestSeverHandshake_Success(t *testing.T) {
	clientConn, serverConn := net.Pipe()

	// The main thread acts as the client. Defer client side closure.
	defer clientConn.Close()

	// Channel to receive any error that occurred in the server Goroutine
	serverDone := make(chan error, 1)

	expectedTicket := "VALID_TEST_TICKET"
	host := &TCPHost{ticket: expectedTicket}

	// 1. Server Action (Goroutine)
	// The server runs concurrently. It handles all server-side logic and ensures
	// serverConn is closed after it finishes writing the response.
	go func() {
		connStream := io.ReadWriteCloser(serverConn)
		err := host.SeverHandshake(connStream)
		// Crucial: The server must close its connection side after its job is done
		// (success or failure) to unblock the client's final read.
		serverConn.Close()
		serverDone <- err
	}()

	// 2. Client Action (Main Thread) - Write Request
	clientPayload := handshake.RequestPayload{Ticket: expectedTicket}
	data, err := json.Marshal(clientPayload)
	if err != nil {
		t.Fatalf("Failed to marshal client payload: %v", err)
	}
	if err := handshake.WriteFrame(clientConn, data); err != nil {
		t.Fatalf("Client failed to write request frame: %v", err)
	}

	// 3. Client Action (Main Thread) - Read Response
	// This read will block until the server writes the response and closes serverConn.
	var resp handshake.Response
	if err := handshake.ReadFrame(clientConn, &resp); err != nil {
		t.Fatalf("Failed to read server response: %v", err)
	}

	// 4. Verification
	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", resp.Status)
	}

	// 5. Wait for the server Goroutine to finish and check its result
	if err := <-serverDone; err != nil {
		t.Errorf("Server Goroutine returned an error: %v", err)
	}
}

func TestSeverHandshake_Failure(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	serverDone := make(chan error, 1)

	host := &TCPHost{ticket: "CORRECT_TICKET"}

	// 1. Server Action (Goroutine)
	go func() {
		connStream := io.ReadWriteCloser(serverConn)
		err := host.SeverHandshake(connStream)
		serverConn.Close() // Server closes its connection after it's done
		serverDone <- err
	}()

	// 2. Client Action (Main Thread) - Write Wrong Request
	clientPayload := handshake.RequestPayload{Ticket: "WRONG_TICKET"}
	data, err := json.Marshal(clientPayload)
	if err != nil {
		t.Fatalf("Failed to marshal client payload: %v", err)
	}
	if err := handshake.WriteFrame(clientConn, data); err != nil {
		t.Fatalf("Client failed to write request frame: %v", err)
	}

	// 3. Client Action (Main Thread) - Read Response (should be 'failure')
	var resp handshake.Response
	if err := handshake.ReadFrame(clientConn, &resp); err != nil {
		t.Fatalf("Failed to read server failure response: %v", err)
	}

	// 4. Verification
	if resp.Status != "failure" {
		t.Errorf("Expected status 'failure', got '%s'", resp.Status)
	}

	// 5. Wait for the server Goroutine to finish and check for the expected verification error
	if err := <-serverDone; err == nil {
		t.Errorf("Server Goroutine was expected to return a ticket verification error but returned nil")
	} else {
		// Verify it was a ticket validation failure (based on VerifyTicket logic in handshake.go)
		if !bytes.Contains([]byte(err.Error()), []byte("ticket validation failed")) {
			t.Errorf("Server Goroutine returned unexpected error: %v", err)
		}
	}
}

// ====================================================================
// Test SendFileInfo
// ====================================================================

func TestSendFileInfo_Success(t *testing.T) {
	// Create a mock net.Conn for the output
	var buf bytes.Buffer

	// Create a mock file and cleanup function
	filePath, fileSize, cleanup := createTestFile(t, "Hello, world!")
	defer cleanup()

	fileInfo, _ := os.Stat(filePath)
	// Assuming TCPHost is defined in sender.go and available in the same package 'p2p'
	host := &TCPHost{file: fileInfo}

	// Call the function under test (sends file info to 'buf')
	// Note: bytes.Buffer implements io.Writer, which is the necessary interface.
	if err := host.SendFileInfo(&buf); err != nil {
		t.Fatalf("SendFileInfo failed: %v", err)
	}

	// Verify the data written to the buffer
	// 1. Read the Frame (buf acts as the io.Reader)
	var payload handshake.FileInfoPayload
	if err := handshake.ReadFrame(&buf, &payload); err != nil {
		t.Fatalf("Failed to read FileInfoPayload: %v", err)
	}

	// 2. Assert the contents
	if payload.FileName != fileInfo.Name() {
		t.Errorf("Expected filename %s, got %s", fileInfo.Name(), payload.FileName)
	}
	if payload.Size != fileSize {
		t.Errorf("Expected size %d, got %d", fileSize, payload.Size)
	}
	if payload.Type != "FILE_INFO" {
		t.Errorf("Expected type FILE_INFO, got %s", payload.Type)
	}
}

// ====================================================================
// Test StreamFile
// ====================================================================

func TestStreamFile_Success(t *testing.T) {
	// Define known content and create the test file
	expectedContent := "This is the content that should be streamed over the wire."
	filePath, _, cleanup := createTestFile(t, expectedContent)
	defer cleanup()

	// Mock the connection output (what the client receives)
	var outputBuffer bytes.Buffer

	// Call the function under test (streams content from filePath to outputBuffer)
	// IMPORTANT: We must pass io.Writer as the first argument, as StreamFile is the
	// destination for the file data.
	if err := StreamFile(io.Writer(&outputBuffer), filePath); err != nil {
		t.Fatalf("StreamFile failed unexpectedly: %v", err)
	}

	// Verify the streamed content
	actualContent := outputBuffer.String()
	if actualContent != expectedContent {
		t.Errorf("Streamed content mismatch.\nExpected: %q\nGot:      %q", expectedContent, actualContent)
	}

	// Further verification: Check for EOF (the actual code does not write EOF explicitly,
	// the client detects EOF when io.ReadFull fails, so we rely on content check).
}

func TestStreamFile_FileNotFound(t *testing.T) {
	// Create a mock net.Conn for the output
	var outputBuffer bytes.Buffer

	// File path that definitely does not exist
	nonExistentPath := "/tmp/nonexistent-file-xyz123"

	// Call the function under test
	err := StreamFile(io.Writer(&outputBuffer), nonExistentPath)

	// Assert that an error was returned
	if err == nil {
		t.Fatalf("StreamFile was expected to fail for non-existent file but succeeded")
	}

	// Assert that the error message indicates file opening failure
	if !bytes.Contains([]byte(err.Error()), []byte("error opening the file")) {
		t.Errorf("Expected file opening error, got: %v", err)
	}
}
