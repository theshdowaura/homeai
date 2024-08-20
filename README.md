# 项目介绍

本项目针对巴法云接入米家，可对接小爱音箱启动指定程序对接homeassistant实现智能家居的联动操作。

## 使用方法

### 编译安装
```shell
go mod tidy
go build .
```
```shell
homeai -h
```
```
Bemfa client for MQTT and TCP

Usage:
  bemfa-client [flags]

Flags:
  -m, --command string         <基于tcp创客云>要执行的命令
  -h, --help                   help for bemfa-client
  -i, --mqtt-clientid string   MQTT 客户端 ID
  -H, --mqtt-host string       MQTT 服务器地址 (default "bemfa.com")
  -P, --mqtt-port int          MQTT 服务器端口 (default 9501)
  -t, --mqtt-topic string      MQTT 订阅主题
  -s, --status string          设置设备开关状态 on/off
  -c, --tcp-clientid string    TCP 巴法云私钥
  -T, --tcp-topic string       TCP 主题值

```

## 发行版下载

请确定您的机器的操作系统类型及其cpu版本，自行下载。