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

var Plugin plugins.Plugin = &Legs{}

type Legs struct {
	cfg          config.Cfg
	openaiClient *openai.Client
	mqttClient   mqtt.Client // MQTT客户端
}

type LegsInput struct {
	MotorID int `json:"motor_id"` // 电机编号
	Angle   int `json:"angle"`    // 角度
}

func (f *Legs) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	f.cfg = cfg
	f.openaiClient = openaiClient

	// 初始化MQTT客户端
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL()).
		SetClientID("motor_client").
		SetUsername(cfg.MQTTUsername()).
		SetPassword(cfg.MQTTPassword())

	f.mqttClient = mqtt.NewClient(opts)
	if token := f.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("无法连接到MQTT代理：%v", token.Error())
	}

	// 订阅电机状态
	f.mqttClient.Subscribe("motor/status", 0, f.messageHandler)

	fmt.Println("Legs plugin initialized successfully")
	return nil
}

func (f *Legs) ID() string {
	return "legs"
}

func (f *Legs) Description() string {
	return "A legs plugin that can control the motors of legs."
}

func (f *Legs) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "legs",
		Description: "Control the motors of the legs.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"motor_id": {
					Type:        jsonschema.Integer,
					Description: "ID of the motor to control 0~3.",
				},
				"angle": {
					Type:        jsonschema.Integer,
					Description: "Angle to set the motor to 0~180°.",
				},
			},
			Required: []string{"motor_id", "angle"},
		},
	}
}

func (f *Legs) Execute(jsonInput string) (string, error) {
	var input LegsInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", fmt.Errorf("无法解析输入数据：%v", err)
	}

	f.controlMotor(input.MotorID, input.Angle)
	return fmt.Sprintf("Motor %d set to angle %d successfully", input.MotorID, input.Angle), nil
}

// controlMotor 发布电机控制消息到MQTT服务器
func (f *Legs) controlMotor(motorID int, angle int) {
	msg := fmt.Sprintf("%d:%d", motorID, angle)
	sendMessageToMQTT(msg, f.mqttClient)
	fmt.Printf("Motor %d set to %d\n", motorID, angle)
}

// messageHandler 处理接收到的MQTT消息并更新电机状态
func (f *Legs) messageHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received motor status: %s \n", msg.Payload())
}

// sendMessageToMQTT 通过MQTT发送消息
func sendMessageToMQTT(msg string, mqttClient mqtt.Client) {
	topic := "motor/control"
	token := mqttClient.Publish(topic, 0, false, msg)
	token.Wait()
	if token.Error() != nil {
		fmt.Printf("发送消息到MQTT服务器失败：%v\n", token.Error())
	}
}

// 测试参考数据
// {
// 	"motor_id": 0,
// 	"angle": 90
// }
