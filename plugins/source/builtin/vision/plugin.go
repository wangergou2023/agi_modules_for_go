package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &Vision{}

type Vision struct {
	cfg          config.Cfg
	openaiClient *openai.Client
}

type VisionInput struct {
	ImagePath string `json:"imagePath"` // 本地图片路径
	MimeType  string `json:"mimeType"`  // 图片的MIME类型，例如：image/jpeg
	Prompt    string `json:"prompt"`    // 要向OpenAI提问的提示语
}

func (v *Vision) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	v.cfg = cfg
	v.openaiClient = openaiClient
	fmt.Println("Vision plugin initialized successfully")
	return nil
}

func (v *Vision) ID() string {
	return "vision"
}

func (v *Vision) Description() string {
	return "A vision plugin that uses OpenAI GPT-4 Vision Preview to analyze images."
}

func (v *Vision) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "vision",
		Description: "Analyze images using OpenAI GPT-4 Vision Preview.",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"imagePath": {
					Type:        jsonschema.String,
					Description: "本地图片的文件路径。",
				},
				"mimeType": {
					Type:        jsonschema.String,
					Description: "图片的MIME类型，例如：image/jpeg。",
				},
				"prompt": {
					Type:        jsonschema.String,
					Description: "向OpenAI提问的提示语。",
				},
			},
			Required: []string{"imagePath", "mimeType", "prompt"},
		},
	}
}

func (v *Vision) Execute(jsonInput string) (string, error) {
	var input VisionInput
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", err
	}

	fmt.Printf("Received input: ImagePath=%s, MimeType=%s, Prompt=%s\n", input.ImagePath, input.MimeType, input.Prompt)

	// Encode the image to base64
	base64Image, err := encodeImageToBase64(input.ImagePath)
	if err != nil {
		return "", err
	}

	// Create a data URL for the image
	dataURL := createDataURL(base64Image, input.MimeType)
	// fmt.Printf("Generated data URL: %s\n", dataURL)

	// Prepare the messages
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: input.Prompt,
				},
			},
		},
	}

	// Add the image information
	messages[0].MultiContent = append(messages[0].MultiContent, openai.ChatMessagePart{
		Type: openai.ChatMessagePartTypeImageURL,
		ImageURL: &openai.ChatMessageImageURL{
			URL:    dataURL,
			Detail: openai.ImageURLDetailAuto,
		},
	})

	// Make a request to OpenAI GPT-4 Vision Preview
	resp, err := v.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			MaxTokens: 300,
			Model:     openai.GPT4VisionPreview,
			Messages:  messages,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %v", err)
	}

	fmt.Printf("Received response from OpenAI GPT-4 Vision: %s\n", resp.Choices[0].Message.Content)

	return resp.Choices[0].Message.Content, nil
}

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
