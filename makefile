PLUGIN_SRC_DIR = ./plugins/source/builtin
PLUGIN_BRAIN_SRC_DIR = ./plugins/source/brain
PLUGIN_SMART_HOME_SRC_DIR = ./plugins/source/smart_home
PLUGIN_DOG_SRC_DIR = ./plugins/source/dog

PLUGIN_FOR_CHAT_DIR = ./plugins/for_chat

PLUGIN_FOR_AFTER_CHAT_DIR = ./plugins/for_after_chat
PLUGIN_FOR_AFTER_CHAT2_DIR = ./plugins/for_after_chat2
PLUGIN_FOR_AFTER_CHAT3_DIR = ./plugins/for_after_chat3

PLUGIN_FOR_BEFORE_CHAT_DIR = ./plugins/for_before_chat
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

clean:
	rm -f $(PLUGIN_FOR_CHAT_DIR)/*.so
	rm -f $(PLUGIN_FOR_BEFORE_CHAT_DIR)/*.so
	rm -f $(PLUGIN_FOR_AFTER_CHAT_DIR)/*.so
	rm -f $(PLUGIN_FOR_AFTER_CHAT2_DIR)/*.so
	rm -f $(PLUGIN_FOR_AFTER_CHAT3_DIR)/*.so

plugin: clean
	# for_chat
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_FOR_CHAT_DIR)/weather2.so $(PLUGIN_SRC_DIR)/weather2/plugin.go
	# for_before_chat
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_FOR_BEFORE_CHAT_DIR)/command.so $(PLUGIN_SRC_DIR)/command/plugin.go

	# 基本插件，分别使用ai去代理 
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_FOR_AFTER_CHAT_DIR)/tts.so $(PLUGIN_SRC_DIR)/tts/plugin.go
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_FOR_AFTER_CHAT2_DIR)/face.so $(PLUGIN_DOG_SRC_DIR)/face/plugin.go
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_FOR_AFTER_CHAT3_DIR)/legs.so $(PLUGIN_DOG_SRC_DIR)/legs/plugin.go
