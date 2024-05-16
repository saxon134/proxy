package main

import (
	"github.com/saxon134/proxy"
	"log"
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
	proxy.Init()

	log.Printf("APP已启动\n")

	for {
		time.Sleep(time.Second)
	}
}
