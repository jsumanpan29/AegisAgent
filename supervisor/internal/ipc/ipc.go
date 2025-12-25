package ipc

type IPC interface {
	Send(msg []byte) error
	Receive() ([]byte, error)
	Close() error
}
