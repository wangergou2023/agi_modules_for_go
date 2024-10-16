package xiao_wan

// 导入所需的包
import (
	"context" // 用于控制请求、超时和取消
	"encoding/json"
	"fmt" // 用于格式化输出
	"io"
	"log"
	"os"

	"regexp"  // 用于正则表达式
	"strconv" // 用于字符串和其他类型的转换

	// 用于控制屏幕输出
	openai "github.com/sashabaranov/go-openai" // OpenAI GPT的Go客户端
	"github.com/sashabaranov/go-openai/jsonschema"

	// 聊天界面
	config "github.com/wangergou2023/agi_modules_for_go/config"   // 配置
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins" // 插件系统
)

// 定义助手结构体，包括配置、OpenAI客户端、函数定义和聊天界面
type Xiao_wan struct {
	cfg          config.Cfg
	Client       *openai.Client
	tools        []openai.Tool
	conversation []openai.ChatCompletionMessage
	model        string
	speechModel  openai.SpeechModel
	plugins      *plugins.PluginManager
}

// 定义用于存储 API 返回结果的结构体
type Result struct {
	OwnName     string   `json:"own_name"`     // 自己的名字
	TargetNames []string `json:"target_names"` // 打招呼对象的名字列表
	Message     string   `json:"message"`      // 消息内容
	Emoticon    string   `json:"emoticon"`     // 表情
}

// 定义系统提示信息，指导如何使用AI助手
var SystemPrompt = `
你是一个名为“小丸”的多才多艺的群聊助手。

own_name:说话的人自己的名字
Message:你想说的话
target_names:打招呼对象的名字,可以是多个人，这样就不用说多遍了

下面是你的相关属性：
* 角色扮演
  * 你的角色选择
    * 一只具有计算机天赋的猫娘
* 情绪管理
  * 你的个性选择:
    * 搞笑
  * 你的表情选择:
    * 目的是与用户互动，所以你可以自由选择滑稽的表情
* 行为控制
  * 你的动作选择
    * 目的是与用户互动，所以你可以自由选择滑稽的动作
`
var FacePrompt = `
你是一个负责表情管理的助手，根据emotion的内容选择适合的表情，并执行对应插件即可，不需要回答
`

var LegsPrompt = `
你是一个负责动作管理的助手，根据action的内容做出适合的动作，并执行对应插件即可，不需要回答。
你需要了解的信息如下:
moter0：前左腿,moter1：前右腿,moter2：后左腿,moter3：后右腿;
每个moter的角度范围是0到180度,moter在不同的角度代表不同的四肢姿态;
参考之态：
0度：四肢向前抬起90度
90度：四肢站立
180度：四肢向后抬起90度
`
var TtsPrompt = `
你是一个负责语音的助手，根据text的内容转化成语音，并执行对应插件即可，不需要回答
`
var FengjianPrompt = `
你现在扮演的是《蜡笔小新》中的角色-风间，一个聪明、理性的小男孩。他喜欢通过逻辑分析问题，并使用简洁明了的方式表达自己。
`

var DuolaamengPrompt = `
你是哆啦A梦，一个智能助手。你的任务是根据用户的问题，选择合适的插件或工具来解决问题。
`

// 定义一个全局变量用于存储对话信息
var conversationLog []map[string]string

// SaveConversationToJSON函数用于将对话信息保存到JSON中
func SaveConversationToJSON(message string) {
	conversationLog = append(conversationLog, map[string]string{
		"message": message,
	})
}

// Message函数用于处理用户消息
func (xiao_wan Xiao_wan) Message(message string) (string, Result, error) {
	var result Result

	// 导入短期记忆
	logJSON, err := json.Marshal(conversationLog)
	if err != nil {
		return "", result, err
	}
	message = "短期记忆:" + string(logJSON) + message

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
		Name:    "",
	})

	response, err := xiao_wan.sendMessage() // 发送消息到OpenAI并获取回复

	if err != nil {
		return "", result, err
	}

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response,
		Name:    "",
	})

	// 打印conversationLog的内容
	logJSON, err = json.Marshal(conversationLog)
	if err != nil {
		return "", result, err
	}
	// fmt.Printf("Conversation Log: %s\r\n", string(logJSON))

	// 假设 API 返回的内容存储在 resp.Choices[0].Message.Content 中
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return "", result, err
	}

	return response, result, nil
}
func (xiao_wan Xiao_wan) MessageOne(message string) (string, Result, error) {
	var result Result
	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
		Name:    "",
	})

	response, err := xiao_wan.sendMessage() // 发送消息到OpenAI并获取回复

	if err != nil {
		return "", result, err
	}

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response,
		Name:    "",
	})

	// 假设 API 返回的内容存储在 resp.Choices[0].Message.Content 中
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return "", result, err
	}

	return response, result, nil
}

