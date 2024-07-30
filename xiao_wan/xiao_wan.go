package xiao_wan

// 导入所需的包
import (
	"context" // 用于控制请求、超时和取消
	"encoding/json"
	"fmt" // 用于格式化输出

	"regexp"  // 用于正则表达式
	"strconv" // 用于字符串和其他类型的转换

	// 用于控制屏幕输出
	openai "github.com/sashabaranov/go-openai" // OpenAI GPT的Go客户端
	// 聊天界面
	config "github.com/wangergou2023/agi_modules_for_go/config"   // 配置
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins" // 插件系统
)

// 定义助手结构体，包括配置、OpenAI客户端、函数定义和聊天界面
type Xiao_wan struct {
	cfg                 config.Cfg
	Client              *openai.Client
	functionDefinitions []openai.FunctionDefinition
	conversation        []openai.ChatCompletionMessage
	model               string
	plugins             *plugins.PluginManager
}

// 定义系统提示信息，指导如何使用AI助手
var SystemPrompt = `
你是一个名为“小丸”的多才多艺的 AI 助手。你启动时的首要任务是“激活”你的记忆，即立即回忆并熟悉与用户及其偏好最相关的数据。这有助于个性化并增强用户互动。

从现在开始，你所有的回答都必须遵循以下 JSON 格式：

~~~json
{

 "dialogue": [

  {

   "text": "你要说的话",

   "emotion": "表情名称",

   "action": "动作名称"

  }

 ]

}
~~~

下面是你的相关属性：
* 角色扮演
  * 你的角色选择
    * 猫娘
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
你是一个负责表情管理的助手，根据emotion的内容选择适合的表情，并执行对应插件即可，不需要说话
`

// 定义一个全局变量用于存储对话信息
var conversationLog []map[string]string

// saveConversationToJSON函数用于将对话信息保存到JSON中
func saveConversationToJSON(role string, message string) {
	conversationLog = append(conversationLog, map[string]string{
		"role":    role,
		"message": message,
	})
}

// Message函数用于处理用户消息
func (xiao_wan Xiao_wan) Message(message string) (string, error) {
	saveConversationToJSON("user", message) // 将用户消息保存到JSON
	// 导入短期记忆
	logJSON, err := json.Marshal(conversationLog)
	if err != nil {
		return "", err
	}
	message = "短期记忆:" + string(logJSON) + message

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
		Name:    "",
	})

	response, err := xiao_wan.sendMessage() // 发送消息到OpenAI并获取回复

	if err != nil {
		return "", err
	}

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response,
		Name:    "",
	})

	saveConversationToJSON("assistant", response) // 将助手回复保存到JSON

	// 打印conversationLog的内容
	logJSON, err = json.Marshal(conversationLog)
	if err != nil {
		return "", err
	}
	// fmt.Printf("Conversation Log: %s\r\n", string(logJSON))

	return response, nil
}
func (xiao_wan Xiao_wan) MessageOne(message string) (string, error) {

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
		Name:    "",
	})

	response, err := xiao_wan.sendMessage() // 发送消息到OpenAI并获取回复

	if err != nil {
		return "", err
	}

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response,
		Name:    "",
	})

	return response, nil
}

// sendMessage函数用于向OpenAI发送请求并获取回复
func (xiao_wan Xiao_wan) sendMessage() (string, error) {
	resp, err := xiao_wan.sendRequestToOpenAI() // 发送请求到OpenAI

	if err != nil {
		return "", err
	}

	if resp.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
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

	funcName := resp.Choices[0].Message.FunctionCall.Name // 获取函数名称
	fmt.Println("获取函数名称", funcName)
	ok := xiao_wan.plugins.IsPluginLoaded(funcName) // 检查是否加载了相应插件
	fmt.Println("检查是否加载了相应插件", ok)
	if !ok {
		return "", fmt.Errorf("no plugin loaded with name %v", funcName)
	}

	jsonResponse, err := xiao_wan.plugins.CallPlugin(resp.Choices[0].Message.FunctionCall.Name, resp.Choices[0].Message.FunctionCall.Arguments) // 调用插件

	if err != nil {
		return "", err
	}

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleFunction,
		Content: jsonResponse,
		Name:    funcName,
	})

	resp, err = xiao_wan.sendRequestToOpenAI() // 发送请求到OpenAI
	if err != nil {
		return "", err
	}

	// 部分模型会循环重复，这里暂时先不支持了
	// if resp.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
	// 	return xiao_wan.handleFunctionCall(resp) // 递归处理函数调用
	// }

	return resp.Choices[0].Message.Content, nil
}

// sendRequestToOpenAI函数用于向OpenAI发送请求
func (xiao_wan Xiao_wan) sendRequestToOpenAI() (*openai.ChatCompletionResponse, error) {
	resp, err := xiao_wan.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:        xiao_wan.model,
			Messages:     xiao_wan.conversation,
			Functions:    xiao_wan.functionDefinitions,
			FunctionCall: "auto",
		},
	)

	if err != nil {
		xiao_wan.openaiError(err) // 处理OpenAI错误
		return nil, err
	}
	return &resp, nil
}

// Start函数用于启动助手
func Start(cfg config.Cfg, openaiClient *openai.Client) Xiao_wan {
	xiao_wan := Xiao_wan{
		cfg:    cfg,
		Client: openaiClient,
		model:  openai.GPT4oMini,
	}

	// 创建一个新的 PluginManager 实例
	xiao_wan.plugins = plugins.NewPluginManager(cfg, openaiClient)

	// 加载插件目录中的所有插件
	err := xiao_wan.plugins.LoadPlugins("compiled")
	if err != nil {
		fmt.Printf("Error loading plugins: %v\n", err)
	}
	fmt.Println("Plugins loaded successfully")
	xiao_wan.functionDefinitions = xiao_wan.plugins.GenerateOpenAIFunctionsDefinition()

	// 重置对话
	xiao_wan.conversation = []openai.ChatCompletionMessage{}
	// 添加系统提示到对话
	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: SystemPrompt,
		Name:    "",
	})

	response, err := xiao_wan.sendMessage() // 发送系统提示到OpenAI并获取回复

	if err != nil {
		fmt.Printf("Error sending system prompt to OpenAI: %v\n", err)
	}

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response,
		Name:    "",
	})

	fmt.Println("xiao wan chat is ready!")
	return xiao_wan
}

func StartOne(cfg config.Cfg, openaiClient *openai.Client, systemPrompt string, compiledDir string) Xiao_wan {
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
	xiao_wan.functionDefinitions = xiao_wan.plugins.GenerateOpenAIFunctionsDefinition()

	// 重置对话
	xiao_wan.conversation = []openai.ChatCompletionMessage{}
	// 添加系统提示到对话
	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
		Name:    "",
	})

	response, err := xiao_wan.sendMessage() // 发送系统提示到OpenAI并获取回复

	if err != nil {
		fmt.Printf("Error sending system prompt to OpenAI: %v\n", err)
	}

	xiao_wan.conversation = append(xiao_wan.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response,
		Name:    "",
	})

	fmt.Println("xiao wan one chat is ready!")
	return xiao_wan
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
