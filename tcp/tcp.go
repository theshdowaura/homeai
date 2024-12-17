// tcp/tcp.go
package tcp

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Client 定义了 TCP 客户端结构体
type Client struct {
	clientID string
	topic    string
	command  string
	status   string
	conn     net.Conn
	mutex    sync.Mutex
	stopChan chan struct{} // 管理 goroutine 生命周期
	stopOnce sync.Once     // 确保 Stop 方法只执行一次
	done     chan struct{} // 通知主程序客户端已停止
}

// InitTCP 初始化 TCP 客户端并开始运行
func InitTCP(cid, t, cmd, st string) (*Client, error) {
	client := &Client{
		clientID: cid,
		topic:    t,
		command:  cmd,
		status:   st,
		stopChan: make(chan struct{}), // 初始化停止通道
		done:     make(chan struct{}), // 初始化完成通道
	}
	err := client.Run()
	return client, err
}

// connectTCP 建立 TCP 连接并订阅主题
func (c *Client) connectTCP() error {
	conn, err := net.Dial("tcp", "bemfa.com:8344")
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	subscribeCmd := fmt.Sprintf("cmd=1&uid=%s&topic=%s\r\n", c.clientID, c.topic)
	_, err = conn.Write([]byte(subscribeCmd))
	if err != nil {
		conn.Close()
		return fmt.Errorf("发送订阅命令失败: %v", err)
	}

	c.conn = conn
	log.Println("TCP 连接已建立并成功订阅。")
	return nil
}

// ping 定时发送心跳包
func (c *Client) ping() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.sendPing()
			if err != nil {
				log.Println("发送心跳失败:", err)
				c.reconnect()
				return
			} else {
				log.Println("心跳包已发送。")
			}
		case <-c.stopChan:
			log.Println("心跳 goroutine 退出。")
			return
		}
	}
}

// sendPing 发送单个心跳包
func (c *Client) sendPing() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, err := c.conn.Write([]byte("ping\r\n"))
	return err
}

// handleReceivedData 处理接收到的数据
func (c *Client) handleReceivedData(data []byte) {
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
func (c *Client) readLoop() {
	reader := bufio.NewReader(c.conn)
	for {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 设置读超时
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// 读超时，继续循环以检查停止信号
				continue
			}
			log.Println("读取消息失败:", err)
			c.reconnect()
			return
		}
		// 同步处理消息，避免创建过多 goroutine
		c.handleReceivedData(line)

		// 检查是否接收到停止信号
		select {
		case <-c.stopChan:
			log.Println("读取循环 goroutine 退出。")
			return
		default:
			// 继续读取
		}
	}
}

// reconnect 尝试重新连接到服务器
func (c *Client) reconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 如果已有连接，先关闭旧连接
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		log.Println("现有 TCP 连接已关闭。")
	}

	for i := 1; i <= 3; i++ {
		select {
		case <-c.stopChan:
			log.Println("在重连过程中收到停止信号。")
			return
		default:
			log.Printf("尝试重新连接 (%d/3)...", i)
			err := c.connectTCP()
			if err == nil {
				log.Println("重新连接成功。")
				go c.ping()
				go c.readLoop()
				return
			}
			log.Printf("重新连接尝试 %d 失败: %v", i, err)
			time.Sleep(2 * time.Second)
		}
	}

	log.Println("3 次重新连接尝试均失败，停止客户端。")
	c.Stop() // 在重连失败后停止客户端
}

// Run 启动 TCP 客户端
func (c *Client) Run() error {
	err := c.connectTCP()
	if err != nil {
		log.Println("连接失败:", err)
		return err
	}

	go c.ping()
	go c.readLoop()

	return nil
}

// Stop 优雅地停止客户端
func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopChan)
		if c.conn != nil {
			c.conn.Close()
		}
		log.Println("客户端已优雅停止。")
		close(c.done)
	})
}

// Done 返回一个在客户端停止时关闭的通道
func (c *Client) Done() <-chan struct{} {
	return c.done
}

// sendStatusToBemfa 发送状态到 Bemfa
func (c *Client) sendStatusToBemfa(status string) error {
	encodedStatus := url.QueryEscape(status)
	apiURL := fmt.Sprintf("https://api.bemfa.com/api/device/v1/data/3/push/get/?uid=%s&topic=%s&msg=%s", c.clientID, c.topic, encodedStatus)
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
