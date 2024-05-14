package main

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &Seat{}

type Seat struct {
	cfg          config.Cfg
	openaiClient *openai.Client
	mqttClient   mqtt.Client // MQTT客户端
	seatStatus   SeatStatus
}

type SeatInput struct {
	Command string `json:"command"` // 控制通风的命令，例如："turn_on", "turn_off", "get_status"
}

type SeatStatus struct {
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
	Ventilation string `json:"ventilation"` // 通风状态，例如："on", "off"
}

func (s *Seat) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	s.cfg = cfg
	s.openaiClient = openaiClient

	// 初始化MQTT客户端
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL()).
		SetClientID("seat_client").
		SetUsername(cfg.MQTTUsername()).
		SetPassword(cfg.MQTTPassword())

	s.mqttClient = mqtt.NewClient(opts)
	if token := s.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("无法连接到MQTT代理：%v", token.Error())
	}

	// 订阅座椅状态
	s.mqttClient.Subscribe("seat/status", 0, s.messageHandler)

	fmt.Println("Seat plugin initialized successfully")
	return nil
}

func (s *Seat) ID() string {
	return "seat"
}

func (s *Seat) Description() string {
	return "A seat plugin that can get temperature, humidity, and ventilation status, and control the ventilation."
}

func (s *Seat) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "seat",
		Description: "Control the seat ventilation and get status of temperature, humidity, and ventilation.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"command": {
					Type:        jsonschema.String,
					Description: "Control command for ventilation or get status, e.g., 'turn_on', 'turn_off', 'get_status'.",
				},
			},
			Required: []string{"command"},
		},
	}
}

func (s *Seat) Execute(jsonInput string) (string, error) {
	var input SeatInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", fmt.Errorf("无法解析输入数据：%v", err)
	}

	switch input.Command {
	case "turn_on":
		s.controlVentilation("on")
	case "turn_off":
		s.controlVentilation("off")
	case "get_status":
		statusJSON, err := json.Marshal(s.seatStatus)
		if err != nil {
			return "", fmt.Errorf("无法序列化座椅状态：%v", err)
		}
		return string(statusJSON), nil
	default:
		return "", fmt.Errorf("无效的命令：%v", input.Command)
	}

	return fmt.Sprintf("Command %s executed successfully", input.Command), nil
}

func (s *Seat) controlVentilation(state string) {
	msg := fmt.Sprintf("set_ventilation:%s", state)
	sendMessageToMQTT(msg, s.mqttClient)
	fmt.Printf("Ventilation turned %s\n", state)
}

func (s *Seat) messageHandler(client mqtt.Client, msg mqtt.Message) {
	var status SeatStatus
	err := json.Unmarshal(msg.Payload(), &status)
	if err != nil {
		fmt.Printf("无法解析座椅状态消息：%v\n", err)
		return
	}

	s.seatStatus = status
	fmt.Printf("Received seat status: %+v\n", s.seatStatus)
}

// sendMessageToMQTT 通过MQTT发送消息
func sendMessageToMQTT(msg string, mqttClient mqtt.Client) {
	topic := "seat/control"
	token := mqttClient.Publish(topic, 0, false, msg)
	token.Wait()
	if token.Error() != nil {
		fmt.Printf("发送消息到MQTT服务器失败：%v\n", token.Error())
	}
}

// 测试参考数据
// {
// 	"temperature": "22.5°C",
// 	"humidity": "45%",
// 	"ventilation": "on"
// }
