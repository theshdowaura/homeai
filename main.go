// main.go
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

// rootCmd 定义了根命令
var rootCmd = &cobra.Command{
	Use:   "bemfa-client",
	Short: "Bemfa client for MQTT and TCP",
}

// mqttCmd 定义了 MQTT 子命令
var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "启动 MQTT 客户端",
	Run: func(cmd *cobra.Command, args []string) {
		mqttClient := mqtt.InitMQTT(mqttHost, mqttPort, mqttClientID, mqttTopic)
		if mqttClient == nil {
			log.Fatal("初始化 MQTT 客户端失败")
		}
		// 假设 mqttClient.Run 会阻塞或自行管理生命周期
		mqttClient.Run()
	},
}

// tcpCmd 定义了 TCP 子命令
var tcpCmd = &cobra.Command{
	Use:   "tcp",
	Short: "启动 TCP 客户端",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := tcp.InitTCP(tcpClientID, tcpTopic, command, status)
		if err != nil {
			log.Fatal("初始化 TCP 客户端失败:", err)
		}

		// 监听中断信号
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(signalChan)

		// 等待信号或客户端停止
		select {
		case sig := <-signalChan:
			log.Printf("接收到信号: %s. 正在关闭...", sig)
			client.Stop()
			// 等待客户端停止
			<-client.Done()
		case <-client.Done():
			log.Println("客户端已停止。")
		}
	},
}

func init() {
	// MQTT 命令行参数
	mqttCmd.Flags().StringVarP(&mqttHost, "mqtt-host", "H", "bemfa.com", "MQTT 服务器地址")
	mqttCmd.Flags().IntVarP(&mqttPort, "mqtt-port", "P", 9501, "MQTT 服务器端口")
	mqttCmd.Flags().StringVarP(&mqttClientID, "mqtt-clientid", "i", "", "MQTT 客户端 ID")
	mqttCmd.Flags().StringVarP(&mqttTopic, "mqtt-topic", "t", "", "MQTT 订阅主题")

	// TCP 命令行参数
	tcpCmd.Flags().StringVarP(&tcpClientID, "tcp-clientid", "c", "", "TCP 巴法云私钥")
	tcpCmd.Flags().StringVarP(&tcpTopic, "tcp-topic", "T", "", "TCP 主题值")
	tcpCmd.Flags().StringVarP(&command, "command", "m", "", "<基于 TCP 创客云>要执行的命令")
	tcpCmd.Flags().StringVarP(&status, "status", "s", "", "设置设备开关状态 on/off")

	// 添加子命令到根命令
	rootCmd.AddCommand(mqttCmd)
	rootCmd.AddCommand(tcpCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func startPprofServer() {
	go func() {
		log.Println("Starting pprof server on :6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("pprof server failed: %v", err)
		}
	}()
}
