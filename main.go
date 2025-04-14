package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/cloudwego/eino/schema"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	loggers  = make(map[*websocket.Conn]struct{})
	logMutex sync.Mutex
)

// 自定义日志写入器
type wsLogger struct {
	conn *websocket.Conn
}

func (w *wsLogger) Write(p []byte) (n int, err error) {
	if err := w.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("[LOG] %s", string(p)))); err != nil {
		return 0, err
	}
	return len(p), nil
}

type Message struct {
	Type       string `json:"type"`
	Content    string `json:"content"`
	IsFrontend bool   `json:"isFrontend"`
}

func main() {
	// 提供静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// WebSocket 处理
	http.HandleFunc("/ws", handleWebSocket)

	// 启动服务器
	log.Println("服务器启动在 :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 注册日志写入器
	logMutex.Lock()
	loggers[conn] = struct{}{}
	logMutex.Unlock()

	// 创建自定义日志写入器
	wsLog := &wsLogger{conn: conn}
	log.SetOutput(io.MultiWriter(os.Stdout, wsLog))

	ctx := context.Background()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("解析消息失败: %v", err)
			continue
		}

		switch msg.Type {
		case "switch_assistant":
			log.Printf("切换到%s面试助手", map[bool]string{true: "前端", false: "后端"}[msg.IsFrontend])
		case "message":
			// 处理消息
			if err := streamAgentResponse(ctx, conn, msg.Content, msg.IsFrontend); err != nil {
				log.Printf("处理消息失败: %v", err)
				break
			}
		}
	}

	// 注销日志写入器
	logMutex.Lock()
	delete(loggers, conn)
	logMutex.Unlock()
}

func streamAgentResponse(ctx context.Context, conn *websocket.Conn, msg string, isFrontend bool) error {
	runner, err := buildeinoLLM(ctx, isFrontend)
	if err != nil {
		return fmt.Errorf("failed to build agent graph: %w", err)
	}

	sr, err := runner.Stream(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to stream: %w", err)
	}
	defer sr.Close()

	var fullResponse strings.Builder
	for {
		resp, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			// 只在有内容时才添加到历史记录
			if fullResponse.Len() > 0 {
				History = append(History, schema.AssistantMessage(fullResponse.String(), nil))
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("流式处理异常: %w", err)
		}

		// 累积完整响应
		fullResponse.WriteString(resp.Content)

		// 发送当前片段
		if err := conn.WriteMessage(websocket.TextMessage, []byte(resp.Content)); err != nil {
			return fmt.Errorf("发送消息失败: %w", err)
		}
	}
}
