package main

import (
	"encoding/json"
	"fmt"
	"github.com/saxon134/proxy/helper"
	"github.com/saxon134/proxy/message"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

type ProxyPool struct {
	ClientConn *net.TCPConn
	TunnelConn *net.TCPConn
	m          sync.Mutex
}

type Pool struct {
	sync.RWMutex
	data map[[16]byte]*net.TCPConn
}

var pool = Pool{}
var cfg = struct {
	Secret     string `json:"secret"`
	PoolPort   int    `json:"poolPort"`
	RemotePort int    `json:"remotePort"`
}{}

func main() {

	//加载配置文件
	loadConfig()

	//监听客户端连接
	go listenClient()

	//监听app请求
	go listenAppRequest()

	//保持程序不退出
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func loadConfig() {
	var bs, err = os.ReadFile("./config.json")
	if err != nil {
		panic("配置读取文件失败")
	}

	err = json.Unmarshal(bs, &cfg)
	if err != nil {
		panic(err)
	}

	if cfg.RemotePort <= 1000 || cfg.PoolPort <= 1000 || cfg.Secret == "" {
		panic("配置有误")
	}
}

// 监听客户端连接
func listenClient() {
	var listener, err = helper.CreateListen(fmt.Sprintf(":%d", cfg.PoolPort))
	if err != nil {
		panic(err)
	}

	//接收客户端连接
	for {
		var conn, err = listener.AcceptTCP()
		if err != nil {
			return
		}

		fmt.Println("Accept Client")

		// 读取client连接消息
		var connInfo = new(message.ConnInfo)
		err = connInfo.Read(conn)
		var sign = helper.MD5(string(connInfo.Time[:]) + cfg.Secret)
		if sign != connInfo.Sign {
			conn.Close()
			return
		}

		//保存映射关系
		var c = pool.get(connInfo.Name)
		if c != nil {
			c.Close()
		}
		pool.set(connInfo.Name, conn)

		//保持连接
		go helper.KeepAlive(conn)
	}
}

// 监听app请求
func listenAppRequest() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var host = helper.MD5(r.Host)
		var client = pool.get(host)
		if client == nil {
			http.Error(w, "未连接 ", http.StatusInternalServerError)
			return
		}

		var err error

		// 转发HTTP请求到TCP连接
		if err = r.Write(client); err != nil {
			var msg = err.Error()
			http.Error(w, msg, http.StatusInternalServerError)
			if strings.Contains(msg, "broken pipe") || strings.Contains(msg, "EOF") {
				pool.set(host, nil)
			}
			return
		}

		// 读取TCP响应
		var buf = make([]byte, 4096)
		_, err = client.Read(buf)
		if err != nil {
			http.Error(w, "Err: "+err.Error(), http.StatusInternalServerError)
			return
		}

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		var c, _, _ = hj.Hijack()
		_, _ = c.Write(buf)
	})
	_ = http.ListenAndServe(fmt.Sprintf(":%d", cfg.RemotePort), nil)
}

func (m *Pool) get(host [16]byte) *net.TCPConn {
	m.RLock()
	defer m.RUnlock()

	value, ok := m.data[host]
	if ok {
		return value
	}
	return nil
}

func (m *Pool) set(host [16]byte, conn *net.TCPConn) {
	m.Lock()
	defer m.Unlock()

	if m.data == nil {
		m.data = map[[16]byte]*net.TCPConn{}
	}

	if conn == nil {
		delete(m.data, host)
	} else {
		m.data[host] = conn
	}
}
