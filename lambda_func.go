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

var History []*schema.Message

// newLambda2 component initialization function of node 'ConveyMap1' in graph 'einoLLM'
func newLambda2(ctx context.Context, input *schema.Message) (output map[string]any, err error) {
	return map[string]any{
		"role":         "资深后端架构师",
		"style":        "技术面试官视角",
		"question":     input.Content,
		"chat_history": History,
		"examples":     Examples,
	}, nil
}
