//go:build windows
// +build windows

package ipc

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// NewNamedPipeClient connects to an existing named pipe.
// Example pipe name: "AegisPipe"
func NewNamedPipeClient(pipeName string) (*NamedPipeIPC, error) {
	pipePath := `\\.\pipe\` + pipeName

	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(pipePath),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,   // no sharing
		nil, // default security
		windows.OPEN_EXISTING,
		0, // default attributes
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open named pipe %s: %w", pipeName, err)
	}

	return &NamedPipeIPC{
		handle: handle,
	}, nil
}

// NewNamedPipeServer creates a new named pipe server.
func NewNamedPipeServer(pipeName string) (*NamedPipeIPC, error) {
	pipePath := `\\.\pipe\` + pipeName

	// Create the named pipe
	handle, err := windows.CreateNamedPipe(
		windows.StringToUTF16Ptr(pipePath),
		windows.PIPE_ACCESS_DUPLEX,
		windows.PIPE_TYPE_MESSAGE|windows.PIPE_READMODE_MESSAGE|windows.PIPE_WAIT,
		windows.PIPE_UNLIMITED_INSTANCES,
		4096,
		4096,
		0,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create named pipe %s: %w", pipeName, err)
	}

	return &NamedPipeIPC{
		handle:   handle,
		isServer: true,
	}, nil
}

type NamedPipeIPC struct {
	handle    windows.Handle
	isServer  bool
	connected bool
}

func (p *NamedPipeIPC) ensureConnected() error {
	if !p.isServer || p.connected {
		return nil
	}
	// ConnectNamedPipe waits for a client to connect.
	// Returns error if failed, or nil if connected.
	// If client already connected between Create and Connect, it returns ERROR_PIPE_CONNECTED, which is a success for us.
	err := windows.ConnectNamedPipe(p.handle, nil)
	if err != nil {
		if err == windows.ERROR_PIPE_CONNECTED {
			p.connected = true
			return nil
		}
		return err
	}
	p.connected = true
	return nil
}

// Send writes a message to the named pipe
func (p *NamedPipeIPC) Send(msg []byte) error {
	if err := p.ensureConnected(); err != nil {
		return fmt.Errorf("failed to connect pipe: %w", err)
	}

	var written uint32
	err := windows.WriteFile(p.handle, msg, &written, nil)
	if err != nil {
		return fmt.Errorf("failed to write to named pipe: %w", err)
	}
	if written != uint32(len(msg)) {
		return fmt.Errorf("partial write to named pipe: wrote %d of %d bytes", written, len(msg))
	}
	return nil
}

// Receive reads a message from the named pipe
func (p *NamedPipeIPC) Receive() ([]byte, error) {
	if err := p.ensureConnected(); err != nil {
		return nil, fmt.Errorf("failed to connect pipe: %w", err)
	}

	buf := make([]byte, 4096) // adjust buffer size if needed
	var read uint32
	err := windows.ReadFile(p.handle, buf, &read, nil)
	if err != nil {
		if err == windows.ERROR_BROKEN_PIPE {
			// Client disconnected
			p.connected = false
			return nil, fmt.Errorf("client disconnected")
		}
		return nil, fmt.Errorf("failed to read from named pipe: %w", err)
	}
	return buf[:read], nil
}

// Close closes the pipe handle
func (p *NamedPipeIPC) Close() error {
	if p.isServer {
		windows.DisconnectNamedPipe(p.handle)
	}
	return windows.CloseHandle(p.handle)
}
