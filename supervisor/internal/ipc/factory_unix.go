//go:build darwin || linux
// +build darwin linux

package ipc

// NewIPC creates a platform-specific IPC client.
// On Unix-like systems (macOS, Linux), this creates a Unix socket client.
func NewIPC(name string) (IPC, error) {
	return NewUnixSocketIPC(name)
}
