package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &Alarm{}

type Alarm struct {
	cfg          config.Cfg
	openaiClient *openai.Client
	serverAddr   string // TCP服务端地址
}

type AlarmInput struct {
	Duration string `json:"duration"` // 闹钟的持续时间，例如："10s", "2m"
	Event    string `json:"event"`    // 事件名称，例如："会议"
	Message  string `json:"message"`  // 闹钟触发时的消息
}

func (a *Alarm) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	a.cfg = cfg
	a.openaiClient = openaiClient
	a.serverAddr = "localhost:8080" // 默认服务端地址
	fmt.Println("Alarm plugin initialized successfully")
	return nil
}

func (a *Alarm) ID() string {
	return "alarm"
}

func (a *Alarm) Description() string {
	return "An alarm plugin that triggers after a specified duration and reminds you of an event."
}

func (a *Alarm) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "alarm",
		Description: "Set an alarm that triggers after a specified duration and reminds you of an event.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"duration": {
					Type:        jsonschema.String,
					Description: "闹钟的持续时间，例如：'10s', '2m', '1h'。",
				},
				"event": {
					Type:        jsonschema.String,
					Description: "事件名称，例如：'会议'。",
				},
				"message": {
					Type:        jsonschema.String,
					Description: "闹钟触发时的消息。",
				},
			},
			Required: []string{"duration", "event", "message"},
		},
	}
}

func (a *Alarm) Execute(jsonInput string) (string, error) {
	var input AlarmInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", fmt.Errorf("无法解析输入数据：%v", err)
	}

	// 解析持续时间
	duration, err := time.ParseDuration(input.Duration)
	if err != nil {
		return "", fmt.Errorf("无法解析持续时间：%v", err)
	}

	fmt.Printf("Setting alarm for %v with event: %s, message: %s\n", duration, input.Event, input.Message)

	// 设置定时器
	timer := time.NewTimer(duration)

	// 使用匿名函数触发闹钟消息
	go func() {
		<-timer.C
		alarmMsg := fmt.Sprintf("Alarm triggered! Event: %s, Message: %s", input.Event, input.Message)
		fmt.Println(alarmMsg)

		// 将消息发送到主程序的TCP服务端
		sendMessageToServer(alarmMsg, a.serverAddr)
	}()

	return fmt.Sprintf("Alarm set for %v with event: %s, message: %s", duration, input.Event, input.Message), nil
}

// sendMessageToServer 通过TCP发送消息到主程序的服务端
func sendMessageToServer(msg, serverAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("无法连接到服务端：%v\n", err)
		return
	}
	defer conn.Close()

	_, err = fmt.Fprintln(conn, msg)
	if err != nil {
		fmt.Printf("发送消息到服务端失败：%v\n", err)
	}
}
