package main

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// newLambda1 component initialization function of node 'ConveyMap' in graph 'einoLLM'
func newLambda1(ctx context.Context, input string) (output map[string]any, err error) {
	return map[string]any{
		"question": input,
	}, nil
}

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

// 分别存储前端和后端的历史记录
var (
	BackendHistory  []*schema.Message
	FrontendHistory []*schema.Message
)

// newLambda2 component initialization function of node 'ConveyMap1' in graph 'einoLLM'
func newLambda2(ctx context.Context, input *schema.Message) (output map[string]any, err error) {
	return map[string]any{
		"role":         "资深后端架构师",
		"style":        "技术面试官视角",
		"question":     input.Content,
		"chat_history": BackendHistory,
		"examples":     Examples,
	}, nil
}

var Examples2 = []*schema.Message{
	schema.UserMessage(`React Hooks 和 Class 组件的区别是什么？`),
	schema.AssistantMessage(
		`1. 核心考点：React Hooks 的设计理念和优势
2. 实际案例：某电商平台从 Class 组件迁移到 Hooks 的性能提升
3. 最佳实践：
   - 使用 useState 管理状态
   - useEffect 处理副作用
   - 自定义 Hooks 封装业务逻辑
4. 性能优化：如何避免不必要的重渲染，使用 useMemo 和 useCallback
5. 深入追问：React 的 Fiber 架构如何影响性能？`, nil),
}

// newLambda3 component initialization function of node 'ConveyMap1' in graph 'einoLLM'
func newLambda3(ctx context.Context, input *schema.Message) (output map[string]any, err error) {
	return map[string]any{
		"role":         "资深前端架构师",
		"style":        "技术面试官视角",
		"question":     input.Content,
		"chat_history": FrontendHistory,
		"examples":     Examples2,
	}, nil
}
