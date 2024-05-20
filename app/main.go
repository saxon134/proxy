package main

import (
	"fmt"
	"github.com/saxon134/proxy"
	"io"
	"net/http"
)

// 本地应用，用于测试内网穿透
func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

		var msg = "hello, i am app\n"
		msg += fmt.Sprintf("Method: %s\n", request.Method)
		msg += fmt.Sprintf("Content-Type: %s\n", request.Header["Content-Type"])

		var body, _ = io.ReadAll(request.Body)
		msg += fmt.Sprintf("Body: %s\n", string(body))

		fmt.Println(msg)

		writer.Write([]byte(msg))
	})

	go http.ListenAndServe(":10311", nil)

	//启动内网穿透
	//proxy.Init(proxy.Config{
	//	PoolAddr:   "127.0.0.1:7005",
	//	RemoteHost: "127.0.0.1:8005",
	//	LocalAddr:  "127.0.0.1:10311",
	//	Secret:     "x7&6rty",
	//})
	proxy.Init(proxy.Config{
		PoolAddr:   "172.16.0.230:7005",
		RemoteHost: "sh02.frp.wcuiqyu.cn",
		LocalAddr:  "127.0.0.1:10311",
		Secret:     "x7&6rty82ux",
	})
}
