package main

import (
	"github.com/saxon134/proxy"
	"net/http"
	"time"
)

// 本地应用，用于测试内网穿透
func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello, i am app"))
	})

	go http.ListenAndServe(":10311", nil)

	//启动内网穿透
	proxy.Init(proxy.Config{
		PoolAddr:   "127.0.0.1:7005",
		RemoteHost: "127.0.0.1:8005",
		LocalAddr:  "127.0.0.1:10311",
		Secret:     "x7&6rty",
	})
	//proxy.Init(proxy.Config{
	//	PoolAddr:   "172.16.0.230:7005",
	//	RemoteHost: "sh02.frp.wcuiqyu.cn",
	//	LocalAddr:  "127.0.0.1:10311",
	//	Secret:     "x7&6rty",
	//})

	for {
		time.Sleep(time.Second)
	}
}
