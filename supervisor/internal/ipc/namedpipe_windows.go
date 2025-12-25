// +build windows

package ipc

import "errors"

type NamedPipeIPC struct{}

func NewNamedPipeIPC(pipeName string) (*NamedPipeIPC, error) {
	// TODO: implement real named pipe connection
	return &NamedPipeIPC{}, nil
}

func (p *NamedPipeIPC) Send(msg []byte) error {
	// TODO: send message via named pipe
	return errors.New("not implemented")
}

func (p *NamedPipeIPC) Receive() ([]byte, error) {
	// TODO: receive message via named pipe
	return nil, errors.New("not implemented")
}

func (p *NamedPipeIPC) Close() error {
	// TODO: close pipe
	return errors.New("not implemented")
}
