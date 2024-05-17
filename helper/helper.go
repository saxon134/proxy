package helper

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// CreateListen 监听，参数为监听地址listenAddr，返回 TCPListener，通过 net.ResolveTCPAddr 解析地址，通过 net.ListenTCP 监听端口
//
//	监听是指服务端监听某个端口，等待客户端的连接，一旦客户端连接上来，服务端就会创建一个新的goroutine处理客户端的请求。
//
// ResolveTCPAddr是一个解析TCP地址的函数，addr为域名或者IP地址加端口号，返回一个TCPAddr，该结构体包含了ip和port
// ListenTCP函数监听TCP地址，addr则是一个TCP地址，如果addr的端口字段为0，函数将选择一个当前可用的端口，返回值l是一个net.Listener接口，可以用来接收连接。
func CreateListen(listenAddr string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	return tcpListener, err
}

// CreateConnect 连接，参数为服务端地址connectAddr，返回 TCPConn，通过 net.ResolveTCPAddr 解析地址，通过 net.DialTCP 连接服务端
// 连接是指客户端连接服务端，连接成功后，客户端就可以向服务端发送数据了，与监听不同的是，连接是客户端发起的，而监听是服务端发起的。
// DialTCP函数在网络协议tcp上连接本地地址laddr和远端地址raddr，如果laddr为nil，则自动选择本地地址，如果raddr为nil，则函数在建立连接之前不会尝试解析地址，一般用于客户端。
func CreateConnect(connectAddr string) (*net.TCPConn, error) {
	// 解析地址,返回TCPAddr
	tcpAddr, err := net.ResolveTCPAddr("tcp", connectAddr)
	if err != nil {
		return nil, err
	}

	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	return tcpConn, err
}

// KeepAlive 保持连接,参数为连接conn，通过循环向连接中写入数据，保持连接,每隔3秒写入一次,如果写入失败，说明连接已经断开，退出循环
func KeepAlive(conn *net.TCPConn) {
	//// 开启keepalive
	//conn.SetKeepAlive(true)
	//
	//// 设置keepalive时间间隔
	//var keepAliveTime = time.Duration(30) * time.Second
	//var err = conn.SetKeepAlivePeriod(keepAliveTime)
	//if err != nil {
	//	return
	//}
	//
	//// 设置keepalive探测次数和超时时间
	//err = conn.SetKeepAlivePeriod(15 * time.Second)
	//if err != nil {
	//	return
	//}

	//for {
	//	_, err := conn.Write([]byte("KeepAlive"))
	//	if err != nil {
	//		var msg = err.Error()
	//		if strings.Contains(msg, "broken pipe") || strings.Contains(msg, "EOF") {
	//			log.Printf("[-] Client broked")
	//		}
	//		return
	//	}
	//	time.Sleep(time.Second * 5)
	//}
}

// GetDataFromConnection for循环获取Connection中的数据
func GetDataFromConnection(bufSize int, conn *net.TCPConn) ([]byte, error) {
	b := make([]byte, 0)
	for {
		// 读取数据
		data := make([]byte, bufSize)
		n, err := conn.Read(data)
		if err != nil {
			return nil, err
		}

		b = append(b, data[:n]...)
		if n < bufSize {
			break
		}
	}
	return b, nil
}

func DataForward(from *net.TCPConn, to *net.TCPConn) (fromErr error, toError error) {
	for {
		var err error
		var bytes [1024]byte
		var cnt int
		cnt, err = from.Read(bytes[:])
		if err != nil {
			fmt.Println("MessageForward Read err:", err)
			return err, nil
		}
		fmt.Println("MessageForward Read")

		if cnt <= 0 {
			time.Sleep(time.Second)
			continue
		}

		_, err = to.Write(bytes[:])
		if err != nil {
			fmt.Println("MessageForward Write err:", err)
			return nil, err
		}
		fmt.Println("MessageForward Write")
	}
}

// DataExchange 数据交换
func DataExchange(conn1, conn2 net.Conn) {
	defer func() {
		conn1.Close()
		conn2.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		buffer := make([]byte, 4096)
		for {
			n, err := conn1.Read(buffer)
			if err != nil {
				return
			}
			n, err = conn2.Write(buffer[:n])
			if err != nil {
				return
			}
		}
	}()

	go func() {
		defer cancel()
		buffer := make([]byte, 4096)
		for {
			n, err := conn2.Read(buffer)
			if err != nil {
				return
			}
			n, err = conn1.Write(buffer[:n])
			if err != nil {
				return
			}
		}
	}()
	<-ctx.Done()
}

func MD5(s string) [16]byte {
	data := []byte(s)
	has := md5.Sum(data)
	//return fmt.Sprintf("%x", has), has
	return has
}

func Read(conn *net.TCPConn) ([]byte, error) {
	var data = make([]byte, 1024)
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	var _, err = conn.Read(data)
	if isErr(err) {
		return nil, err
	}
	return data, nil
}

func Write(conn *net.TCPConn, data []byte) error {
	_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	var _, err = conn.Write(data)
	return err
}

func isErr(err error) bool {
	if err != nil && err != io.EOF && strings.Contains(err.Error(), "i/o timeout") == false {
		return true
	}
	return false
}
