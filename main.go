package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// 巴法云私钥
const clientID = "0c02529a94f04e41a4999b8dc7058047"

// 主题值
const topic = "PVElxc100001"

// ipmitool命令路径

// 发送状态到巴法云
func sendStatusToBemfa(status string) error {
	encodedStatus := url.QueryEscape(status)
	apiURL := fmt.Sprintf("https://api.bemfa.com/api/device/v1/data/3/push/get/?uid=%s&topic=%s&msg=%s", clientID, topic, encodedStatus)
	_, err := http.Get(apiURL)
	return err
}

// 连接到TCP服务器
func connectTCP() (*net.Conn, error) {
	conn, err := net.Dial("tcp", "bemfa.com:8344")
	if err != nil {
		return nil, err
	}

	// 发送订阅指令
	subscribeCmd := fmt.Sprintf("cmd=1&uid=%s&topic=%s\r\n", clientID, topic)
	_, err = conn.Write([]byte(subscribeCmd))
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

// 心跳
func ping(conn *net.Conn, mutex *sync.Mutex) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mutex.Lock()
			_, err := (*conn).Write([]byte("ping\r\n"))
			mutex.Unlock()
			if err != nil {
				fmt.Println("发送心跳失败:", err)
				// 尝试重连
				for i := 0; i < 3; i++ {
					fmt.Println("尝试重连...")
					conn, err = connectTCP()
					if err == nil {
						fmt.Println("重连成功")
						go ping(conn, mutex) // 重启心跳线程
						return
					}
					fmt.Println("重连失败:", err)
					time.Sleep(2 * time.Second)
				}
				fmt.Println("重连失败，退出程序")
				// 可以选择在这里退出程序，或者进行其他处理
			}
		}
	}
}

// 处理接收到的数据
func handleReceivedData(data []byte, mutex *sync.Mutex) {
	// 解析数据
	dataStr := string(data)
	if strings.Contains(dataStr, "cmd=2") && strings.Contains(dataStr, "msg=off") {
		// 执行命令
		fmt.Println("执行命令: shutdown now")
		cmd := exec.Command("shutdown", "now")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			fmt.Println("命令执行失败:", err)
		} else {
			fmt.Println("命令执行成功:", out.String())
		}
	}
}

func main() {
	// 连接到TCP服务器
	var conn *net.Conn
	var err error
	for i := 0; i < 3; i++ {
		conn, err = connectTCP()
		if err != nil {
			fmt.Println("连接失败:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		fmt.Println("连接失败，退出程序")
		return
	}
	defer (*conn).Close()

	// 使用互斥锁保护TCP连接
	var mutex sync.Mutex

	// 启动心跳线程
	go ping(conn, &mutex)

	// 循环读取消息
	reader := bufio.NewReader(*conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println("读取消息失败:", err)
			// 尝试重连
			for i := 0; i < 3; i++ {
				fmt.Println("尝试重连...")
				conn, err = connectTCP()
				if err == nil {
					fmt.Println("重连成功")
					reader = bufio.NewReader(*conn)
					go ping(conn, &mutex) // 重启心跳线程
					break
				}
				fmt.Println("重连失败:", err)
				time.Sleep(2 * time.Second)
			}
			if err != nil {
				fmt.Println("重连失败，退出程序")
				return
			}
			continue
		}

		// 处理接收到的数据
		go handleReceivedData(line, &mutex)
	}
}
