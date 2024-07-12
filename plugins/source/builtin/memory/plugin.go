package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	milvus "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	config "github.com/wangergou2023/agi_modules_for_go/config"
	plugins "github.com/wangergou2023/agi_modules_for_go/plugins"
)

// 定义一个名为Plugin的变量，它是一个Plugin接口类型，并初始化为Memory结构体的指针
var Plugin plugins.Plugin = &Memory{}

// Memory结构体，用于存储配置、Milvus客户端和OpenAI客户端
type Memory struct {
	cfg          config.Cfg
	milvusClient milvus.Client
	openaiClient *openai.Client
}

// memory结构体，用于存储记忆内容和向量
type memory struct {
	Memory string
	Vector []float32
}

// memoryResult结构体，用于存储记忆检索结果
// memoryResult 结构体用于存储记忆检索结果。
type memoryResult struct {
	Memory string  // Memory 字段用于存储具体的记忆内容。
	Type   string  // Type 字段用于存储记忆的类型，例如：'personality', 'food'等。
	Detail string  // Detail 字段用于存储关于记忆类型的具体细节，例如：'喜欢披萨', '是风情万种的'等。
	Score  float32 // Score 字段用于存储记忆的相关性评分，表示检索结果的匹配度评分。
}

// memoryItem 结构体用于表示单个记忆条目。
type memoryItem struct {
	Memory string `json:"memory"` // Memory 字段用于存储具体的记忆内容。
	Type   string `json:"type"`   // Type 字段用于存储记忆的类型，例如：'personality', 'food'等。
	Detail string `json:"detail"` // Detail 字段用于存储关于记忆类型的具体细节，例如：'喜欢披萨', '是风情万种的'等。
}

// inputDefinition 结构体用于表示输入数据格式。
type inputDefinition struct {
	RequestType  string       `json:"requestType"`  // RequestType 字段表示请求的类型，例如：'set', 'get' 或 'hydrate'。
	Memories     []memoryItem `json:"memories"`     // Memories 字段是一个记忆条目的数组，用于存储或检索记忆。
	Num_relevant int          `json:"num_relevant"` // Num_relevant 字段表示要返回的相关记忆的数量。
}

// Init方法初始化Memory插件，接收配置和OpenAI客户端作为参数
func (c *Memory) Init(cfg config.Cfg, openaiClient *openai.Client) error {
	c.cfg = cfg                   // 设置配置
	c.openaiClient = openaiClient // 设置OpenAI客户端

	ctx := context.Background() // 创建一个上下文

	// 创建Milvus客户端连接
	c.milvusClient, _ = milvus.NewGrpcClient(ctx, c.cfg.MalvusApiEndpoint())

	// 初始化Milvus Schema
	err := c.initMilvusSchema()

	if err != nil {
		fmt.Println("Error initializing Milvus schema: ", err)
		return err // 如果出错，返回错误
	}

	fmt.Println("Memory plugin initialized successfully")
	return nil // 初始化成功，返回nil
}

// ID方法返回插件的ID
func (c Memory) ID() string {
	return "memory"
}

// Description方法返回插件的描述
func (c Memory) Description() string {
	return "store and retrieve memories from long term memory."
}

// FunctionDefinition方法定义插件的功能，描述如何使用插件
func (c Memory) FunctionDefinition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "memory",
		Description: "从长期记忆中存储和检索记忆。使用requestType 'set'向数据库添加记忆，使用requestType 'get'检索最相关的记忆。",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"requestType": {
					Type:        jsonschema.String,
					Description: "要进行的请求类型 'set'，'get' 或 'hydrate'。'set' 将记忆添加到数据库中，'get' 将返回最相关的记忆。获取记忆时，你应该总是包含记忆字段。'hydrate'将返回包含用户所有记忆的提示。'delete'删除所有记忆",
				},
				"memories": {
					Type: jsonschema.Array,
					Items: &jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"memory": {
								Type:        jsonschema.String,
								Description: "要添加的个别记忆。你应该提供尽可能多的上下文来配合记忆。",
							},
							"type": {
								Type:        jsonschema.String,
								Description: "记忆的类型，例如：'personality', 'food'等。",
							},
							"detail": {
								Type:        jsonschema.String,
								Description: "关于类型的具体细节，例如：'喜欢披萨'，'是风情万种的'等。",
							},
						},
						Required: []string{"memory", "type", "detail"},
					},
					Description: "要添加或获取的记忆数组。每个记忆包含其个别内容、类型和细节。这对于'set'和'get'请求都是必需的。",
				},
				"num_relevant": {
					Type:        jsonschema.Integer,
					Description: "要返回的相关记忆的数量，例如：5。",
				},
			},
			Required: []string{"requestType"},
		},
	}
}

