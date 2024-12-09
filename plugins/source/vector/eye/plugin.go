package main

import (
	"context"
	"encoding/json"
	"fmt"

	sdk_wrapper "github.com/fforchino/vector-go-sdk/pkg/sdk-wrapper"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

// EyeControlPlugin作为plugins.Plugin的实现
var Plugin plugins.Plugin = &EyeControlPlugin{}

// EyeControlPlugin结构体定义
type EyeControlPlugin struct {
	cfg          config.Cfg
	openaiClient *openai.Client
}

// Init方法用于初始化插件
func (e *EyeControlPlugin) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	e.cfg = cfg
	e.openaiClient = openaiClient
	return nil
}

// ID方法返回插件的唯一标识符
func (e EyeControlPlugin) ID() string {
	return "control_eye"
}

// Description方法返回插件的描述
func (e EyeControlPlugin) Description() string {
	return "控制机器人眼睛获取图片或视频。"
}

// FunctionDefinition方法返回OpenAI函数定义
func (e EyeControlPlugin) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "control_eye",
		Description: "根据指令控制机器人眼睛获取图片或视频。",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"action": {
					Type: jsonschema.String,
					Enum: []string{"photo", "video"},
				},
			},
		},
	}
}

// Execute方法执行插件的主要功能，控制手臂动作
func (e EyeControlPlugin) Execute(jsonInput string) (string, error) {
	// 解析输入
	var input struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal([]byte(jsonInput), &input); err != nil {
		return "", fmt.Errorf("无法解析输入: %v", err)
	}

	ctx := context.Background()
	start := make(chan bool)
	stop := make(chan bool)
	go func() {
		_ = sdk_wrapper.Robot.BehaviorControl(ctx, start, stop)
	}()

	for {
		select {
		case <-start:
			switch input.Action {
			case "photo":
				sdk_wrapper.SetLocale("en-US")
				sdk_wrapper.SayText("are you ok ?")
				sdk_wrapper.SaveHiResCameraPicture("robot_photo.jpg")
				fmt.Println("正在获取图片")
				stop <- true
				return fmt.Sprintf("获取图片完毕，图片名称: %s", "robot_photo.jpg"), nil
			case "video":
				stop <- true
				fmt.Println("正在获取视频")
				return fmt.Sprintf("获取视频完毕，视频名称: %s", "robot_video.jpg"), nil
			default:
				return "", fmt.Errorf("未知的动作指令: %s", input.Action)
			}
		}
	}
}
