package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"runtime"

	"github.com/sashabaranov/go-openai"
	config "github.com/wangergou2023/agi_modules_for_go/config"
)

// Plugin接口定义了所有插件必须实现的方法
type Plugin interface {
	Init(cfg config.Cfg, openaiClient *openai.Client) error
	ID() string
	Description() string
	FunctionDefinition() openai.FunctionDefinition
	Execute(string) (string, error)
}

// PluginResponse结构体用于封装插件执行的响应
type PluginResponse struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

// PluginManager 管理插件的加载和调用
type PluginManager struct {
	loadedPlugins map[string]Plugin
	cfg           config.Cfg
	openaiClient  *openai.Client
}

// NewPluginManager 创建一个新的PluginManager实例
func NewPluginManager(cfg config.Cfg, openaiClient *openai.Client) *PluginManager {
	return &PluginManager{
		loadedPlugins: make(map[string]Plugin),
		cfg:           cfg,
		openaiClient:  openaiClient,
	}
}

// LoadPlugins 加载指定目录下的所有插件
func (pm *PluginManager) LoadPlugins(compiledDir string) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("cannot get current file path")
	}

	files, err := os.ReadDir(filepath.Dir(filename) + "/" + compiledDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".so" {
			err := pm.loadSinglePlugin(filepath.Dir(filename) + "/" + compiledDir + "/" + file.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// loadSinglePlugin 加载单个插件
func (pm *PluginManager) loadSinglePlugin(path string) error {
	plug, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := plug.Lookup("Plugin")
	if err != nil {
		return err
	}

	p, ok := symbol.(*Plugin)
	if !ok {
		return fmt.Errorf("unexpected type from module symbol: %s", path)
	}

	err = (*p).Init(pm.cfg, pm.openaiClient)
	if err != nil {
		return err
	}

	pm.loadedPlugins[(*p).ID()] = *p
	return nil
}

// CallPlugin 通过ID查找并执行插件
func (pm *PluginManager) CallPlugin(id string, jsonInput string) (string, error) {
	response := PluginResponse{}

	plugin, exists := pm.GetPluginByID(id)
	if !exists {
		response.Error = fmt.Sprintf("plugin with ID %s not found", id)
		jsonResponse, err := json.Marshal(response)
		return string(jsonResponse), err
	}

	result, err := plugin.Execute(jsonInput)
	if err != nil {
		response.Error = err.Error()
	} else {
		response.Result = result
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("error marshaling response to JSON: %v", err)
	}

	return string(jsonResponse), nil
}

// IsPluginLoaded 检查指定ID的插件是否已加载
func (pm *PluginManager) IsPluginLoaded(id string) bool {
	_, exists := pm.loadedPlugins[id]
	return exists
}

// GetPluginByID 通过ID获取插件
func (pm *PluginManager) GetPluginByID(id string) (Plugin, bool) {
	p, exists := pm.loadedPlugins[id]
	return p, exists
}

// GetAllPlugins 返回所有已加载的插件
func (pm *PluginManager) GetAllPlugins() map[string]Plugin {
	return pm.loadedPlugins
}

// GenerateOpenAIFunctionsDefinition 生成所有插件的OpenAI函数定义
func (pm *PluginManager) GenerateOpenAIFunctionsDefinition() []openai.FunctionDefinition {
	var definitions []openai.FunctionDefinition

	for _, plugin := range pm.loadedPlugins {
		def := plugin.FunctionDefinition()
		definitions = append(definitions, def)
	}

	return definitions
}
