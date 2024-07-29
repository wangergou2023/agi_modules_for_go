package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	openai "github.com/sashabaranov/go-openai"
	"github.com/wangergou2023/agi_modules_for_go/config"
	"github.com/wangergou2023/agi_modules_for_go/xiao_wan"
)

var cfg = config.New()

func main() {
	fmt.Println("xiao wan is starting up... Please wait a moment.")

	config := openai.DefaultConfig(cfg.OpenAiAPIKey())
	//need"/v1"
	config.BaseURL = cfg.OpenAibaseURL()
	openaiClient := openai.NewClientWithConfig(config)
	openaiClient_face := openai.NewClientWithConfig(config)

	xiao_wan_chat := xiao_wan.Start(cfg, openaiClient)
	xiao_wan_chat_face := xiao_wan.StartOne(cfg, openaiClient_face, xiao_wan.FacePrompt)

	// 启动MQTT订阅
	go startMQTTClient(&xiao_wan_chat)
	// 等3秒订阅成功
	time.Sleep(3 * time.Second)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		response, _ := xiao_wan_chat.Message(text)
		fmt.Printf("xiao wan:%s\r\n", response)
		response2, _ := xiao_wan_chat_face.MessageOne(response)
		fmt.Printf("xiao wan face:%s\r\n", response2)

	}
}

// 启动MQTT客户端，订阅消息
func startMQTTClient(xiao_wan_chat *xiao_wan.Xiao_wan) {
	// 生成随机客户端ID
	clientID := fmt.Sprintf("xiao_wan_client_%d", time.Now().UnixNano())
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL()).
		SetClientID(clientID).
		SetUsername(cfg.MQTTUsername()).
		SetPassword(cfg.MQTTPassword())
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Error connecting to MQTT broker: %v\n", token.Error())
		return
	}

	topic := "plugin/messages"
	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		message := string(msg.Payload())
		fmt.Printf("Received message from plugin: %s\n", message)
		xiao_wan_chat.Message(message)
	}); token.Wait() && token.Error() != nil {
		fmt.Printf("Error subscribing to topic %s: %v\n", topic, token.Error())
		return
	}

	fmt.Println("Subscribed to MQTT topic:", topic)
}
