package main

import (
	"github.com/spf13/cobra"
	"homeai/mqtt"
	"homeai/tcp"
	"log"
)

var (
	mqttHost     string
	mqttPort     int
	mqttClientID string
	mqttTopic    string

	tcpClientID string
	tcpTopic    string
	command     string
	status      string
)

var rootCmd = &cobra.Command{
	Use:   "bemfa-client",
	Short: "Bemfa client for MQTT and TCP",
	Run: func(cmd *cobra.Command, args []string) {
		go mqtt.InitMQTT(mqttHost, mqttPort, mqttClientID, mqttTopic)
		tcp.InitTCP(tcpClientID, tcpTopic, command, status)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&mqttHost, "mqtt-host", "H", "bemfa.com", "MQTT 服务器地址")
	rootCmd.Flags().IntVarP(&mqttPort, "mqtt-port", "P", 9501, "MQTT 服务器端口")
	rootCmd.Flags().StringVarP(&mqttClientID, "mqtt-clientid", "i", "", "MQTT 客户端 ID")
	rootCmd.Flags().StringVarP(&mqttTopic, "mqtt-topic", "t", "", "MQTT 订阅主题")

	rootCmd.Flags().StringVarP(&tcpClientID, "tcp-clientid", "c", "", "TCP 巴法云私钥 ")
	rootCmd.Flags().StringVarP(&tcpTopic, "tcp-topic", "T", "", "TCP 主题值 ")
	rootCmd.Flags().StringVarP(&command, "command", "m", "", "<基于tcp创客云>要执行的命令")
	rootCmd.Flags().StringVarP(&status, "status", "s", "", "设置设备开关状态 on/off")

	//_ = rootCmd.MarkFlagRequired("tcp-clientid")
	//_ = rootCmd.MarkFlagRequired("tcp-topic")
	//_ = rootCmd.MarkFlagRequired("command")
	//_ = rootCmd.MarkFlagRequired("status")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
