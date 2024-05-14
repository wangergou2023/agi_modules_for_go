package main

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &Alarm{}

type Alarm struct {
	cfg          config.Cfg
	openaiClient *openai.Client
	mqttClient   mqtt.Client // MQTT客户端
}

type AlarmInput struct {
	Duration string `json:"duration"` // 闹钟的持续时间，例如："10s", "2m"
	Event    string `json:"event"`    // 事件名称，例如："会议"
	Message  string `json:"message"`  // 闹钟触发时的消息
}

func (a *Alarm) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	a.cfg = cfg
	a.openaiClient = openaiClient

	// 初始化MQTT客户端
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL()).
		SetClientID("alarm_client").
		SetUsername(cfg.MQTTUsername()).
		SetPassword(cfg.MQTTPassword())

	a.mqttClient = mqtt.NewClient(opts)
	if token := a.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("无法连接到MQTT代理：%v", token.Error())
	}

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

		// 将消息发送到MQTT服务器
		sendMessageToMQTT(alarmMsg, a.mqttClient)
	}()

	return fmt.Sprintf("Alarm set for %v with event: %s, message: %s", duration, input.Event, input.Message), nil
}

// sendMessageToMQTT 通过MQTT发送消息
func sendMessageToMQTT(msg string, mqttClient mqtt.Client) {
	topic := "plugin/messages"
	token := mqttClient.Publish(topic, 0, false, msg)
	token.Wait()
	if token.Error() != nil {
		fmt.Printf("发送消息到MQTT服务器失败：%v\n", token.Error())
	}
}
