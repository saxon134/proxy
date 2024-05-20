package proxy

import (
	"fmt"
	"github.com/saxon134/proxy/helper"
	"github.com/saxon134/proxy/message"
	"log"
	"net"
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
		//重新连接
		if e := recover(); e != nil {
			log.Println(e)
			time.Sleep(time.Second * 5)
			start(cfg)
		}
	}()

	// 参数校验
	if cfg.Secret == "" || strings.Contains(cfg.RemoteHost, " ") {
		panic("[Proxy] config err")
	}

	//开启
	for {
		start(cfg)
		time.Sleep(time.Second * 5)
	}
}

func start(cfg Config) {

	// 连接服务端
	var servConn, err = helper.CreateConnect(cfg.PoolAddr)
	if err != nil {
		fmt.Println("[Proxy] Connect err, ", err)
		return
	}
	servConn.SetKeepAlive(true)

	// 发送消息
	var timestamp = helper.MD5(fmt.Sprintf("%d", time.Now().Unix()))
	err = message.ConnInfo{
		Time: timestamp,
		Sign: helper.MD5(string(timestamp[:]) + cfg.Secret),
		Name: helper.MD5(cfg.RemoteHost),
	}.Write(servConn)

	// 已连接
	fmt.Println("[Proxy] Connected")

	// 数据转发
	for {
		//从远程server读数据
		var data []byte
		data, err = helper.Read(servConn)
		if err != nil {
			servConn.Close()
			fmt.Println("[Proxy] Disconnected, retry...")
			return
		}

		//无数据
		if len(data) == 0 || data[0] == 0 {
			continue
		}

		//连接本地app
		var appConn = connectApp(cfg.LocalAddr)
		if appConn == nil {
			err = helper.Write(servConn, []byte("[Proxy] Connect local app err."))
			if err != nil {
				return
			}
			continue
		}

		//往本地app写数据
		err = helper.Write(appConn, data)
		if err != nil {
			var msg = fmt.Sprintf("[Proxy] Write local app err.\n%s", err.Error())
			err = helper.Write(servConn, []byte(msg))
			if err != nil {
				return
			}
			continue
		}

		//从本地读数据
		data, err = helper.Read(appConn)
		if err != nil {
			var msg = fmt.Sprintf("[Proxy] Read local app err.\n%s", err.Error())
			err = helper.Write(servConn, []byte(msg))
			if err != nil {
				return
			}
			continue
		}

		//无数据
		if len(data) == 0 || data[0] == 0 {
			continue
		}

		//回复远程server数据
		err = helper.Write(servConn, data)
		if err != nil {
			fmt.Println("[Proxy] Write remote server err.\n", err)
			return
		}
	}
}

func connectApp(addr string) (conn *net.TCPConn) {
	var retry = 0
	for {
		var err error
		conn, err = helper.CreateConnect(addr)
		if conn == nil {
			var msg = fmt.Sprintf("Connect local app err.\n%s", err.Error())
			fmt.Println(msg)
			time.Sleep(time.Second * 3)
			retry++
			if retry > 3 {
				return nil
			}
			continue
		}

		return conn
	}
}