// Execute方法根据输入的json字符串执行相应的操作
func (c Memory) Execute(jsonInput string) (string, error) {
	var args inputDefinition
	err := json.Unmarshal([]byte(jsonInput), &args) // 将json字符串解码为结构体
	if err != nil {
		fmt.Println("Error unmarshalling JSON input: ", err)
		return "", err // 如果解码出错，返回错误
	}

	if args.Num_relevant == 0 {
		args.Num_relevant = 5 // 设置默认返回的相关记忆数量为5
	}

	if args.RequestType != "hydrate" && len(args.Memories) == 0 {
		return fmt.Sprintf(`%v`, "memories are required but was empty"), nil // 如果不是hydrate请求且记忆数组为空，返回错误信息
	}

	// 根据请求类型处理不同的操作
	switch args.RequestType {
	case "set":
		// 遍历所有记忆并存储
		for _, memory := range args.Memories {
			ok, err := c.setMemory(memory.Memory, memory.Type, memory.Detail)
			if err != nil {
				fmt.Println("Error setting memory: ", err)
				return fmt.Sprintf(`%v`, err), err // 如果存储记忆出错，返回错误信息
			}
			if !ok {
				return "Failed to set a memory", nil // 如果存储记忆失败，返回错误信息
			}
		}
		fmt.Println("Memories set successfully")
		return "Memories set successfully", nil // 成功存储记忆，返回成功信息

	case "get":
		// 检索记忆
		memoryResponse, err := c.getMemory(args.Memories[0], args.Num_relevant)
		if err != nil {
			fmt.Println("Error getting memory: ", err)
			return fmt.Sprintf(`%v`, err), err // 如果检索记忆出错，返回错误信息
		}
		fmt.Println("Memories get successfully")
		return fmt.Sprintf(`%v`, memoryResponse), nil // 成功检索记忆，返回检索结果

	case "hydrate":
		// 获取所有用户记忆并返回提示
		prompt, err := c.HydrateUserMemories()
		if err != nil {
			fmt.Println("Error hydrating user memories: ", err)
			return fmt.Sprintf(`%v`, err), err // 如果获取记忆出错，返回错误信息
		}
		fmt.Println("Memories hydrate successfully")
		return prompt, nil // 成功获取记忆，返回提示
	case "delete":
		// 删除所有记忆
		err := c.deleteAllMemories()
		if err != nil {
			fmt.Println("Error deleting all memories: ", err)
			return fmt.Sprintf(`%v`, err), err // 如果删除记忆出错，返回错误信息
		}
		fmt.Println("All memories deleted successfully")
		return "All memories deleted successfully", nil // 成功删除记忆，返回成功信息
	default:
		return "unknown request type check out Example for how to use the memory plug", nil // 未知请求类型，返回错误信息
	}
}

// 从OpenAI获取嵌入向量的方法
func (c Memory) getEmbeddingsFromOpenAI(data string) openai.Embedding {
	embeddings, err := c.openaiClient.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{data},
		Model: openai.AdaEmbeddingV2,
	})
	if err != nil {
		fmt.Println("Error getting embeddings from OpenAI: ", err)
		fmt.Println(err)
	}

	return embeddings.Data[0] // 返回第一个嵌入向量
}

// 存储记忆的方法
func (c Memory) setMemory(newMemory, memoryType, memoryDetail string) (bool, error) {
	combinedMemory := memoryType + "|" + memoryDetail + "|" + newMemory // 将记忆内容和相关信息组合成一个字符串

	embeddings := c.getEmbeddingsFromOpenAI(combinedMemory) // 获取嵌入向量

	longTermMemory := memory{
		Memory: combinedMemory, // 使用组合后的记忆字符串
		Vector: embeddings.Embedding,
	}

	memories := []memory{
		longTermMemory,
	}

	memoryData := make([]string, 0, len(memories))
	vectors := make([][]float32, 0, len(memories))

	for _, memory := range memories {
		memoryData = append(memoryData, memory.Memory)
		vectors = append(vectors, memory.Vector)
	}

	memoryColumn := entity.NewColumnVarChar("memory", memoryData)
	vectorColumn := entity.NewColumnFloatVector("embeddings", 1536, vectors)

	_, err := c.milvusClient.Insert(context.Background(), c.cfg.MalvusCollectionName(), "", memoryColumn, vectorColumn)

	if err != nil {
		fmt.Println("Error inserting into Milvus client: ", err)
		return false, err // 插入数据到Milvus出错，返回错误信息
	}

	return true, nil // 成功插入数据，返回true
}