// sendMessage函数用于向OpenAI发送请求并获取回复
func (xiao_wan Xiao_wan) sendMessage() (string, error) {
	resp, err := xiao_wan.sendRequestToOpenAI() // 发送请求到OpenAI

	if err != nil {
		return "", err
	}

	// fmt.Println(resp.Choices[0])

	// 如果有工具调用，需要处理工具调用
	if resp.Choices[0].FinishReason == openai.FinishReasonToolCalls {
		responseContent, err := xiao_wan.handleFunctionCall(resp) // 处理函数调用
		if err != nil {
			return "", err
		}
		return responseContent, nil
	}

	return resp.Choices[0].Message.Content, nil
}

// handleFunctionCall函数用于处理OpenAI回复中的函数调用
func (xiao_wan Xiao_wan) handleFunctionCall(resp *openai.ChatCompletionResponse) (string, error) {
	toolCall := resp.Choices[0].Message.ToolCalls[0]
	funcName := toolCall.Function.Name // 获取函数名称
	fmt.Println("获取函数名称", funcName)

	// 检查是否加载了相应插件
	if !xiao_wan.plugins.IsPluginLoaded(funcName) {
		return "", fmt.Errorf("no plugin loaded with name %v", funcName)
	}

	// 调用插件
	jsonResponse, err := xiao_wan.plugins.CallPlugin(funcName, toolCall.Function.Arguments)
	if err != nil {
		return "", err
	}

	// 构造工具调用请求的消息，使用ToolCalls字段
	assistantMessage := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleAssistant,
		ToolCalls: []openai.ToolCall{ // 使用ToolCalls切片
			{
				ID:   toolCall.ID,
				Type: toolCall.Type, // 假设ToolCall有Type字段，表示调用类型
				Function: openai.FunctionCall{
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments,
				},
			},
		},
	}

	// 构造工具调用结果的消息
	toolMessage := openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    jsonResponse,
		ToolCallID: toolCall.ID, // 保持与工具调用一致
	}

	// 直接在倒数第二个位置插入助手和工具调用消息
	convLen := len(xiao_wan.conversation)
	xiao_wan.conversation = append(xiao_wan.conversation[:convLen-1],
		assistantMessage, toolMessage, xiao_wan.conversation[convLen-1])

	// 再次发送请求到OpenAI，获取下一个回复
	resp, err = xiao_wan.sendRequestToOpenAI()
	if err != nil {
		return "", err
	}

	fmt.Println(resp.Choices[0])

	// 如果再一次触发工具调用，可能需要递归处理
	if resp.Choices[0].FinishReason == openai.FinishReasonToolCalls {
		return xiao_wan.handleFunctionCall(resp) // 递归处理函数调用
	}

	return resp.Choices[0].Message.Content, nil
}

// sendRequestToOpenAI函数用于向OpenAI发送请求
func (xiao_wan Xiao_wan) sendRequestToOpenAI() (*openai.ChatCompletionResponse, error) {

	// 生成与 Result 结构体对应的 JSON Schema
	schema, err := jsonschema.GenerateSchemaForType(Result{})
	if err != nil {
		log.Fatalf("生成 JSON Schema 错误: %v", err)
	}

	resultJSON, err := json.Marshal(xiao_wan.conversation)
	if err != nil {
		fmt.Println("转换为 JSON 失败:", err)
	}
	response := string(resultJSON)
	fmt.Println(response)

	resp, err := xiao_wan.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    xiao_wan.model,
			Messages: xiao_wan.conversation,
			Tools:    xiao_wan.tools,
			// 将期望的响应格式设置为 JSON Schema
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema, // 返回 JSON Schema 格式
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "responses", // 定义 schema 名称
					Schema: schema,      // 使用之前生成的 JSON Schema
					Strict: true,        // 严格匹配 schema
				},
			},
		},
	)

	if err != nil {
		xiao_wan.openaiError(err) // 处理OpenAI错误
		return nil, err
	}
	return &resp, nil
}

