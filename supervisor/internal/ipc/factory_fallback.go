//go:build !windows && !darwin && !linux

package ipc

import "fmt"

// NewIPC provides a fallback for unsupported platforms.
func NewIPC(name string) (IPC, error) {
	return nil, fmt.Errorf("IPC not implemented for this platform")
}
