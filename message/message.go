package message

import (
	"bytes"
	"encoding/binary"
	"net"
)

type Message interface {
	Write(conn net.Conn) error
	Read(conn net.Conn) error
}

// ConnInfo 请求建立连接  共21字节
type ConnInfo struct {
	Name [16]byte // 占16 字节
	Time [16]byte // 占16 字节
	Sign [16]byte // 占16 字节
}

func (c ConnInfo) Write(conn net.Conn) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, &c); err != nil {
		return err
	}
	_, err := conn.Write(buf.Bytes())
	return err
}

func (c *ConnInfo) Read(conn net.Conn) error {
	data := make([]byte, 48)
	n, err := conn.Read(data)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	buf.Write(data[:n])
	err = binary.Read(buf, binary.BigEndian, c)
	return err
}
