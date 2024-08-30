package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

// 声明JSONPlugin作为plugins.Plugin的实现
var Plugin plugins.Plugin = &JSONPlugin{}

// QAEntry结构体定义，用于存储每个条目的详细信息
type QAEntry struct {
	PluginName   string `json:"plugin_name"`
	UsageMethod  string `json:"usage_method"`
	InputParams  string `json:"input_params"`
	OutputResult string `json:"output_result"`
	Timestamp    string `json:"timestamp"`
}

// JSONPlugin结构体定义
type JSONPlugin struct {
	cfg      config.Cfg
	filePath string               // JSON文件路径
	store    map[string][]QAEntry // 用于存储问题和解决方法的数据
}

// Init方法用于初始化插件
func (j *JSONPlugin) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	j.cfg = cfg
	j.filePath = "qa_data.json" // 默认的JSON文件路径

	// 加载数据
	if err := j.loadFromFile(); err != nil {
		return fmt.Errorf("加载JSON文件失败: %v", err)
	}
	return nil
}

// 从JSON文件加载数据
func (j *JSONPlugin) loadFromFile() error {
	file, err := os.Open(j.filePath)
	if os.IsNotExist(err) {
		// 如果文件不存在，则创建一个新文件并初始化一个空存储
		j.store = make(map[string][]QAEntry)
		return j.saveToFile()
	} else if err != nil {
		return err
	}
	defer file.Close()

	// 读取文件并解析为map
	bytes, err := os.ReadFile(j.filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &j.store); err != nil {
		return err
	}
	return nil
}

// 将数据保存到JSON文件
func (j *JSONPlugin) saveToFile() error {
	bytes, err := json.MarshalIndent(j.store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(j.filePath, bytes, 0644)
}

// ID方法返回插件的唯一标识符
func (j JSONPlugin) ID() string {
	return "qa_store"
}

// Description方法返回插件的描述
func (j JSONPlugin) Description() string {
	return "这个插件用于存储和检索问题及其解决方法，并将其保存到文件中。"
}

// FunctionDefinition方法返回OpenAI函数定义
func (j JSONPlugin) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "qa_store",
		Description: "存储或检索问题及其解决方法。",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"action": {
					Type:        jsonschema.String,
					Description: "要执行的操作：'add', 'get', 'delete', 或 'update'",
				},
				"question": {
					Type:        jsonschema.String,
					Description: "要添加、检索或删除的问题",
				},
				"solution": {
					Type:        jsonschema.String,
					Description: "问题的解决方法，仅在'add'或'update'操作时需要",
				},
				"command": {
					Type:        jsonschema.String,
					Description: "获取结果时使用的具体命令，仅在'add'或'update'操作时需要",
				},
			},
			Required: []string{"action", "question"},
		},
	}
}

// Execute方法执行插件的主要功能，根据操作存储或检索问题及其解决方法
func (j *JSONPlugin) Execute(jsonInput string) (string, error) {
	var input struct {
		Action   string `json:"action"`
		Question string `json:"question"`
		Solution string `json:"solution,omitempty"`
		Command  string `json:"command,omitempty"` // 增加command字段来记录具体的命令调用
	}

	if err := json.Unmarshal([]byte(jsonInput), &input); err != nil {
		return "", fmt.Errorf("输入解析错误: %v", err)
	}

	// 检查JSON文件中是否已经存在相应问题
	if entries, exists := j.store[input.Question]; exists {
		switch input.Action {
		case "get":
			latestEntry := entries[len(entries)-1] // 获取最新的解决方法
			return fmt.Sprintf("解决方法: %s", latestEntry.OutputResult), nil
		case "add":
			return "", fmt.Errorf("问题 '%s' 已存在。使用 'update' 操作来更新解决方法。", input.Question)
		case "update":
			// 继续进行更新操作
		case "delete":
			// 继续进行删除操作
		}
	}

	// 如果执行成功才记录
	switch input.Action {
	case "add":
		// 添加新问题
		entry := QAEntry{
			PluginName:   "qa_store",
			UsageMethod:  "add",
			InputParams:  "使用 'curl' 命令行工具请求 wttr.in 网站并指定查询参数，如：curl -s 'http://wttr.in/{城市名}?format=3'", // 通用查询天气的方法
			OutputResult: "使用此方法可以查询任何城市的当前天气信息。",
			Timestamp:    time.Now().Format(time.RFC3339),
		}
		j.store[input.Question] = append(j.store[input.Question], entry)

	case "update":
		// 更新已有问题
		entry := QAEntry{
			PluginName:   "qa_store",
			UsageMethod:  "update",
			InputParams:  "使用 'curl' 命令行工具请求 wttr.in 网站并指定查询参数，如：curl -s 'http://wttr.in/{城市名}?format=3'", // 通用查询天气的方法
			OutputResult: "使用此方法可以查询任何城市的当前天气信息。",
			Timestamp:    time.Now().Format(time.RFC3339),
		}
		j.store[input.Question] = append(j.store[input.Question], entry)

	case "delete":
		// 删除已有问题
		if _, exists := j.store[input.Question]; exists {
			delete(j.store, input.Question)
		}
	}

	// 成功执行后保存到文件
	if err := j.saveToFile(); err != nil {
		return "", fmt.Errorf("保存数据到文件时出错: %v", err)
	}

	return fmt.Sprintf("操作成功：'%s'", input.Question), nil
}

func main() {
	// 示例：如何初始化和使用JSONPlugin
	var cfg config.Cfg
	var openaiClient *openai.Client

	// 初始化插件
	plugin := &JSONPlugin{}
	err := plugin.Init(cfg, openaiClient)
	if err != nil {
		fmt.Println("插件初始化失败:", err)
		return
	}

	// 示例操作
	addInput := `{"action": "add", "question": "如何查询天气", "solution": "使用此方法可以查询任何城市的当前天气信息。", "command": "curl -s 'http://wttr.in/{城市名}?format=3'"}`
	getInput := `{"action": "get", "question": "如何查询天气"}`
	updateInput := `{"action": "update", "question": "如何查询天气", "solution": "使用此方法可以查询任何城市的当前天气信息。", "command": "curl -s 'http://wttr.in/{城市名}?format=3'"}`
	deleteInput := `{"action": "delete", "question": "如何查询天气"}`

	// 执行“add”操作
	if result, err := plugin.Execute(addInput); err == nil {
		fmt.Println(result)
	} else {
		fmt.Println("执行失败:", err)
	}

	// 执行“get”操作
	if result, err := plugin.Execute(getInput); err == nil {
		fmt.Println("获取的解决方法:", result)
	} else {
		fmt.Println("执行失败:", err)
	}

	// 执行“update”操作
	if result, err := plugin.Execute(updateInput); err == nil {
		fmt.Println(result)
	} else {
		fmt.Println("执行失败:", err)
	}

	// 执行“delete”操作
	if result, err := plugin.Execute(deleteInput); err == nil {
		fmt.Println(result)
	} else {
		fmt.Println("执行失败:", err)
	}
}
