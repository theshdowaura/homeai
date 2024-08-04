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
Bemfa client for remote command execution

Usage:
  bemfa-client [flags]

Flags:
  -c, --clientid string   巴法云私钥 (必填)
  -m, --command string    要执行的命令 (必填)
  -h, --help              help for bemfa-client
  -s, --status string     设置设备开关状态on/off(必填)
  -t, --topic string      主题值 (必填)

```

## 发行版下载

请确定您的机器的操作系统类型及其cpu版本，自行下载。