package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

// 声明CommandPlugin作为plugins.Plugin的实现
var Plugin plugins.Plugin = &CommandPlugin{}

// CommandPlugin结构体定义
type CommandPlugin struct {
	cfg          config.Cfg
	openaiClient *openai.Client
}

// Init方法用于初始化插件
func (c *CommandPlugin) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	c.cfg = cfg
	c.openaiClient = openaiClient
	return nil
}

// ID方法返回插件的唯一标识符
func (c CommandPlugin) ID() string {
	return "command"
}

// Description方法返回插件的描述
func (c CommandPlugin) Description() string {
	return "执行不需要交互的Linux命令。"
}

// FunctionDefinition方法返回OpenAI函数定义
func (c CommandPlugin) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "command",
		Description: "执行指定的Linux命令并返回结果。注意：该命令必须是不需要用户交互的。",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"command": {
					Type:        jsonschema.String,
					Description: "要执行的不需要交互的命令",
				},
			},
			Required: []string{"command"},
		},
	}
}

// Execute方法执行插件的主要功能，执行指定命令
func (c CommandPlugin) Execute(jsonInput string) (string, error) {
	// 解析输入参数
	var input struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(jsonInput), &input); err != nil {
		return "", fmt.Errorf("输入解析错误: %v", err)
	}

	// 执行命令
	cmd := exec.Command("bash", "-c", input.Command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("命令执行错误: %v", err)
	}

	// 返回命令输出
	return out.String(), nil
}

func main() {
	// 示例：如何初始化和使用CommandPlugin
	var cfg config.Cfg
	var openaiClient *openai.Client

	// 初始化插件
	plugin := &CommandPlugin{}
	err := plugin.Init(cfg, openaiClient)
	if err != nil {
		fmt.Println("插件初始化失败:", err)
		return
	}

	// 执行插件功能
	command := "date" // 你可以替换成任何其他不需要交互的Linux命令
	jsonInput := fmt.Sprintf(`{"command": "%s"}`, command)
	result, err := plugin.Execute(jsonInput)
	if err != nil {
		fmt.Println("执行失败:", err)
		return
	}

	// 打印结果
	fmt.Println("命令输出:", result)
}
