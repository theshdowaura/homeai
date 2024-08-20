package tcp

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

var (
	clientID string
	topic    string
	command  string
	status   string
)

// InitTCP 初始化 TCP 连接并启动相关处理
func InitTCP(cid string, t string, cmd string, st string) {
	clientID = cid
	topic = t
	command = cmd
	status = st

	runTCP()
}

func sendStatusToBemfa(status string) error {
	encodedStatus := url.QueryEscape(status)
	apiURL := fmt.Sprintf("https://api.bemfa.com/api/device/v1/data/3/push/get/?uid=%s&topic=%s&msg=%s", clientID, topic, encodedStatus)
	_, err := http.Get(apiURL)
	return err
}

func connectTCP() (*net.Conn, error) {
	conn, err := net.Dial("tcp", "bemfa.com:8344")
	if err != nil {
		return nil, err
	}

	subscribeCmd := fmt.Sprintf("cmd=1&uid=%s&topic=%s\r\n", clientID, topic)
	_, err = conn.Write([]byte(subscribeCmd))
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

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
				for i := 0; i < 3; i++ {
					fmt.Println("尝试重连...")
					conn, err = connectTCP()
					if err == nil {
						fmt.Println("重连成功")
						go ping(conn, mutex)
						return
					}
					fmt.Println("重连失败:", err)
					time.Sleep(2 * time.Second)
				}
				fmt.Println("重连失败，退出程序")
				return
			}
		}
	}
}

func handleReceivedData(data []byte, mutex *sync.Mutex) {
	dataStr := string(data)
	msgstatus := fmt.Sprintf("msg=%s", status)
	if strings.Contains(dataStr, "cmd=2") && strings.Contains(dataStr, msgstatus) {
		fmt.Printf("执行命令: %s\n", command)
		cmd := exec.Command("bash", "-c", command)
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

func runTCP() {
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

	var mutex sync.Mutex
	go ping(conn, &mutex)

	reader := bufio.NewReader(*conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println("读取消息失败:", err)
			for i := 0; i < 3; i++ {
				fmt.Println("尝试重连...")
				conn, err = connectTCP()
				if err == nil {
					fmt.Println("重连成功")
					reader = bufio.NewReader(*conn)
					go ping(conn, &mutex)
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

		go handleReceivedData(line, &mutex)
	}
}
