package message

import (
	"bytes"
	"encoding/binary"
	"net"
)

const TypeRegister = 1 // 注册请求
const TypeRequest = 2  // 请求连接
const TypeResponse = 3 // 回应

type Message interface {
	Write(conn net.Conn) error
	Read(conn net.Conn) error
}

// ConnInfo 请求建立连接  共21字节
type ConnInfo struct {
	Method uint8    // 占1字节
	Name   [16]byte // 占16 字节
	ConnId uint32   // 占4字节
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
	data := make([]byte, 21)
	n, err := conn.Read(data)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	buf.Write(data[:n])
	err = binary.Read(buf, binary.BigEndian, c)
	return err
}