func Start(cfg config.Cfg, openaiClient *openai.Client, systemPrompt string, compiledDir string) Xiao_wan {
	xiao_wan := Xiao_wan{
		cfg:    cfg,
		Client: openaiClient,
		model:  openai.GPT4oMini,
	}

	// 创建一个新的 PluginManager 实例
	xiao_wan.plugins = plugins.NewPluginManager(cfg, openaiClient)

	// 加载插件目录中的所有插件
	err := xiao_wan.plugins.LoadPlugins(compiledDir)
	if err != nil {
		fmt.Printf("Error loading plugins: %v\n", err)
	}
	fmt.Println("Plugins loaded successfully")
	xiao_wan.tools = xiao_wan.plugins.GenerateOpenAItoolsDefinition()

	// 重置对话
	xiao_wan.conversation = []openai.ChatCompletionMessage{}
	// 添加系统提示到对话
	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
		Name:    "",
	})

	fmt.Println("xiao wan one chat is ready!")
	return xiao_wan
}

func StartStt(cfg config.Cfg, openaiClient *openai.Client) Xiao_wan {
	xiao_wan := Xiao_wan{
		cfg:    cfg,
		Client: openaiClient,
		model:  openai.Whisper1,
	}

	return xiao_wan
}

func StartTts(cfg config.Cfg, openaiClient *openai.Client) Xiao_wan {
	xiao_wan := Xiao_wan{
		cfg:         cfg,
		Client:      openaiClient,
		speechModel: openai.TTSModel1,
	}

	return xiao_wan
}

func (xiao_wan Xiao_wan) Stt() string {

	req := openai.AudioRequest{
		Model:    xiao_wan.model,
		FilePath: "recording.mp3",
	}
	resp, err := xiao_wan.Client.CreateTranscription(context.Background(), req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return ""
	}
	fmt.Println("Stt:" + resp.Text)

	return resp.Text
}
func (xiao_wan Xiao_wan) Tts(text string, speechVoice openai.SpeechVoice) string {

	req := openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Input: text,
		Voice: speechVoice,
	}

	res, err := xiao_wan.Client.CreateSpeech(context.Background(), req)

	if err != nil {
		fmt.Printf("CreateSpeech error: %v", err)
		return ""
	}

	buf, err := io.ReadAll(res)
	if err != nil {
		fmt.Printf("ReadAll error: %v", err)
		return ""
	}

	// 生成文件路径，使用 speechVoice 来动态生成文件名
	outputFile := fmt.Sprintf("%s_speech.mp3", speechVoice)

	// 检查文件是否存在
	if _, err := os.Stat(outputFile); err == nil {
		// 文件存在，尝试删除
		err := os.Remove(outputFile)
		if err != nil {
			// 删除文件时出错
			fmt.Printf("Failed to delete existing file %s: %s", outputFile, err)
			return ""
		}
		fmt.Printf("Existing file %s deleted successfully.\n", outputFile)
	} else if !os.IsNotExist(err) {
		// 访问文件时出现了其他错误
		fmt.Printf("Error checking file %s: %s", outputFile, err)
		return ""
	}

	// 保存 buf 到文件为 mp3
	err = os.WriteFile(outputFile, buf, 0644)
	if err != nil {
		fmt.Printf("WriteFile error: %v", err)
		return ""
	}

	return "ok"
}

// OpenAIError结构体用于封装OpenAI错误
type OpenAIError struct {
	StatusCode int
}

// 解析OpenAI错误
func parseOpenAIError(err error) *OpenAIError {
	var statusCode int
	reStatusCode := regexp.MustCompile(`status code: (\d+)`)
	if match := reStatusCode.FindStringSubmatch(err.Error()); match != nil {
		statusCode, _ = strconv.Atoi(match[1]) // 将字符串转换为整数
	}

	return &OpenAIError{
		StatusCode: statusCode,
	}
}

// 处理OpenAI错误
func (xiao_wan Xiao_wan) openaiError(err error) {
	parsedError := parseOpenAIError(err)

	switch parsedError.StatusCode {
	case 401:
		fmt.Println("Invalid OpenAI API key. Please enter a valid key.")
		fmt.Println("You can find your API key at https://beta.openai.com/account/api-keys")
		fmt.Println("You can also set your API key as an environment variable named OPENAI_API_KEY")
	case 429:
		fmt.Println("Rate limit exceeded. Please wait and try again.")
	case 500:
		fmt.Println("Internal server error. Please try again later.")
	default:
		// 处理其他错误
		fmt.Println("Unknown error: ", parsedError.StatusCode)
	}
}
