package main

import (
	"context" // 用于创建和管理跨API调用的上下文
	"log"     // 用于记录日志信息

	openai "github.com/sashabaranov/go-openai" // OpenAI Go SDK
)

func main() {
	client := openai.NewClient("your-token") // 使用你的OpenAI API令牌初始化客户端

	// 为用户查询创建一个嵌入请求
	queryReq := openai.EmbeddingRequest{
		Input: []string{"苹果"},        // 用户查询文本
		Model: openai.AdaEmbeddingV2, // 使用的模型版本
	}

	// 为用户查询创建嵌入
	queryResponse, err := client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		log.Fatal("Error creating query embedding:", err) // 如果创建嵌入失败，记录错误并终止程序
	}

	// 为目标文本创建一个嵌入请求
	targetReq := openai.EmbeddingRequest{
		Input: []string{"电脑"},        // 目标文本
		Model: openai.AdaEmbeddingV2, // 使用的模型版本
	}

	// 为目标文本创建嵌入
	targetResponse, err := client.CreateEmbeddings(context.Background(), targetReq)
	if err != nil {
		log.Fatal("Error creating target embedding:", err) // 如果创建嵌入失败，记录错误并终止程序
	}

	// 现在我们有了用户查询和目标文本的嵌入，我们可以计算它们的相似度
	queryEmbedding := queryResponse.Data[0]   // 获取用户查询的嵌入
	targetEmbedding := targetResponse.Data[0] // 获取目标文本的嵌入

	// 计算点积来评估相似度
	similarity, err := queryEmbedding.DotProduct(&targetEmbedding)
	if err != nil {
		log.Fatal("Error calculating dot product:", err) // 如果计算点积失败，记录错误并终止程序
	}

	// 打印出查询和目标之间的相似度分数
	log.Printf("The similarity score between the query and the target is %f", similarity)
}
