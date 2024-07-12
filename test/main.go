package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	openai "github.com/sashabaranov/go-openai"
	"github.com/wangergou2023/agi_modules_for_go/config"
	"github.com/wangergou2023/agi_modules_for_go/xiao_wan"
)

var cfg = config.New()

func main() {
	fmt.Println("xiao wan is starting up... Please wait a moment.")

	cfg = cfg.SetOpenAibaseURL("https://llxspace.website/v1")
	cfg = cfg.SetOpenAiAPIKey("sk-ltSixDlOAzM7jrt1E6E2A2F82cF5436fA0367103Be6e09F3")
	cfg = cfg.SetMalvusApiEndpoint("http://llxspace.store:19530")
	cfg = cfg.SetMQTTBrokerURL("tcp://llxspace.store:1883")
	cfg = cfg.SetMQTTUsername("xiao_wan")
	cfg = cfg.SetMQTTPassword("xiao_wan")

	config := openai.DefaultConfig(cfg.OpenAiAPIKey())
	//need"/v1"
	config.BaseURL = cfg.OpenAibaseURL()
	openaiClient := openai.NewClientWithConfig(config)

	xiao_wan_chat := xiao_wan.Start(cfg, openaiClient)

	// 启动MQTT订阅
	go startMQTTClient(&xiao_wan_chat)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		xiao_wan_chat.Message(text)
	}
}

// 启动MQTT客户端，订阅消息
func startMQTTClient(xiao_wan_chat *xiao_wan.Xiao_wan) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL()).
		SetClientID("xiao_wan_client").
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
