PLUGIN_SRC_DIR = ./plugins/source/builtin
PLUGIN_BRAIN_SRC_DIR = ./plugins/source/brain
PLUGIN_SMART_HOME_SRC_DIR = ./plugins/source/smart_home
PLUGIN_DOG_SRC_DIR = ./plugins/source/dog

PLUGIN_COMPILED_DIR = ./plugins/compiled
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

clean:
	rm -f $(PLUGIN_COMPILED_DIR)/*.so

plugin: clean
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/alarm.so $(PLUGIN_SRC_DIR)/alarm/plugin.go
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/time.so $(PLUGIN_SRC_DIR)/time/plugin.go
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/face.so $(PLUGIN_DOG_SRC_DIR)/face/plugin.go

	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/memory.so $(PLUGIN_SRC_DIR)/memory/plugin.go
	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/time.so $(PLUGIN_SRC_DIR)/time/plugin.go
	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/weather.so $(PLUGIN_SRC_DIR)/weather/plugin.go
	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/role_player.so $(PLUGIN_SRC_DIR)/role_player/plugin.go
	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/vision.so $(PLUGIN_SRC_DIR)/vision/plugin.go
	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/alarm.so $(PLUGIN_SRC_DIR)/alarm/plugin.go
	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/left_frontal_lobe.so $(PLUGIN_BRAIN_SRC_DIR)/left_frontal_lobe/plugin.go
	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/right_frontal_lobe.so $(PLUGIN_BRAIN_SRC_DIR)/right_frontal_lobe/plugin.go
	# GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o $(PLUGIN_COMPILED_DIR)/seat.so $(PLUGIN_SMART_HOME_SRC_DIR)/seat/plugin.go
