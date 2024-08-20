package mqtt

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	host     string
	port     int
	clientID string
	topic    string
)

func InitMQTT(h string, p int, cid string, t string) {
	host = h
	port = p
	clientID = cid
	topic = t

	runMQTT()
}

func onConnect(client mqtt.Client) {
	fmt.Println("Connected with result code 0")
	token := client.Subscribe(topic, 0, onMessage)
	token.Wait()
	if token.Error() != nil {
		fmt.Println("Subscription error:", token.Error())
	} else {
		fmt.Println("Subscribed to topic:", topic)
	}
}

func onMessage(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("主题: %s 消息: %s\n", msg.Topic(), string(msg.Payload()))
	executeCommand(string(msg.Payload()))
}

func onConnectionLost(client mqtt.Client, err error) {
	if err != nil {
		fmt.Printf("Connection lost: %v\n", err)
	}
}

func runMQTT() {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", host, port)).SetClientID(clientID)
	opts.OnConnect = onConnect
	opts.OnConnectionLost = onConnectionLost

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		log.Fatal(token.Error())
	}
	defer client.Disconnect(250)

	// 使用 Loop() 方法替代 LoopForever()
	for {
		client.Subscribe(topic, 0, onMessage)
	}
}

func executeCommand(command string) {
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
