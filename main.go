package main

import (
	"log"

	"github.com/spf13/cobra"
	"homeai/mqtt"
	"homeai/tcp"
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
}

var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "Start MQTT client",
	Run: func(cmd *cobra.Command, args []string) {
		mqttClient := mqtt.InitMQTT(mqttHost, mqttPort, mqttClientID, mqttTopic)
		if mqttClient == nil {
			log.Fatal("Failed to initialize MQTT client")
		}
		mqttClient.Run()
	},
}

var tcpCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Start TCP client",
	Run: func(cmd *cobra.Command, args []string) {
		tcp.InitTCP(tcpClientID, tcpTopic, command, status)
		//if tcpClient == nil {
		//	log.Fatal("Failed to initialize TCP client")
		//}
		//tcpClient.Run()
	},
}

func init() {
	// MQTT flags
	mqttCmd.Flags().StringVarP(&mqttHost, "mqtt-host", "H", "bemfa.com", "MQTT 服务器地址")
	mqttCmd.Flags().IntVarP(&mqttPort, "mqtt-port", "P", 9501, "MQTT 服务器端口")
	mqttCmd.Flags().StringVarP(&mqttClientID, "mqtt-clientid", "i", "", "MQTT 客户端 ID")
	mqttCmd.Flags().StringVarP(&mqttTopic, "mqtt-topic", "t", "", "MQTT 订阅主题")

	// TCP flags
	tcpCmd.Flags().StringVarP(&tcpClientID, "tcp-clientid", "c", "", "TCP 巴法云私钥")
	tcpCmd.Flags().StringVarP(&tcpTopic, "tcp-topic", "T", "", "TCP 主题值")
	tcpCmd.Flags().StringVarP(&command, "command", "m", "", "<基于 TCP 创客云>要执行的命令")
	tcpCmd.Flags().StringVarP(&status, "status", "s", "", "设置设备开关状态 on/off")

	// Add subcommands to root command
	rootCmd.AddCommand(mqttCmd)
	rootCmd.AddCommand(tcpCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
