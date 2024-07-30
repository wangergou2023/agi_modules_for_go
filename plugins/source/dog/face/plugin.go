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

var Plugin plugins.Plugin = &Face{}

type Face struct {
	cfg          config.Cfg
	openaiClient *openai.Client
	mqttClient   mqtt.Client // MQTT客户端
}

type FaceInput struct {
	Emotion string `json:"emotion"` // 表情控制命令，例如："Normal", "Angry", "Happy"
}

func (f *Face) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	f.cfg = cfg
	f.openaiClient = openaiClient

	// 初始化MQTT客户端
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL()).
		SetClientID("face_client").
		SetUsername(cfg.MQTTUsername()).
		SetPassword(cfg.MQTTPassword())

	f.mqttClient = mqtt.NewClient(opts)
	if token := f.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("无法连接到MQTT代理：%v", token.Error())
	}

	// 订阅表情状态
	f.mqttClient.Subscribe("emotion/status", 0, f.messageHandler)

	fmt.Println("Face plugin initialized successfully")
	return nil
}

func (f *Face) ID() string {
	return "face"
}

func (f *Face) Description() string {
	return "A face plugin that can control the expression of an ESP8266-based face."
}

func (f *Face) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "face",
		Description: "Control the face expression.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"emotion": {
					Type:        jsonschema.String,
					Description: "Emotion command for face control, e.g., 'Normal', 'Angry', 'Happy', 'Glee', 'Sad', 'Worried', 'Focused', 'Annoyed', 'Surprised', 'Skeptic', 'Frustrated', 'Unimpressed', 'Sleepy', 'Suspicious', 'Squint', 'Furious', 'Scared', 'Awe'.",
				},
			},
			Required: []string{"emotion"},
		},
	}
}

func (f *Face) Execute(jsonInput string) (string, error) {
	var input FaceInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", fmt.Errorf("无法解析输入数据：%v", err)
	}

	f.controlEmotion(input.Emotion)
	return fmt.Sprintf("Emotion %s executed successfully", input.Emotion), nil
}

// controlEmotion 发布表情控制消息到MQTT服务器
func (f *Face) controlEmotion(emotion string) {
	msg := fmt.Sprintf("%s", emotion)
	sendMessageToMQTT(msg, f.mqttClient)
	fmt.Printf("Emotion set to %s\n", emotion)
}

// messageHandler 处理接收到的MQTT消息并更新表情状态
func (f *Face) messageHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received face status: %s \n", msg.Payload())
}

// sendMessageToMQTT 通过MQTT发送消息
func sendMessageToMQTT(msg string, mqttClient mqtt.Client) {
	topic := "emotion/control"
	token := mqttClient.Publish(topic, 0, false, msg)
	token.Wait()
	if token.Error() != nil {
		fmt.Printf("发送消息到MQTT服务器失败：%v\n", token.Error())
	}
}

// 测试参考数据
// {
// 	"emotion/control": "Happy"
// }
