package main

import (
	"fmt"
	"github.com/saxon134/proxy/helper"
	"github.com/saxon134/proxy/message"
	"log"
	"net"
	"net/http"
	"sync"
)

type ProxyPool struct {
	ClientConn *net.TCPConn
	TunnelConn *net.TCPConn
	m          sync.Mutex
}

var ProxyPoolMap = map[[16]byte]*ProxyPool{}

func main() {

	//监听客户端连接
	go listenClient()

	//监听app请求
	go listenAppRequest()

	//监听数据通道
	go listenTunnel()

	//保持程序不退出
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

// 监听客户端连接
func listenClient() {
	var listener, err = helper.CreateListen(":7005")
	if err != nil {
		panic(err)
	}

	//接收客户端连接
	for {
		var conn, err = listener.AcceptTCP()
		if err != nil {
			log.Printf("接收连接失败，错误信息为：%s\n", err.Error())
			return
		}

		fmt.Println("Accept Client")

		// 读取请求信息，超时会断开连接
		var connInfo = new(message.ConnInfo)
		err = connInfo.Read(conn)
		if err != nil {
			conn.Close()
			continue
		}

		//保存映射关系
		var pool = ProxyPoolMap[connInfo.Name]
		if pool == nil {
			pool = &ProxyPool{}
		} else {
			if pool.ClientConn != nil {
				pool.ClientConn.Close()
			}
			if pool.TunnelConn != nil {
				pool.TunnelConn.Close()
			}
		}
		pool.ClientConn = conn
		ProxyPoolMap[connInfo.Name] = pool

		//保持连接
		go helper.KeepAlive(conn)
	}
}

// 监听通道连接
func listenTunnel() {
	var listener, err = helper.CreateListen(":7008")
	if err != nil {
		panic(err)
	}

	for {
		var conn, err = listener.AcceptTCP()
		if err != nil {
			log.Printf("接收连接失败，错误信息为：%s\n", err.Error())
			return
		}

		fmt.Println("Accept Tunnel")

		// 读取请求信息
		var connInfo = new(message.ConnInfo)
		err = connInfo.Read(conn)
		if err != nil {
			conn.Close()
			continue
		}

		//保存映射关系
		var pool = ProxyPoolMap[connInfo.Name]
		if pool == nil {
			pool = &ProxyPool{}
		} else if pool.TunnelConn != nil {
			pool.TunnelConn.Close()
		}
		pool.TunnelConn = conn
		ProxyPoolMap[connInfo.Name] = pool
		go helper.KeepAlive(conn)
	}
}

// 监听app请求
func listenAppRequest() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var name = helper.GetMD5(r.Host)
		var pool = ProxyPoolMap[name]
		if pool == nil || pool.ClientConn == nil {
			http.Error(w, "未初始化 ", http.StatusInternalServerError)
			return
		}

		var err error
		if pool.TunnelConn == nil {
			_, err = pool.ClientConn.Write([]byte("New Connection"))
			http.Error(w, "未初始化 ", http.StatusInternalServerError)
			return
		}

		// 转发HTTP请求到TCP连接
		if err = r.Write(pool.TunnelConn); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 读取TCP响应
		var buf = make([]byte, 4096)
		_, err = pool.TunnelConn.Read(buf)
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
	_ = http.ListenAndServe(":8005", nil)
}
