package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &LeftFrontalLobe{}

type LeftFrontalLobe struct {
	cfg          config.Cfg
	openaiClient *openai.Client
}

type LFLInput struct {
	Prompt string `json:"prompt"` // 要向OpenAI提问的提示语
}

func (l *LeftFrontalLobe) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	l.cfg = cfg
	l.openaiClient = openaiClient
	fmt.Println("LeftFrontalLobe plugin initialized successfully")
	return nil
}

func (l *LeftFrontalLobe) ID() string {
	return "left_frontal_lobe"
}

func (l *LeftFrontalLobe) Description() string {
	return "A plugin responsible for planning, reasoning, problem-solving, and logical thinking."
}

func (l *LeftFrontalLobe) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "left_frontal_lobe",
		Description: "Responsible for planning, reasoning, problem-solving, and logical thinking.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"prompt": {
					Type:        jsonschema.String,
					Description: "向OpenAI提问的提示语。",
				},
			},
			Required: []string{"prompt"},
		},
	}
}

func (l *LeftFrontalLobe) Execute(jsonInput string) (string, error) {
	var input LFLInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", err
	}

	fmt.Printf("Received input: Prompt=%s\n", input.Prompt)

	// 创建请求给 OpenAI
	resp, err := l.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: input.Prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %v", err)
	}

	fmt.Printf("Received response from OpenAI GPT-3.5 Turbo: %s\n", resp.Choices[0].Message.Content)

	return resp.Choices[0].Message.Content, nil
}
