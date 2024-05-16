# prox

### 介绍

内网穿透

server编译后，部署到带公网IP的服务器上

app是测试样例，app是部署在内网的项目，引用client.go来初始化

外网访问server服务器，会将请求通过client转发到app，进而实现内网穿透

支持多个内网app映射，通过访问的域名（host）来区分具体指向哪个app

故app调用client的时候，初始化时，需指定host，以及app的内网地址

支持设置秘钥，秘钥不对server会主动断开连接

详见：app/main.go


### 安装
```
go get -u github.com/saxon134/proxy

```
