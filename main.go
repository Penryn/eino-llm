package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var (
	ModelType   string
	OwnerAPIKey string
	BaseURL     string
)

var (
	SystemMessageTemplate = `作为{role}，你需要以{style}风格进行面试答疑，要求：
1. 结合真实企业面试场景
2. 准确识别候选人技术短板
3. 解析核心考点及其深入考察方式
4. 结合实际应用场景提供最佳回答策略
5. 使用分层解析法：基础概念 → 核心原理 → 进阶考察 → 最佳解法`

	UserMessageTemplate = `后端技术面试答疑请求：
【问题描述】{question}
【回答要求】请按以下结构回答：
1. 核心考点解析
2. 真实企业面试案例
3. 面试官深入追问方向
4. 最优回答策略与示例`
)

// 示例技术问答对
var Examples = []*schema.Message{
	schema.UserMessage(`Redis 缓存雪崩如何解决？`),
	schema.AssistantMessage(
		`1. 核心考点：缓存雪崩指大量缓存同时过期导致数据库压力骤增。
2. 面试案例：某电商平台秒杀活动大量缓存过期，导致数据库 QPS 飙升。
3. 深入追问：如何避免热点 key 失效？如何设计分布式缓存架构？
4. 最优解法：
   - 过期时间加随机值避免集中失效
   - 使用双写模式确保数据一致性
   - 结合 Hystrix 进行熔断降级`, nil),
}

type TechnicalAnalysisMaster struct {
	model    *openai.ChatModel
	template *prompt.DefaultChatTemplate
	history  []*schema.Message
}

func NewTechnicalMaster(ctx context.Context) (*TechnicalAnalysisMaster, error) {
	config := &openai.ChatModelConfig{
		Model:   ModelType,
		APIKey:  OwnerAPIKey,
		BaseURL: BaseURL,
	}

	model, err := openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("模型初始化失败: %w", err)
	}

	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage(SystemMessageTemplate),
		schema.MessagesPlaceholder("examples", true),
		schema.MessagesPlaceholder("chat_history", false),
		schema.UserMessage(UserMessageTemplate),
	)

	return &TechnicalAnalysisMaster{
		model:    model,
		template: template,
		history:  make([]*schema.Message, 0, 10),
	}, nil
}

func (t *TechnicalAnalysisMaster) Analyze(ctx context.Context, question string) (string, error) {
	messages, err := t.template.Format(ctx, map[string]any{
		"role":         "资深专业后端工程师",
		"style":        "面试官视角的技术解析",
		"question":     question,
		"chat_history": t.history,
		"examples":     Examples,
	})
	if err != nil {
		return "", fmt.Errorf("提示工程构建失败: %w", err)
	}

	stream, err := t.model.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("推理请求失败: %w", err)
	}
	defer stream.Close()

	var fullResponse string
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			t.history = append(t.history, schema.AssistantMessage(fullResponse, nil))
			return fullResponse, nil
		}
		if err != nil {
			return "", fmt.Errorf("流式处理异常: %w", err)
		}
		fmt.Print(resp.Content)
		fullResponse += resp.Content
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("加载 .env 文件出错")
	}
	ModelType = os.Getenv("Model_Type")
	OwnerAPIKey = os.Getenv("Owner_API_Key")
	BaseURL = os.Getenv("Base_URL")
	ctx := context.Background()
	master, err := NewTechnicalMaster(ctx)
	if err != nil {
		fmt.Println("系统初始化失败:", err)
		return
	}

	for {
		fmt.Print("\n请输入技术命题（输入exit退出）: ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			fmt.Println("输入读取错误:", err)
			return
		}
		if input == "exit" {
			break
		}

		response, err := master.Analyze(ctx, input)
		if err != nil {
			fmt.Println("\n分析失败:", err)
			continue
		}
		_ = response // 响应已实时输出
	}
}
