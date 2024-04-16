package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

// encodeImageToBase64 reads an image file and returns a base64 encoded string
func encodeImageToBase64(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the file content into a byte slice
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return "", err
	}

	// Encode the file content to base64
	encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encodedString, nil
}

// createDataURL creates a data URL from the base64 encoded image
func createDataURL(base64Image string, mimeType string) string {
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)
}

func main() {
	// 是否上传图片标志位
	uploadImage := false // 根据需要设置为true或false

	// Path to your local image file
	imagePath := "./image.jpg"

	// Encode the image to base64
	base64Image, err := encodeImageToBase64(imagePath)
	if err != nil {
		log.Fatalf("Failed to encode image: %v", err)
	}

	// Create a data URL for the image
	dataURL := createDataURL(base64Image, "image/jpeg")

	// Print the data URL
	fmt.Println(dataURL)

	client := openai.NewClient("your token")

	// Prepare the messages
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: "图片中你看到了什么?",
				},
			},
		},
	}

	// If uploadImage is true, add the image information
	if uploadImage {
		messages[0].MultiContent = append(messages[0].MultiContent, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				//1、图片链接
				// URL:    "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg",
				//2、本地图片
				URL:    dataURL,
				Detail: openai.ImageURLDetailAuto,
			},
		})
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			MaxTokens: 300,
			Model:     openai.GPT4VisionPreview,
			Messages:  messages,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
