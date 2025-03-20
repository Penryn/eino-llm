package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"strings"
)

// 模型基本信息
var (
	ModelType   string // 模型名称
	OwnerAPIKey string // apiKey
	BaseURL     string // 模型api地址
)

var (
	// 系统信息背景
	SystemMessageTemplate = `作为{role}，你需要以{style}风格进行面试答疑，要求：  
1. 结合真实企业面试场景  
2. 准确识别候选人技术短板  
3. 解析核心考点及其深入考察方式  
4. 结合实际应用场景提供最佳回答策略  
5. 使用分层解析法：基础概念 → 核心原理 → 进阶考察 → 最佳解法`

	UserMessageTemplate = `后端技术面试答疑请求：
【问题描述】{question}  // 你的问题
【参考资料】{source}	  // 联网搜集到的提供的相关资料链接
【回答要求】请按以下结构回答：
1. 核心考点解析
2. 真实企业面试案例
3. 最优回答策略与示例
4. 面试官深入追问方向`
)

// 示例技术问答对
var Examples = []*schema.Message{
	schema.UserMessage(`Redis 缓存雪崩如何解决？`),
	schema.AssistantMessage(
		`1. 核心考点：缓存雪崩指大量缓存同时过期导致数据库压力骤增。
2. 面试案例：某电商平台秒杀活动大量缓存过期，导致数据库 QPS 飙升。
3. 最优解法：
   - 过期时间加随机值避免集中失效
   - 使用双写模式确保数据一致性
   - 结合 Hystrix 进行熔断降级
4. 深入追问：如何避免热点 key 失效？如何设计分布式缓存架构？`, nil),
}

type TechnicalAnalysisMaster struct {
	model    *openai.ChatModel
	template *prompt.DefaultChatTemplate
	history  []*schema.Message
}

// 配置agent基本信息
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

// 调用大模型解析问题获取答案
func (t *TechnicalAnalysisMaster) Analyze(ctx context.Context, question string, source []string) (string, error) {
	messages, err := t.template.Format(ctx, map[string]any{
		"role":         "资深专业后端工程师",
		"style":        "面试官视角的技术解析",
		"question":     question,
		"source":       strings.Join(source, "+"),
		"chat_history": t.history,
		"examples":     Examples,
	})
	if err != nil {
		return "", fmt.Errorf("提示工程构建失败: %w", err)
	}
	// 流式输出
	stream, err := t.model.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("推理请求失败: %w", err)
	}
	defer stream.Close()

	var fullResponse strings.Builder
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			t.history = append(t.history, schema.AssistantMessage(fullResponse.String(), nil))
			return fullResponse.String(), nil
		}
		if err != nil {
			return "", fmt.Errorf("流式处理异常: %w", err)
		}
		fmt.Print(resp.Content)
		fullResponse.WriteString(resp.Content)
	}
}

// 用户交互
func getMultilineInput() string {
	fmt.Println("\n请输入技术问题（输入空行结束）:")

	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func main() {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatal("加载 .env 文件出错")
	}

	ModelType = os.Getenv("Model_Type")
	OwnerAPIKey = os.Getenv("Owner_API_Key")
	BaseURL = os.Getenv("Base_URL")

	// 检查环境变量
	if ModelType == "" || OwnerAPIKey == "" || BaseURL == "" {
		log.Fatal("请确保 .env 配置了 Model_Type, Owner_API_Key, Base_URL")
	}

	ctx := context.Background()

	master, err := NewTechnicalMaster(ctx)
	if err != nil {
		log.Fatal("系统初始化失败:", err)
	}

	for {
		input := getMultilineInput()
		if input == "" {
			continue
		}
		if input == "exit" {
			break
		}

		fmt.Println("正在搜索相关资料，请稍候...")
		// 创建 DuckDuckGo 搜索工具
		config := &duckduckgo.Config{
			MaxResults: 3,
			Region:     ddgsearch.RegionCN,
			DDGConfig: &ddgsearch.Config{
				Cache:      true,
				MaxRetries: 5,
			},
		}
		searchTool, err := duckduckgo.NewTool(ctx, config)
		if err != nil {
			log.Println("搜索工具初始化失败:", err)
			continue
		}

		searchReq := &duckduckgo.SearchRequest{
			Query: input,
			Page:  1,
		}
		jsonReq, err := json.Marshal(searchReq)
		if err != nil {
			log.Fatalf("搜索请求序列化失败: %v", err)
		}

		resp, err := searchTool.InvokableRun(ctx, string(jsonReq))
		if err != nil {
			log.Println("搜索失败:", err)
			continue
		}

		var searchResp duckduckgo.SearchResponse
		if err := json.Unmarshal([]byte(resp), &searchResp); err != nil {
			log.Println("解析搜索结果失败:", err)
			continue
		}

		sources := make([]string, 0, len(searchResp.Results))

		// 使用默认配置初始化加载器
		loader, err := url.NewLoader(ctx, nil)
		if err != nil {
			panic(err)
		}

		for _, result := range searchResp.Results {
			docs, err := loader.Load(ctx, document.Source{
				URI: result.Link,
			})
			if err != nil {
				panic(err)
			}
			for _, doc := range docs {
				content := doc.Content
				sources = append(sources, content)
			}
		}

		fmt.Println("正在分析，请稍候...")
		response, err := master.Analyze(ctx, input, sources)
		if err != nil {
			log.Println("\n分析失败:", err)
			continue
		}
		_ = response
	}
}
