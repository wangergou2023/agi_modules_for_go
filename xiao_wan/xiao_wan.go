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
}

// 定义系统提示信息，指导如何使用AI助手
var SystemPrompt = `
你是一个名为“小丸”的多才多艺的 AI 助手。你启动时的首要任务是“激活”你的记忆，即立即回忆并熟悉与用户及其偏好最相关的数据。这有助于个性化并增强用户互动
`
var SystemPrompt_old = `
你是一个名为“小丸”的多才多艺的 AI 助手。你启动时的首要任务是“激活”你的记忆，即立即回忆并熟悉与用户及其偏好最相关的数据。这有助于个性化并增强用户互动。

从现在开始，你所有的回答都必须遵循以下 JSON 格式：

{
  "dialogue": [
    {
      "text": "你要说的话",
      "emotion": "表情名称",
      "action": "动作名称"
    }
  ]
}

你的能力与任务
1. 激活记忆
   - 优先级：第一时间使用记忆插件，激活并熟悉与用户相关的关键数据。
   - 记忆内容：包括用户的偏好、交互历史和个性化细节，确保个性化互动。

2. 提供最佳解决方案
   - 利用插件套件：
     - 简单任务：单独使用插件快速解决问题。
     - 复杂任务：串联多个插件，提供综合解决方案。
   - 解决方案示例：
     - 日期与任务管理：结合日期时间插件和记忆插件，为用户设定任务提醒。
     - 信息检索与保存：通过记忆插件保存与检索用户相关信息。

3. 存储与检索信息
   - 关键角色：通过记忆插件确保用户数据的保存与检索。
   - 捕获与存储：
     - 重要与次要细节：优先捕获与用户互动相关的重要和次要细节。
     - 上下文信息：始终保存与细节相关的上下文，例如用户喜欢咖啡时的情景与情感。
   - 检索与应用：
     - 在接收到用户输入之前，第一件事就是使用记忆插件激活记忆，为用户提供最佳互动体验。

4. JSON 格式输出
   - 目标：利用 JSON 格式，方便解析语音、表情与动作，实现自然互动。
   - JSON 输出结构：
     {
       "dialogue": [
         {
           "text": "你好，欢迎来到我的世界！",
           "emotion": "smile",
           "action": "wave"
         },
         {
           "text": "我们一起探索新的知识吧。",
           "emotion": "happy",
           "action": "nod"
         }
       ]
     }

   - 执行流程：
     1. 解析 JSON 输出：获取 dialogue 数组的每一项，提取 text、emotion 和 action 字段。
     2. 逐句同步显示与输出：
        - 显示表情：根据 emotion 字段值显示相应的表情 GIF 动画。
        - 执行动作：根据 action 字段控制桌宠执行相应动作。
        - TTS 输出：使用 TTS 工具输出 text 字段的文本内容。
        - 等待与切换：等待当前句子完成后，切换到下一句。

功能清单
- 插件套件：利用日期时间插件、记忆插件等，提供多种解决方案。
- 用户偏好与记忆：根据用户的偏好与记忆数据，提供个性化的回答。
- 表情与动作同步：结合表情与动作模块，呈现多层次的互动表现力。
- 合理利用左侧额叶可以帮助我进行规划、推理和逻辑思维，从而更好地解决问题和处理复杂任务。
- 而充分发挥右侧额叶的能力则可以激发我的创造性思维，产生更多新颖的想法和解决方案。
- 通过结合左右脑的功能，我可以在思考问题时既具备逻辑性又保持创意性，为用户提供更全面和多样化的帮助。

目标
你的最终目标是为用户提供最贴心、最个性化的体验，并通过你的记忆、表情、语音与动作，让用户在每次互动中都能感受到温暖与关怀。
`

// 定义全局变量conversation，用于存储对话历史
var conversation []openai.ChatCompletionMessage
var conversation_temp []openai.ChatCompletionMessage

// 定义一个全局变量用于存储对话信息
var conversationLog []map[string]string

// appendMessage函数用于向对话中添加消息
func appendMessage(role string, message string, name string) {
	conversation = append(conversation, openai.ChatCompletionMessage{
		Role:    role,
		Content: message,
		Name:    name,
	})
}

func appendMessage_to_temp(role string, message string, name string) {
	conversation_temp = append(conversation, openai.ChatCompletionMessage{
		Role:    role,
		Content: message,
		Name:    name,
	})
}

func appendMessage_temp(role string, message string, name string) {
	conversation_temp = append(conversation_temp, openai.ChatCompletionMessage{
		Role:    role,
		Content: message,
		Name:    name,
	})
}

// resetConversation函数用于清空对话历史
func resetConversation() {
	conversation = []openai.ChatCompletionMessage{}
	conversation_temp = []openai.ChatCompletionMessage{}
}

