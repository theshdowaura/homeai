package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
)

const (
	defaultHost     = "bemfa.com"
	defaultPort     = 9501
	defaultClientID = "4d9ec352e0376f2110a0c601a2857225"
	defaultTopic    = "led00202"
)

var (
	host     string
	port     int
	clientID string
	topic    string
	userName string
	passwd   string
)

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
}

func onConnectionLost(client mqtt.Client, err error) {
	if err != nil {
		fmt.Printf("Connection lost: %v\n", err)
	}
}

func runMQTT() {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", host, port)).SetClientID(clientID)
	//opts.SetUsername(userName)
	//opts.SetPassword(passwd)
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

var rootCmd = &cobra.Command{
	Use:   "mqtt-client",
	Short: "MQTT 客户端",
	Run: func(cmd *cobra.Command, args []string) {
		runMQTT()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&host, "host", "H", defaultHost, "MQTT 服务器地址")
	rootCmd.Flags().IntVarP(&port, "port", "P", defaultPort, "MQTT 服务器端口")
	rootCmd.Flags().StringVarP(&clientID, "client-id", "i", defaultClientID, "客户端 ID")
	rootCmd.Flags().StringVarP(&topic, "topic", "t", defaultTopic, "订阅主题")
	//rootCmd.Flags().StringVarP(&userName, "username", "u", "", "用户名")
	//rootCmd.Flags().StringVarP(&passwd, "password", "p", "", "密码")
	//rootCmd.MarkFlagRequired("username")
	//rootCmd.MarkFlagRequired("password")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
