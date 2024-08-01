package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &Tts{}

type Tts struct {
	cfg          config.Cfg
	openaiClient *openai.Client
}

type TtsInput struct {
	Text string `json:"text"` // 要转换为语音的文本
}

func (v *Tts) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	v.cfg = cfg
	v.openaiClient = openaiClient
	fmt.Println("Tts plugin initialized successfully")
	return nil
}

func (v *Tts) ID() string {
	return "tts"
}

func (v *Tts) Description() string {
	return "A tts plugin that uses OpenAI Tts Preview to generate speech from text."
}

func (v *Tts) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "tts",
		Description: "Generate speech from text using OpenAI Tts Preview.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"text": {
					Type:        jsonschema.String,
					Description: "The text to convert to speech.",
				},
			},
			Required: []string{"text"},
		},
	}
}

func (v *Tts) Execute(jsonInput string) (string, error) {
	var input TtsInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", err
	}

	fmt.Printf("Received input: Text=%s\n", input.Text)

	// Make a request to OpenAI GPT-4 Tts Preview
	res, err := v.openaiClient.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Input: input.Text,
		Voice: openai.VoiceAlloy,
	})
	if err != nil {
		return "", fmt.Errorf("CreateSpeech error: %v", err)
	}

	buf, err := io.ReadAll(res)
	if err != nil {
		return "", fmt.Errorf("ReadAll error: %v", err)
	}

	// 输出文件路径
	outputFile := "speech.mp3"
	// 检查文件是否存在
	if _, err := os.Stat(outputFile); err == nil {
		// 文件存在，尝试删除
		err := os.Remove(outputFile)
		if err != nil {
			// 删除文件时出错
			return "", fmt.Errorf("Failed to delete existing file %s: %s", outputFile, err)
		}
		fmt.Printf("Existing file %s deleted successfully.\n", outputFile)
	} else if !os.IsNotExist(err) {
		// 访问文件时出现了其他错误
		return "", fmt.Errorf("Error checking file %s: %s", outputFile, err)
	}

	// 保存 buf 到文件为 mp3
	err = os.WriteFile(outputFile, buf, 0644)
	if err != nil {
		return "", fmt.Errorf("WriteFile error: %v", err)
	}

	return "Speech generated and saved to speech.mp3", nil
}
