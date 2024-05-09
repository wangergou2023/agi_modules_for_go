package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/wangergou2023/agi_modules_for_go/config"
	"github.com/wangergou2023/agi_modules_for_go/xiao_wan"
)

var cfg = config.New()

func main() {
	fmt.Println("xiao wan is starting up... Please wait a moment.")

	config := openai.DefaultConfig(cfg.OpenAiAPIKey())
	//need"/v1"
	config.BaseURL = cfg.OpenAibaseURL()
	openaiClient := openai.NewClientWithConfig(config)

	xiao_wan_chat := xiao_wan.Start(cfg, openaiClient)

	// 启动TCP服务端
	go startTCPServer(&xiao_wan_chat)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		xiao_wan_chat.Message(text)
	}
}

// 启动TCP服务端，接收闹钟插件消息
func startTCPServer(xiao_wan_chat *xiao_wan.Xiao_wan) {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Error starting TCP server: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Println("TCP server started on localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go handleConnection(conn, xiao_wan_chat)
	}
}

// 处理插件消息
func handleConnection(conn net.Conn, xiao_wan_chat *xiao_wan.Xiao_wan) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, bufio.ErrBufferFull) || err.Error() == "EOF" {
				// 忽略EOF和其他连接关闭错误
				return
			}
			fmt.Printf("Error reading message: %v\n", err)
			return
		}
		msg = strings.TrimSpace(msg)
		fmt.Printf("Received message from plugin: %s\n", msg)
		xiao_wan_chat.Message(msg)
	}
}
