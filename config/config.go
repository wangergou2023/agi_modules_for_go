// 定义配置包
package config

// 导入必要的包

// 用于格式化输出
// 用于操作系统相关的操作，如文件操作

// 定义Milvus数据库配置的结构体
type MalvusCfg struct {
	apiEndpoint    string // Milvus服务器的API终端地址
	collectionName string // Milvus中用于存储数据的集合名称
}

// 定义主配置结构体
type Cfg struct {
	openAiAPIKey         string    // OpenAI API的密钥
	openAibaseURL        string    // OpenAI 中转地址
	openWeatherMapAPIKey string    // OpenWeatherMap API的密钥
	malvusCfg            MalvusCfg // Milvus数据库的配置
	mqttBrokerURL        string    // MQTT 代理服务器地址
	mqttUsername         string    // MQTT 用户名
	mqttPassword         string    // MQTT 密码
}

// New函数用于创建并初始化Cfg配置实例
func New() Cfg {
	// 初始化Milvus配置
	malvusCfg := MalvusCfg{
		apiEndpoint:    "your:19530", // Milvus API终端地址
		collectionName: "CGPTMemory", // Milvus集合名称
	}

	// 初始化主配置
	cfg := Cfg{
		openAiAPIKey:         "your",      // OpenAI API的密钥
		openAibaseURL:        "your/v1",   // 中转地址
		openWeatherMapAPIKey: "your",      // OpenWeatherMap API的密钥
		malvusCfg:            malvusCfg,   // 设置Milvus配置
		mqttBrokerURL:        "your:1883", // MQTT 代理服务器地址
		mqttUsername:         "your",      // MQTT 用户名
		mqttPassword:         "your",      // MQTT 密码
	}

	return cfg // 返回配置实例
}

// OpenAiAPIKey方法返回OpenAI API的密钥
func (c Cfg) OpenAiAPIKey() string {
	return c.openAiAPIKey
}

func (c Cfg) SetOpenAiAPIKey(openAiAPIKey string) Cfg {
	c.openAiAPIKey = openAiAPIKey
	return c
}

func (c Cfg) OpenAibaseURL() string {
	return c.openAibaseURL
}

func (c Cfg) SetOpenAibaseURL(openAibaseURL string) Cfg {
	c.openAibaseURL = openAibaseURL
	return c
}

func (c Cfg) OpenWeatherMapAPIKey() string {
	return c.openWeatherMapAPIKey
}

// MalvusApiEndpoint方法返回Milvus API终端的地址
func (c Cfg) MalvusApiEndpoint() string {
	return c.malvusCfg.apiEndpoint
}

// SetMalvusApiEndpoint方法设置Milvus API终端的地址
func (c Cfg) SetMalvusApiEndpoint(apiEndpoint string) Cfg {
	c.malvusCfg.apiEndpoint = apiEndpoint
	return c
}

// MalvusCollectionName方法返回Milvus中用于存储数据的集合名称
func (c Cfg) MalvusCollectionName() string {
	return c.malvusCfg.collectionName
}

// SetMalvusCollectionName方法设置Milvus中用于存储数据的集合名称
func (c Cfg) SetMalvusCollectionName(collectionName string) Cfg {
	c.malvusCfg.collectionName = collectionName
	return c
}

// 设置和获取MQTT Broker URL的方法
func (c Cfg) SetMQTTBrokerURL(url string) Cfg {
	c.mqttBrokerURL = url
	return c
}

func (c Cfg) MQTTBrokerURL() string {
	return c.mqttBrokerURL
}

// 设置和获取MQTT用户名的方法
func (c Cfg) SetMQTTUsername(username string) Cfg {
	c.mqttUsername = username
	return c
}

func (c Cfg) MQTTUsername() string {
	return c.mqttUsername
}

// 设置和获取MQTT密码的方法
func (c Cfg) SetMQTTPassword(password string) Cfg {
	c.mqttPassword = password
	return c
}

func (c Cfg) MQTTPassword() string {
	return c.mqttPassword
}
