package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &RolePlayingPlugin{}

type RolePlayingPlugin struct {
	cfg           config.Cfg
	openaiClient  *openai.Client
	ScriptCatalog map[string]string // 存储剧本的目录，键为剧本的名称，值为具体剧本内容
}

type Request struct {
	RequestType string `json:"requestType"` // 请求类型，"catalog" 或 "script"
	ScriptName  string `json:"scriptName"`  // 剧本名称
}

func (rp *RolePlayingPlugin) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	// 保存配置信息和OpenAI客户端
	rp.cfg = cfg
	rp.openaiClient = openaiClient

	// 获取当前函数的执行文件路径
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Error: Cannot get current file path")
		return fmt.Errorf("cannot get current file path")
	}

	// 打印当前文件所在目录
	fmt.Println("Current file path:", filename)
	fmt.Println("Current directory:", filepath.Dir(filename))

	// 构造prompts-zh.json文件的完整路径
	jsonFilePath := filepath.Join(filepath.Dir(filename), "prompts-zh.json")
	bytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return err
	}

	var prompts []struct {
		Act    string `json:"act"`
		Prompt string `json:"prompt"`
	}
	err = json.Unmarshal(bytes, &prompts)
	if err != nil {
		return err
	}

	rp.ScriptCatalog = make(map[string]string)
	for _, p := range prompts {
		rp.ScriptCatalog[p.Act] = p.Prompt
	}

	fmt.Println("Role-playing plugin initialized successfully")
	return nil
}

func (rp *RolePlayingPlugin) Execute(jsonInput string) (string, error) {
	var request Request
	err := json.Unmarshal([]byte(jsonInput), &request)
	if err != nil {
		return "", err
	}

	switch request.RequestType {
	case "catalog":
		// 返回所有剧本的目录
		keys := make([]string, 0, len(rp.ScriptCatalog))
		for k := range rp.ScriptCatalog {
			keys = append(keys, k)
		}
		catalog, _ := json.Marshal(keys)
		return string(catalog), nil

	case "script":
		// 根据名称返回特定的剧本内容
		if script, ok := rp.ScriptCatalog[request.ScriptName]; ok {
			return script, nil
		}
		return "", fmt.Errorf("script named '%s' not found", request.ScriptName)

	default:
		return "", fmt.Errorf("unknown request type")
	}
}

func (rp *RolePlayingPlugin) ID() string {
	return "role_playing"
}

func (rp *RolePlayingPlugin) Description() string {
	return "Retrieve role-playing scenarios and scripts from the catalog."
}

func (rp *RolePlayingPlugin) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "role_playing",
		Description: "获取角色扮演剧本的目录或根据剧本名称检索剧本内容。",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"requestType": {
					Type:        jsonschema.String,
					Description: "请求类型：'catalog' 获取剧本目录；'script' 根据剧本名称获取内容。",
				},
				"scriptName": {
					Type:        jsonschema.String,
					Description: "剧本的名称。仅在'requestType'为'script'时使用。",
				},
			},
			Required: []string{"requestType"},
		},
	}
}
