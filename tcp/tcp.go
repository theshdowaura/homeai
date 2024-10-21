package tcp

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TCPClient 定义了 TCP 客户端结构体
type TCPClient struct {
	clientID string
	topic    string
	command  string
	status   string
	conn     net.Conn
	mutex    sync.Mutex
}

// InitTCP 初始化 TCP 客户端并开始运行
func InitTCP(cid string, t string, cmd string, st string) {
	client := &TCPClient{
		clientID: cid,
		topic:    t,
		command:  cmd,
		status:   st,
	}
	client.Run()
}

// connectTCP 建立 TCP 连接并订阅主题
func (c *TCPClient) connectTCP() error {
	conn, err := net.Dial("tcp", "bemfa.com:8344")
	if err != nil {
		return err
	}

	subscribeCmd := fmt.Sprintf("cmd=1&uid=%s&topic=%s\r\n", c.clientID, c.topic)
	_, err = conn.Write([]byte(subscribeCmd))
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

// ping 定时发送心跳包
func (c *TCPClient) ping() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.mutex.Lock()
		_, err := c.conn.Write([]byte("ping\r\n"))
		c.mutex.Unlock()
		if err != nil {
			log.Println("发送心跳失败:", err)
			c.reconnect()
			return
		}
	}
}

// handleReceivedData 处理接收到的数据
func (c *TCPClient) handleReceivedData(data []byte) {
	dataStr := string(data)
	msgstatus := fmt.Sprintf("msg=%s", c.status)
	if strings.Contains(dataStr, "cmd=2") && strings.Contains(dataStr, msgstatus) {
		log.Printf("执行命令: %s\n", c.command)
		cmd := exec.Command("bash", "-c", c.command)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Println("命令执行失败:", err)
		} else {
			log.Println("命令执行成功:", out.String())
		}
	}
}

// readLoop 持续读取服务器发送的数据
func (c *TCPClient) readLoop() {
	reader := bufio.NewReader(c.conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println("读取消息失败:", err)
			c.reconnect()
			return
		}
		go c.handleReceivedData(line)
	}
}

// reconnect 尝试重新连接服务器
func (c *TCPClient) reconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.conn.Close()

	for i := 0; i < 3; i++ {
		log.Println("尝试重连...")
		err := c.connectTCP()
		if err == nil {
			log.Println("重连成功")
			go c.ping()
			go c.readLoop()
			return
		}
		log.Println("重连失败:", err)
		time.Sleep(2 * time.Second)
	}
	log.Println("重连失败，退出程序")
}

// Run 启动 TCP 客户端
func (c *TCPClient) Run() {
	err := c.connectTCP()
	if err != nil {
		log.Println("连接失败，退出程序")
		return
	}
	defer c.conn.Close()

	go c.ping()
	c.readLoop()
}

// sendStatusToBemfa 发送状态到 Bemfa
func (c *TCPClient) sendStatusToBemfa(status string) error {
	encodedStatus := url.QueryEscape(status)
	apiURL := fmt.Sprintf("https://api.bemfa.com/api/device/v1/data/3/push/get/?uid=%s&topic=%s&msg=%s", c.clientID, c.topic, encodedStatus)
	_, err := http.Get(apiURL)
	return err
}
