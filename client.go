package proxy

import (
	"fmt"
	"github.com/saxon134/proxy/helper"
	"github.com/saxon134/proxy/message"
	"strings"
	"time"
)

type Config struct {
	RemoteHost string
	PoolAddr   string
	LocalAddr  string
	Secret     string
}

// Init 初始化内网穿透，阻塞函数
func Init(cfg Config) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()

	// 参数校验
	if cfg.Secret == "" || strings.Contains(cfg.RemoteHost, " ") {
		panic("config err")
	}

	// 连接服务端
	var servConn, err = helper.CreateConnect(cfg.PoolAddr)
	if err != nil {
		panic(err)
	}

	// 发送消息
	var timestamp = helper.MD5(fmt.Sprintf("%d", time.Now().Unix()))
	err = message.ConnInfo{
		Time: timestamp,
		Sign: helper.MD5(string(timestamp[:]) + cfg.Secret),
		Name: helper.MD5(cfg.RemoteHost),
	}.Write(servConn)

	// 读取服务端消息
	for {
		var data, err = helper.GetDataFromConnection(2048, servConn)
		if err != nil {
			Init(cfg)
			time.Sleep(time.Second * 2)
		}

		var appConn, _ = helper.CreateConnect(cfg.LocalAddr)
		appConn.Write(data)

		appConn.Read(data)
		servConn.Write(data)
	}
}
