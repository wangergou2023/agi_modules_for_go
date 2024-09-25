package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
	config := openai.DefaultConfig("your token")
	//need"/v1"
	config.BaseURL = "your url/v1"
	client := openai.NewClientWithConfig(config)
	// 初始化上下文，用于控制 API 请求生命周期
	ctx := context.Background()

	// 定义用于存储 API 返回结果的结构体
	type Result struct {
		YourName  string `json:"your_name"` // 自己的名字
		Responses []struct {
			TargetName string `json:"target_name"` // 打招呼对象的名字
			Message    string `json:"message"`     // 消息内容
		} `json:"responses"`
	}

	var result Result

	// 生成与 Result 结构体对应的 JSON Schema
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		log.Fatalf("生成 JSON Schema 错误: %v", err)
	}

	// 创建一个聊天请求，发送给 OpenAI API
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini, // 使用 GPT-4 轻量模型
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一个群聊助手，你的名字是小白。你的职责是帮助主人传递消息，并通知群里的其他成员，例如小明和小红。请确保正确传达消息，并在必要时回应。",
			},
			{
				Role: openai.ChatMessageRoleUser,
				// Content: "主人：小白请通知小明明天来我家吃饭",
				Content: "主人：小白请给大家打招呼",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "小明：你好我叫小明",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "小红：你好我叫小红",
			},
		},
		// 将期望的响应格式设置为 JSON Schema
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema, // 返回 JSON Schema 格式
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "responses", // 定义 schema 名称
				Schema: schema,      // 使用之前生成的 JSON Schema
				Strict: true,        // 严格匹配 schema
			},
		},
	})

	// 检查请求是否出错
	if err != nil {
		log.Fatalf("创建聊天请求错误: %v", err)
	}

	fmt.Println(resp.Choices[0].Message.Content)

	// 将 API 返回的内容反序列化为我们定义的 Result 结构体
	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		log.Fatalf("解析 JSON Schema 错误: %v", err)
	}

	// 打印输出解析后的结果
	fmt.Println(result)
}
