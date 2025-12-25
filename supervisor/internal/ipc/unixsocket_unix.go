// +build darwin linux

package ipc

import "errors"

type UnixSocketIPC struct{}

func NewUnixSocketIPC(path string) (*UnixSocketIPC, error) {
	// TODO: implement UNIX socket connection
	return &UnixSocketIPC{}, nil
}

func (u *UnixSocketIPC) Send(msg []byte) error {
	return errors.New("not implemented")
}

func (u *UnixSocketIPC) Receive() ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (u *UnixSocketIPC) Close() error {
	return errors.New("not implemented")
}
