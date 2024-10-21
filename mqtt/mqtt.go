package mqtt

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClient 封装了 MQTT 客户端的配置和状态
type MQTTClient struct {
	host     string
	port     int
	clientID string
	topic    string
	client   mqtt.Client
}

var logFileMutex sync.Mutex

// InitMQTT 创建并初始化一个 MQTTClient 实例
func InitMQTT(h string, p int, cid string, t string) *MQTTClient {
	client := &MQTTClient{
		host:     h,
		port:     p,
		clientID: cid,
		topic:    t,
	}
	return client
}

// validateConfig 验证配置参数
func (m *MQTTClient) validateConfig() error {
	if m.host == "" || m.port == 0 || m.clientID == "" || m.topic == "" {
		return fmt.Errorf("配置参数不完整")
	}
	return nil
}

// run 启动 MQTT 客户端
func (m *MQTTClient) Run() {
	if err := m.validateConfig(); err != nil {
		log.Fatal(err)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", m.host, m.port)).
		SetClientID(m.clientID).
		SetAutoReconnect(true).
		SetConnectionLostHandler(m.onConnectionLost).
		SetOnConnectHandler(m.onConnect)

	m.client = mqtt.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	defer m.client.Disconnect(250)

	// 阻塞主线程
	select {}
}

// onConnect 连接成功的回调
func (m *MQTTClient) onConnect(client mqtt.Client) {
	log.Println("连接成功")
	token := client.Subscribe(m.topic, 0, m.onMessage)
	token.Wait()
	if token.Error() != nil {
		log.Println("订阅失败:", token.Error())
	} else {
		log.Println("已订阅主题:", m.topic)
	}
}

// onMessage 接收到消息的回调
func (m *MQTTClient) onMessage(_ mqtt.Client, msg mqtt.Message) {
	payload := string(msg.Payload())
	log.Printf("收到消息 - 主题: %s, 内容: %s\n", msg.Topic(), payload)
	m.executeCommand(payload)
}

// onConnectionLost 连接丢失的回调
func (m *MQTTClient) onConnectionLost(_ mqtt.Client, err error) {
	log.Printf("连接丢失: %v\n", err)
}

// executeCommand 执行命令
func (m *MQTTClient) executeCommand(command string) {
	allowedCommands := map[string]string{
		"start": "systemctl start myservice",
		"stop":  "systemctl stop myservice",
		"ping":  "ping.exe www.baidu.com",
		"ip":    "ipconfig",
		// 添加更多命令映射
	}

	cmdStr, ok := allowedCommands[command]
	if !ok {
		log.Println("收到非法命令:", command)
		return
	}

	log.Printf("执行命令: %s\n", cmdStr)
	cmd := exec.Command(cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	// 准备日志内容
	logEntry := time.Now().Format("2006-01-02 15:04:05") + "\n"
	logEntry += "命令: " + cmdStr + "\n"
	if err != nil {
		logEntry += "错误: " + err.Error() + "\n"
	}
	logEntry += "输出:\n" + out.String() + "\n"
	logEntry += "----------------------------------------\n"

	// 写入日志文件
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	logFile, fileErr := os.OpenFile("command.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if fileErr != nil {
		log.Println("无法打开日志文件:", fileErr)
		return
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {

		}
	}(logFile)

	_, writeErr := logFile.WriteString(logEntry)
	if writeErr != nil {
		log.Println("无法写入日志文件:", writeErr)
	}
	if err != nil {
		log.Println("命令执行失败:", err, out.String())
	} else {
		log.Println("命令执行成功:", out.String())
	}
}