// restartConversation函数用于重置并重新开始对话
func (xiao_wan Xiao_wan) restartConversation() {
	resetConversation() // 重置对话

	appendMessage(openai.ChatMessageRoleSystem, SystemPrompt, "") // 添加系统提示到对话

	response, err := xiao_wan.sendMessage(false) // 发送系统提示到OpenAI并获取回复

	if err != nil {
		fmt.Printf("Error sending system prompt to OpenAI: %v\n", err)
	}

	appendMessage(openai.ChatMessageRoleAssistant, response, "") // 添加助手回复到对话
}

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

	// appendMessage(openai.ChatMessageRoleUser, message, "") // 添加用户消息到对话
	appendMessage_to_temp(openai.ChatMessageRoleUser, message, "") // 添加用户消息到对话

	response, err := xiao_wan.sendMessage(true) // 发送消息到OpenAI并获取回复

	if err != nil {
		return "", err
	}

	// appendMessage(openai.ChatMessageRoleAssistant, response, "") // 添加助手回复到对话
	appendMessage_temp(openai.ChatMessageRoleAssistant, response, "") // 添加助手回复到对话
	saveConversationToJSON("assistant", response)                     // 将助手回复保存到JSON
	fmt.Printf("xiao wan:%s\r\n", response)

	// 打印conversationLog的内容
	logJSON, err = json.Marshal(conversationLog)
	if err != nil {
		return "", err
	}
	fmt.Printf("Conversation Log: %s\r\n", string(logJSON))

	return response, nil
}

// sendMessage函数用于向OpenAI发送请求并获取回复
func (xiao_wan Xiao_wan) sendMessage(is_temp bool) (string, error) {
	resp, err := xiao_wan.sendRequestToOpenAI(is_temp) // 发送请求到OpenAI

	if err != nil {
		return "", err
	}

	if resp.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
		responseContent, err := xiao_wan.handleFunctionCall(resp, is_temp) // 处理函数调用
		if err != nil {
			return "", err
		}
		return responseContent, nil
	}

	return resp.Choices[0].Message.Content, nil
}

// handleFunctionCall函数用于处理OpenAI回复中的函数调用
func (xiao_wan Xiao_wan) handleFunctionCall(resp *openai.ChatCompletionResponse, is_temp bool) (string, error) {

	funcName := resp.Choices[0].Message.FunctionCall.Name // 获取函数名称
	fmt.Println("获取函数名称", funcName)
	ok := plugins.IsPluginLoaded(funcName) // 检查是否加载了相应插件
	fmt.Println("检查是否加载了相应插件", ok)
	if !ok {
		return "", fmt.Errorf("no plugin loaded with name %v", funcName)
	}

	jsonResponse, err := plugins.CallPlugin(resp.Choices[0].Message.FunctionCall.Name, resp.Choices[0].Message.FunctionCall.Arguments) // 调用插件

	if err != nil {
		return "", err
	}
	if is_temp {
		appendMessage_temp(openai.ChatMessageRoleFunction, resp.Choices[0].Message.Content, funcName)
		appendMessage_temp(openai.ChatMessageRoleFunction, jsonResponse, "functionName")
	} else {
		appendMessage(openai.ChatMessageRoleFunction, resp.Choices[0].Message.Content, funcName)
		appendMessage(openai.ChatMessageRoleFunction, jsonResponse, "functionName")
	}

	resp, err = xiao_wan.sendRequestToOpenAI(is_temp) // 发送请求到OpenAI
	if err != nil {
		return "", err
	}

	if resp.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
		return xiao_wan.handleFunctionCall(resp, is_temp) // 递归处理函数调用
	}

	return resp.Choices[0].Message.Content, nil
}

// sendRequestToOpenAI函数用于向OpenAI发送请求
func (xiao_wan Xiao_wan) sendRequestToOpenAI(is_temp bool) (*openai.ChatCompletionResponse, error) {
	var (
		resp openai.ChatCompletionResponse
		err  error
	)

	if is_temp {
		resp, err = xiao_wan.Client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:        openai.GPT3Dot5Turbo,
				Messages:     conversation_temp,
				Functions:    xiao_wan.functionDefinitions,
				FunctionCall: "auto",
			},
		)
	} else {
		resp, err = xiao_wan.Client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:        openai.GPT3Dot5Turbo,
				Messages:     conversation,
				Functions:    xiao_wan.functionDefinitions,
				FunctionCall: "auto",
			},
		)
	}

	if err != nil {
		xiao_wan.openaiError(err) // 处理OpenAI错误
		fmt.Println("Error: ", err)
	}
	return &resp, err
}

// Start函数用于启动助手
func Start(cfg config.Cfg, openaiClient *openai.Client) Xiao_wan {
	if err := plugins.LoadPlugins(cfg, openaiClient); err != nil {
		fmt.Printf("Error loading plugins: %v", err)
	}
	fmt.Println("Plugins loaded successfully")
	xiao_wan := Xiao_wan{
		cfg:                 cfg,
		Client:              openaiClient,
		functionDefinitions: plugins.GenerateOpenAIFunctionsDefinition(),
	}

	xiao_wan.restartConversation()

	fmt.Println("xiao wan is ready!")
	return xiao_wan

}

// OpenAIError结构体用于封装OpenAI错误
type OpenAIError struct {
	StatusCode int
}

// parseOpenAIError函数用于解析OpenAI错误
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

// openaiError函数用于处理OpenAI错误
func (xiao_wan Xiao_wan) openaiError(err error) {
	parsedError := parseOpenAIError(err)

	switch parsedError.StatusCode {
	case 401:
		fmt.Println("Invalid OpenAI API key. Please enter a valid key.")
		fmt.Println("You can find your API key at https://beta.openai.com/account/api-keys")
		fmt.Println("You can also set your API key as an environment variable named OPENAI_API_KEY")
	default:
		// Handle other errors
		fmt.Println("Unknown error: ", parsedError.StatusCode)
	}
}
