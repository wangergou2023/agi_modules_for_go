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

var Plugin plugins.Plugin = &RightFrontalLobe{}

type RightFrontalLobe struct {
	cfg          config.Cfg
	openaiClient *openai.Client
}

type RFLInput struct {
	Prompt string `json:"prompt"` // 要向OpenAI提问的提示语
}

func (r *RightFrontalLobe) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	r.cfg = cfg
	r.openaiClient = openaiClient
	fmt.Println("RightFrontalLobe plugin initialized successfully")
	return nil
}

func (r *RightFrontalLobe) ID() string {
	return "right_frontal_lobe"
}

func (r *RightFrontalLobe) Description() string {
	return "A plugin responsible for creative thinking and generating novel ideas."
}

func (r *RightFrontalLobe) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "right_frontal_lobe",
		Description: "Responsible for creative thinking and generating novel ideas.",
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

func (r *RightFrontalLobe) Execute(jsonInput string) (string, error) {
	var input RFLInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", err
	}

	fmt.Printf("Received input: Prompt=%s\n", input.Prompt)

	// 创建请求给 OpenAI
	resp, err := r.openaiClient.CreateChatCompletion(
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