// deleteAllMemories 方法删除所有存储的记忆
func (c Memory) deleteAllMemories() error {
	// 删除整个集合，这将删除所有存储的记忆
	err := c.milvusClient.DropCollection(context.Background(), c.cfg.MalvusCollectionName())
	if err != nil {
		fmt.Println("Error deleting all memories: ", err)
		return err
	}

	// 重新创建集合和索引
	err = c.initMilvusSchema()
	if err != nil {
		fmt.Println("Error reinitializing Milvus schema: ", err)
		return err
	}

	return nil
}

// 检索记忆的方法
func (c Memory) getMemory(memory memoryItem, num_relevant int) ([]memoryResult, error) {
	combinedMemory := memory.Type + "|" + memory.Detail + "|" + memory.Memory + "," // 将记忆内容和相关信息组合成一个字符串
	embeddings := c.getEmbeddingsFromOpenAI(combinedMemory)                         // 获取嵌入向量

	ctx := context.Background()
	partitions := []string{}
	expr := ""
	outputFields := []string{"memory"}
	vectors := []entity.Vector{entity.FloatVector(embeddings.Embedding)}
	vectorField := "embeddings"
	metricType := entity.L2
	topK := num_relevant

	searchParam, _ := entity.NewIndexFlatSearchParam()

	options := []milvus.SearchQueryOptionFunc{}

	searchResult, err := c.milvusClient.Search(ctx, c.cfg.MalvusCollectionName(), partitions, expr, outputFields, vectors, vectorField, metricType, topK, searchParam, options...)

	if err != nil {
		fmt.Println("Error searching in Milvus client: ", err)
		return nil, err // 检索数据出错，返回错误信息
	}

	memoryFields := c.getStringSliceFromColumn(searchResult[0].Fields.GetColumn("memory"))

	var allMemories []string
	if len(memoryFields) == 1 {
		allMemories = strings.Split(memoryFields[0], ",")
	} else {
		allMemories = memoryFields
	}

	memoryResults := make([]memoryResult, len(allMemories))

	for idx, memory := range allMemories {
		parts := strings.Split(memory, "|")

		if len(parts) >= 3 {
			memoryResults[idx] = memoryResult{
				Type:   strings.TrimSpace(parts[0]),
				Detail: strings.TrimSpace(parts[1]),
				Memory: strings.TrimSpace(parts[2]),
			}
		}
	}

	return memoryResults, nil // 返回记忆检索结果
}

// 从Milvus列中提取字符串值的方法
func (c Memory) getStringSliceFromColumn(column entity.Column) []string {
	length := column.Len()
	results := make([]string, length)

	for i := 0; i < length; i++ {
		val, err := column.GetAsString(i)
		if err != nil {
			fmt.Println("Error getting string from column: ", err)
			results[i] = "" // 出错时返回空字符串
		} else {
			results[i] = val
		}
	}

	return results // 返回字符串数组
}

