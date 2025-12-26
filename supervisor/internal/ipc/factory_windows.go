//go:build windows
// +build windows

package ipc

// NewIPC creates a platform-specific IPC client.
// On Windows, this creates a Named Pipe client.
func NewIPC(name string) (IPC, error) {
	return NewNamedPipeServer(name)
}
