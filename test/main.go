package main

import (
	"bufio"
	"encoding/json"
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

// 定义一个宏控制TTS的使用，因为目前只是生成了mp3文件并没有播放
const enableTTS = false
const enableSTT = false
const enableNeed = false

func main() {
	targets := map[string]string{
		"1": "小丸",
		"2": "风间",
	}

	fmt.Println("xiao wan is starting up... Please wait a moment.")

	config := openai.DefaultConfig(cfg.OpenAiAPIKey())
	//need"/v1"
	config.BaseURL = cfg.OpenAibaseURL()
	openaiClient := openai.NewClientWithConfig(config)
	openaiClient_friend_fengjian := openai.NewClientWithConfig(config)

	var xiao_wan_chat_stt xiao_wan.Xiao_wan
	var xiao_wan_chat_tts xiao_wan.Xiao_wan
	var xiao_wan_chat_face xiao_wan.Xiao_wan
	var xiao_wan_chat_legs xiao_wan.Xiao_wan
	var xiao_wan_friend_duolaameng xiao_wan.Xiao_wan
	if enableSTT {
		openaiClient_stt := openai.NewClientWithConfig(config)
		xiao_wan_chat_stt = xiao_wan.StartStt(cfg, openaiClient_stt)
	}

	if enableTTS {
		openaiClient_tts := openai.NewClientWithConfig(config)
		xiao_wan_chat_tts = xiao_wan.StartOne(cfg, openaiClient_tts, xiao_wan.TtsPrompt, "for_after_chat")
	}

	if enableNeed {
		openaiClient_face := openai.NewClientWithConfig(config)
		openaiClient_legs := openai.NewClientWithConfig(config)
		openaiClient_friend_duolaameng := openai.NewClientWithConfig(config)
		xiao_wan_chat_face = xiao_wan.StartOne(cfg, openaiClient_face, xiao_wan.FacePrompt, "for_after_chat2")
		xiao_wan_chat_legs = xiao_wan.StartOne(cfg, openaiClient_legs, xiao_wan.LegsPrompt, "for_after_chat3")
		xiao_wan_friend_duolaameng = xiao_wan.StartOne(cfg, openaiClient_friend_duolaameng, xiao_wan.DuolaamengPrompt, "for_before_chat")
	}

	xiao_wan_chat := xiao_wan.Start(cfg, openaiClient)
	xiao_wan_friend_fengjian := xiao_wan.StartOne(cfg, openaiClient_friend_fengjian, xiao_wan.FengjianPrompt, "for_before_chat")

	// 启动MQTT订阅
	go startMQTTClient(&xiao_wan_chat)
	// 等3秒订阅成功
	time.Sleep(3 * time.Second)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	// 目前只是测试
	if enableSTT {
		xiao_wan_chat_stt.Stt()
	}

	for {
		fmt.Println("选择对话对象: 1. 小丸 2. 风间")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		targetName, ok := targets[choice]
		if !ok {
			fmt.Println("无效选择，请重试")
			continue
		}

		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		// 构建Result
		result_zhuren := xiao_wan.Result{
			Responses: []struct {
				TargetName string `json:"target_name"`
				Message    string `json:"message"`
			}{
				{
					TargetName: targetName,
					Message:    text,
				},
			},
			YourName: "主人",
		}

		// 将 Result 转换为 JSON 字符串
		resultJSON, err := json.Marshal(result_zhuren)
		if err != nil {
			fmt.Println("转换为 JSON 失败:", err)
			continue
		}

		fmt.Printf("zhu ren:%s\r\n", string(resultJSON))

		if enableNeed {
			duolaameng_response, _, _ := xiao_wan_friend_duolaameng.MessageOne(text)
			fmt.Printf("duolaameng:%s\r\n", duolaameng_response)
			xiao_wan_friend_duolaameng.SaveConversationToJSON(duolaameng_response)
		}

		response, result, _ := xiao_wan_chat.Message(string(resultJSON))
		fmt.Printf("xiao wan:%s\r\n", response)

		for _, res := range result.Responses {
			if res.TargetName == "风间" {
				response2, result2, _ := xiao_wan_friend_fengjian.MessageOne(result.YourName + "：" + res.Message)
				fmt.Printf("feng jian:%s\r\n", response2)
				fmt.Printf("feng jian:%s\r\n", result2)
				// for _, res := range result2.Responses {
				// 	if res.TargetName == "小丸" {
				// 		response3, result3, _ := xiao_wan_chat.Message(result2.YourName + "：" + res.Message)
				// 		fmt.Printf("xiao wan:%s\r\n", response3)
				// 		fmt.Printf("xiao wan:%s\r\n", result3)
				// 	}
				// }
			}
		}

		if enableTTS {
			response2, _, _ := xiao_wan_chat_tts.MessageOne(response)
			fmt.Printf("xiao wan tts:%s\r\n", response2)
		}

		if enableNeed {
			go xiao_wan_chat_face.MessageOne(response)
			go xiao_wan_chat_legs.MessageOne(response)
		}
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
