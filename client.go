package proxy

import (
	"fmt"
	"github.com/saxon134/proxy/helper"
	"github.com/saxon134/proxy/message"
	"io"
	"log"
	"time"
)

type Config struct {
	ServAddr   string
	TunnelAddr string
	Sets       []Set
}

type Set struct {
	RemoteHost string
	LocalHost  string
}

func Init() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)

			time.Sleep(time.Second * 2)
			Init()
		}
	}()

	var confg = Config{
		ServAddr:   "172.16.0.230:7005",
		TunnelAddr: "172.16.0.230:7008",
		Sets:       make([]Set, 0, 1),
	}
	confg.Sets = append(confg.Sets, Set{
		RemoteHost: "sh02.frp.wcuiqyu.cn",
		LocalHost:  "127.0.0.1:10311",
	})

	//var confg = Config{
	//	ServAddr:   "127.0.0.1:7005",
	//	TunnelAddr: "127.0.0.1:7008",
	//	Sets:       make([]Set, 0, 1),
	//}
	//confg.Sets = append(confg.Sets, Set{
	//	RemoteHost: "127.0.0.1:8005",
	//	LocalHost:  "127.0.0.1:10311",
	//})

	// 连接服务端
	var servConn, err = helper.CreateConnect(confg.ServAddr)
	if err != nil {
		panic(err)
	}

	// 发送消息
	err = message.ConnInfo{Method: message.TypeRegister, Name: helper.GetMD5(confg.Sets[0].RemoteHost)}.Write(servConn)
	if err != nil {
		panic(err)
	}

	// 读取服务端消息
	for {
		data, err := helper.GetDataFromConnection(2048, servConn)
		if err != nil {
			log.Printf("读取数据失败，错误信息为：%s\n", err.Error())
			servConn.Close()
			panic(err)
		}

		// 判断是否为新连接，如果是新连接，则连接隧道服务器，否则转发消息
		if string(data) == "New Connection" {
			fmt.Println("New Connection")
			var appConn, _ = helper.CreateConnect(confg.Sets[0].LocalHost)

			// 连接隧道服务器
			var tunnelConn, _ = helper.CreateConnect(confg.TunnelAddr)
			err = message.ConnInfo{Method: message.TypeRegister, Name: helper.GetMD5(confg.Sets[0].RemoteHost)}.Write(tunnelConn)

			go io.Copy(tunnelConn, appConn)
			go io.Copy(appConn, tunnelConn)
		}
	}
}
