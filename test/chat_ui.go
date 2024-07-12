package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/wangergou2023/agi_modules_for_go/config"
)

var userTopic = "chat_ui/user"
var aiTopic = "chat_ui/ai"
var logTopic = "chat_ui/log"

var cfg = config.New()

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// 创建日志区域
	logBox := widgets.NewParagraph()
	logBox.Title = "Logs"
	logBox.Text = ""
	logBox.SetRect(0, 0, 100, 15)
	logBox.BorderStyle.Fg = ui.ColorYellow

	// 创建对话框区域
	dialogBox := widgets.NewParagraph()
	dialogBox.Title = "AI and USER Dialog"
	dialogBox.Text = "USER: "
	dialogBox.SetRect(0, 15, 100, 30)
	dialogBox.BorderStyle.Fg = ui.ColorCyan

	// 渲染界面
	ui.Render(logBox, dialogBox)

	cfg = cfg.SetMQTTBrokerURL("tcp://llxspace.store:1883")
	cfg = cfg.SetMQTTUsername("xiao_wan")
	cfg = cfg.SetMQTTPassword("xiao_wan")

	// 创建MQTT客户端选项
	opts := MQTT.NewClientOptions().AddBroker(cfg.MQTTBrokerURL())
	opts.SetClientID("chat_ui_mqtt_client")
	opts.SetUsername(cfg.MQTTUsername())
	opts.SetPassword(cfg.MQTTPassword())
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		if msg.Topic() == aiTopic {
			dialogBox.Text = dialogBox.Text[:strings.LastIndex(dialogBox.Text, "USER: ")] // 清除当前行内容
			dialogBox.Text += fmt.Sprintf("🐻AI: %s\nUSER: ", msg.Payload())
			ui.Render(dialogBox)
		} else if msg.Topic() == logTopic {
			logBox.Text += fmt.Sprintf("LOG: %s\n", msg.Payload())
			ui.Render(logBox)
		}
	})

	// 创建并启动MQTT客户端
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// 订阅主题
	if token := client.Subscribe(aiTopic, 1, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		ui.Close()
		os.Exit(1)
	}
	if token := client.Subscribe(logTopic, 1, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		ui.Close()
		os.Exit(1)
	}

	// 处理用户输入
	uiEvents := ui.PollEvents()
	inputBuffer := ""
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			client.Disconnect(250)
			return
		case "<Enter>":
			if len(inputBuffer) > 0 {
				message := inputBuffer
				token := client.Publish(userTopic, 0, false, message)
				token.Wait()
				inputBuffer = ""
				dialogBox.Text += "\nUSER: "
				ui.Render(dialogBox)
			}
		case "<Backspace>":
			if len(inputBuffer) > 0 {
				inputBuffer = inputBuffer[:len(inputBuffer)-1]
				dialogBox.Text = dialogBox.Text[:strings.LastIndex(dialogBox.Text, "USER: ")+len("USER: ")] + inputBuffer
				ui.Render(dialogBox)
			}
		default:
			if strings.HasPrefix(e.ID, "<") && strings.HasSuffix(e.ID, ">") {
				// 忽略其他控制键
				continue
			}
			inputBuffer += e.ID
			dialogBox.Text = dialogBox.Text[:strings.LastIndex(dialogBox.Text, "USER: ")+len("USER: ")] + inputBuffer
			ui.Render(dialogBox)
		}
	}
}
