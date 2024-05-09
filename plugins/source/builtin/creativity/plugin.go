package main

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &Creativity{}

type Creativity struct {
	cfg          config.Cfg
	openaiClient *openai.Client
}

type CreativityInput struct {
	Prompt string `json:"prompt"` // 要向OpenAI提问的提示语
}

func (c *Creativity) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	c.cfg = cfg
	c.openaiClient = openaiClient
	fmt.Println("Creativity plugin initialized successfully")
	return nil
}

func (c *Creativity) ID() string {
	return "creativity"
}

func (c *Creativity) Description() string {
	return "A creativity plugin that uses OpenAI GPT-3.5 Turbo to generate creative ideas."
}

func (c *Creativity) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "creativity",
		Description: "Generate creative ideas using OpenAI GPT-3.5 Turbo.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"prompt": {
					Type:        jsonschema.String,
					Description: "要向OpenAI提问的提示语。",
				},
			},
			Required: []string{"prompt"},
		},
	}
}

func (c *Creativity) Execute(jsonInput string) (string, error) {
	var input CreativityInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", err
	}

	fmt.Printf("Received input: Prompt=%s\n", input.Prompt)

	// Prepare the messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: input.Prompt,
		},
	}

	// Make a request to OpenAI GPT-3.5 Turbo
	resp, err := c.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %v", err)
	}

	fmt.Printf("Received response from OpenAI GPT-3.5 Turbo: %s\n", resp.Choices[0].Message.Content)

	return resp.Choices[0].Message.Content, nil
}