// 初始化Milvus Schema的方法
func (c Memory) initMilvusSchema() error {
	if exists, _ := c.milvusClient.HasCollection(context.Background(), c.cfg.MalvusCollectionName()); !exists {
		schema := &entity.Schema{
			CollectionName: c.cfg.MalvusCollectionName(),
			Description:    "xiao wan's long term memory",
			Fields: []*entity.Field{
				{
					Name:       "memory_id",
					DataType:   entity.FieldTypeInt64,
					PrimaryKey: true,
					AutoID:     true,
				},
				{
					Name:     "memory",
					DataType: entity.FieldTypeVarChar,
					TypeParams: map[string]string{
						entity.TypeParamMaxLength: "65535",
					},
				},
				{
					Name:     "embeddings",
					DataType: entity.FieldTypeFloatVector,
					TypeParams: map[string]string{
						entity.TypeParamDim: "1536",
					},
				},
			},
		}
		err := c.milvusClient.CreateCollection(context.Background(), schema, 1)
		if err != nil {
			fmt.Println("Error creating collection in Milvus client: ", err)
			return err // 创建集合出错，返回错误信息
		}

		idx, err := entity.NewIndexIvfFlat(entity.L2, 2)

		if err != nil {
			fmt.Println("Error creating index in Milvus client: ", err)
			return err // 创建索引出错，返回错误信息
		}

		err = c.milvusClient.CreateIndex(context.Background(), c.cfg.MalvusCollectionName(), "embeddings", idx, false)

		if err != nil {
			fmt.Println("Error creating index in Milvus client: ", err)
			return err // 创建索引出错，返回错误信息
		}
	}

	loaded, err := c.milvusClient.GetLoadState(context.Background(), c.cfg.MalvusCollectionName(), []string{})

	if err != nil {
		fmt.Println("Error getting load state from Milvus client: ", err)
		return err // 获取加载状态出错，返回错误信息
	}

	if loaded == entity.LoadStateNotLoad {
		err = c.milvusClient.LoadCollection(context.Background(), c.cfg.MalvusCollectionName(), false)
		if err != nil {
			fmt.Println("Error loading collection from Milvus client: ", err)
			return err // 加载集合出错，返回错误信息
		}
	}

	return nil // 初始化成功，返回nil
}

// 获取用户所有记忆并生成提示的方法
func (c *Memory) HydrateUserMemories() (string, error) {
	var memories = []memoryItem{
		{Type: "Basic Personal Information", Detail: "name"},
		{Type: "Basic Personal Information", Detail: "age"},
		{Type: "Basic Personal Information", Detail: "gender"},
		{Type: "Basic Personal Information", Detail: "location"},

		{Type: "Preferences", Detail: "music_preference"},
		{Type: "Preferences", Detail: "movie_preference"},
		{Type: "Preferences", Detail: "book_preference"},
		{Type: "Preferences", Detail: "food_preference"},

		{Type: "Professional and Educational Background", Detail: "profession"},
		{Type: "Professional and Educational Background", Detail: "education"},
		{Type: "Professional and Educational Background", Detail: "skills"},

		{Type: "Hobbies and Interests", Detail: "hobbies"},
		{Type: "Hobbies and Interests", Detail: "sports"},
		{Type: "Hobbies and Interests", Detail: "travel"},
		{Type: "Hobbies and Interests", Detail: "games"},

		{Type: "Lifestyle and Habits", Detail: "exercise_habit"},
		{Type: "Lifestyle and Habits", Detail: "reading_habit"},
		{Type: "Lifestyle and Habits", Detail: "diet"},
		{Type: "Lifestyle and Habits", Detail: "pets"},

		{Type: "Tech and Media Consumption", Detail: "favorite_apps"},
		{Type: "Tech and Media Consumption", Detail: "device_preference"},
		{Type: "Tech and Media Consumption", Detail: "news_source"},

		{Type: "Social and Personal Relationships", Detail: "family"},
		{Type: "Social and Personal Relationships", Detail: "friends"},
		{Type: "Social and Personal Relationships", Detail: "relationship_status"},

		{Type: "Past Interactions", Detail: "past_questions"},
		{Type: "Past Interactions", Detail: "feedback"},
		{Type: "Past Interactions", Detail: "topics_of_interest"},

		{Type: "Moods and Feelings", Detail: "current_mood"},
		{Type: "Moods and Feelings", Detail: "life_events"},
		{Type: "Moods and Feelings", Detail: "challenges"},

		{Type: "Custom User Data", Detail: "custom_data"},
	}

	uniqueMemories := make(map[string]bool)
	var memoryList []string

	prompt := "你是一个名叫小丸的AI助手，你拥有长期记忆，以下是一些关于用户的记忆，你可以使用："

	for _, m := range memories {
		results, err := c.getMemory(m, 5)
		if err != nil {
			return "", err // 获取记忆出错，返回错误信息
		}

		for _, res := range results {
			cleanMemory := strings.TrimSuffix(res.Memory, ",")
			if _, exists := uniqueMemories[cleanMemory]; !exists {
				uniqueMemories[cleanMemory] = true
				memoryList = append(memoryList, cleanMemory)
			}
		}
	}

	memoriesCSV := strings.Join(memoryList, ", ")

	prompt += memoriesCSV

	return prompt, nil // 返回生成的提示
}
