package main

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var (
	SystemMessageTemplate = `你是一个智能搜索优化助手，专门将用户的提问转换成更精准、更容易搜索到相关资料的关键词或查询语句。转换要求：
1. 提取问题的核心关键词  
2. 采用简洁、直接的搜索表达方式  
3. 根据问题类型优化查询方式（如使用更具体的技术术语、常见问题格式等）  
4. 避免过于主观或模糊的描述，确保高效检索  
5. 输出优化后的搜索查询语句，保持自然流畅`

	UserMessageTemplate = `请优化以下问题，使其更适合搜索引擎查询：
【原始问题】{question}  // 用户的原始问题
【优化查询】请转换成简洁、精准的搜索关键词或查询语句`
)

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate component initialization function of node 'ChatTemplate1' in graph 'einoLLM'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	config := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(SystemMessageTemplate1),
			schema.MessagesPlaceholder("examples", true),
			schema.MessagesPlaceholder("chat_history", false),
			schema.UserMessage(UserMessageTemplate1),
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}

var (
	SystemMessageTemplate1 = `作为{role}，你需要以{style}风格进行面试答疑，要求：  
1. 结合真实企业面试场景  
2. 准确识别候选人技术短板  
3. 解析核心考点及其深入考察方式  
4. 结合实际应用场景提供最佳回答策略  
5. 使用分层解析法：基础概念 → 核心原理 → 进阶考察 → 最佳解法`

	UserMessageTemplate1 = `后端技术面试答疑请求：
【问题描述】{question}
【回答要求】请按以下结构回答：
1. 核心考点解析
2. 真实企业面试案例
3. 最优回答策略与示例
4. 面试官深入追问方向`
)

type ChatTemplate1Config struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate1 component initialization function of node 'SearchTemplate' in graph 'einoLLM'
func newChatTemplate1(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	config := &ChatTemplate1Config{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(SystemMessageTemplate),
			schema.UserMessage(UserMessageTemplate),
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
