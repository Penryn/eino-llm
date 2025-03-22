package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"os"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// 模型基本信息
var (
	ModelType   string // 模型名称
	OwnerAPIKey string // apiKey
	BaseURL     string // 模型api地址

	arkAPIKey    string // ark模型apiKey
	arkModelName string // ark模型名称
)

func buildeinoLLM(ctx context.Context) (r compose.Runnable[string, *schema.Message], err error) {
	const (
		agent          = "agent"
		ChatTemplate1  = "ChatTemplate1"
		ChatModel2     = "ChatModel2"
		SearchTemplate = "SearchTemplate"
		ConveyMap      = "ConveyMap"
		ConveyMap1     = "ConveyMap1"
	)
	err = godotenv.Load()
	if err != nil {
		log.Fatal("加载 .env 文件出错")
	}

	ModelType = os.Getenv("Model_Type")
	OwnerAPIKey = os.Getenv("Owner_API_Key")
	BaseURL = os.Getenv("Base_URL")
	arkAPIKey = os.Getenv("ARK_API_KEY")
	arkModelName = os.Getenv("ARK_MODEL_NAME")

	g := compose.NewGraph[string, *schema.Message]()
	agentKeyOfLambda, err := newLambda(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(agent, agentKeyOfLambda)
	chatTemplate1KeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate1, chatTemplate1KeyOfChatTemplate)
	chatModel2KeyOfChatModel, err := newChatModel1(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(ChatModel2, chatModel2KeyOfChatModel)
	searchTemplateKeyOfChatTemplate, err := newChatTemplate1(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(SearchTemplate, searchTemplateKeyOfChatTemplate)
	_ = g.AddLambdaNode(ConveyMap, compose.InvokableLambda(newLambda1))
	_ = g.AddLambdaNode(ConveyMap1, compose.InvokableLambda(newLambda2))
	_ = g.AddEdge(compose.START, ConveyMap)
	_ = g.AddEdge(ChatModel2, compose.END)
	_ = g.AddEdge(SearchTemplate, agent)
	_ = g.AddEdge(agent, ConveyMap1)
	_ = g.AddEdge(ConveyMap1, ChatTemplate1)
	_ = g.AddEdge(ChatTemplate1, ChatModel2)
	_ = g.AddEdge(ConveyMap, SearchTemplate)
	r, err = g.Compile(ctx, compose.WithGraphName("einoLLM"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
