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
		Steps []struct {
			Explanation string `json:"explanation"` // 每一步的解释
			Output      string `json:"output"`      // 每一步的输出结果
		} `json:"steps"`
		FinalAnswer string `json:"final_answer"` // 最终的答案
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
				Content: "你是一位数学老师，请逐步引导用户解决问题。", // 系统角色：提供中文指导
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "如何求解 8x + 7 = -23？", // 用户输入的数学问题
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema, // 请求返回 JSON Schema 格式的结果
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "math_reasoning", // 定义 schema 名称
				Schema: schema,           // 使用之前生成的 JSON Schema
				Strict: true,             // 严格匹配 schema
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
