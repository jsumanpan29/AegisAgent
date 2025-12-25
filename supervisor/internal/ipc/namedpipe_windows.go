//go:build windows
// +build windows

package ipc

import (
	"fmt"

	"golang.org/x/sys/windows"
)

type NamedPipeIPC struct {
	handle windows.Handle
}

// NewNamedPipeIPC connects to an existing named pipe.
// Example pipe name: "AegisPipe"
func NewNamedPipeIPC(pipeName string) (*NamedPipeIPC, error) {
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

// Send writes a message to the named pipe
func (p *NamedPipeIPC) Send(msg []byte) error {
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
	buf := make([]byte, 4096) // adjust buffer size if needed
	var read uint32
	err := windows.ReadFile(p.handle, buf, &read, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read from named pipe: %w", err)
	}
	return buf[:read], nil
}

// Close closes the pipe handle
func (p *NamedPipeIPC) Close() error {
	return windows.CloseHandle(p.handle)
}
