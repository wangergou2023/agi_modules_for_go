package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"encoding/json"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

var Plugin plugins.Plugin = &WeatherPlugin{}

type WeatherPlugin struct {
	cfg          config.Cfg
	openaiClient *openai.Client
}

func (w *WeatherPlugin) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	w.cfg = cfg
	w.openaiClient = openaiClient
	return nil
}

func (w WeatherPlugin) ID() string {
	return "weather"
}

func (w WeatherPlugin) Description() string {
	return "获取指定地点的当前天气情况。"
}

func (w WeatherPlugin) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "weather",
		Description: "返回指定地点的当前天气情况。",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"location": {
					Type:        jsonschema.String,
					Description: "查询天气的地点。",
				},
			},
			Required: []string{"location"},
		},
	}
}

func (w WeatherPlugin) Execute(jsonInput string) (string, error) {
	var input struct {
		Location string `json:"location"`
	}
	err := json.Unmarshal([]byte(jsonInput), &input)
	if err != nil {
		return "", err
	}

	weatherInfo, err := w.getWeather(input.Location)
	if err != nil {
		return "", err
	}

	return weatherInfo, nil
}

func (w WeatherPlugin) getWeather(location string) (string, error) {
	// 使用 wttr.in 获取天气信息
	url := fmt.Sprintf("http://wttr.in/%s?format=3", location)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	weatherInfo := strings.TrimSpace(string(body))
	return weatherInfo, nil
}
