package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	searchTool, err := NewSearchTool(ctx)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{
		searchTool,
	}, nil
}

func NewSearchTool(ctx context.Context) (tn tool.BaseTool, err error) {
	tn, err = newTool(ctx)
	if err != nil {
		return nil, err
	}
	return tn, nil
}

type ToolImpl struct {
	config *ToolConfig
}

type ToolConfig struct {
}

func newTool(ctx context.Context) (bt tool.BaseTool, err error) {
	// TODO Modify component configuration here.
	config := &ToolConfig{}
	bt = &ToolImpl{config: config}
	return bt, nil
}

func (impl *ToolImpl) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "SearchTool",
		Desc: "用来搜集问题相关网页的URL,并从返回的 URL 加载网页内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "搜索关键词",
			},
		}),
	}, nil
}

type SearchQuery struct {
	Query string `json:"query"`
}

func (impl *ToolImpl) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	fmt.Println("正在搜索...")
	var query SearchQuery
	err := json.Unmarshal([]byte(argumentsInJSON), &query)
	if err != nil {
		return "", err
	}
	config := &duckduckgo.Config{
		ToolName:   "duckduckgo_search",
		ToolDesc:   "用来搜集问题相关网页的信息",
		MaxResults: 3,
		Region:     ddgsearch.RegionCN,
		DDGConfig: &ddgsearch.Config{
			Cache:      true,
			MaxRetries: 5,
		},
	}
	searchTool, err := duckduckgo.NewTool(ctx, config)
	if err != nil {
		log.Printf("创建搜索工具失败: %v", err)
		return "", err
	}

	searchReq := &duckduckgo.SearchRequest{
		Query: query.Query,
		Page:  1,
	}
	jsonReq, err := json.Marshal(searchReq)
	if err != nil {
		log.Fatalf("搜索请求序列化失败: %v", err)
		return "", err
	}

	resp, err := searchTool.InvokableRun(ctx, string(jsonReq))
	if err != nil {
		log.Println("搜索失败:", err)
		return "", err
	}

	var searchResp duckduckgo.SearchResponse
	if err := json.Unmarshal([]byte(resp), &searchResp); err != nil {
		log.Println("解析搜索结果失败:", err)
		return "", err
	}

	sources := make([]string, 0, len(searchResp.Results))

	for _, result := range searchResp.Results {
		fmt.Println(result.Link)
		docs := extractMainContent(result.Link)
		if err != nil {
			log.Printf("加载网页内容失败: %v", err)
			continue
		}
		sources = append(sources, docs)
	}

	out := strings.Join(sources, "\n")
	return out, nil
}

// 增强版内容提取
func extractMainContent(url string) string {
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}

	var resp *http.Response
	var err error

	// 重试逻辑
	for retry := 0; retry < 3; retry++ {
		resp, err = client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(retry+1) * 500 * time.Millisecond)
	}
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}

	// 智能内容提取
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	content := findMainContent(doc)
	return strings.TrimSpace(content)
}

func findMainContent(doc *goquery.Document) string {
	// 优先查找标准语义标签
	selectors := []string{
		"article", "main", "[role='main']",
		".post-content", ".article-body",
		"#content", "#main-content",
	}

	for _, selector := range selectors {
		if content := extractBySelector(doc, selector); content != "" {
			return content
		}
	}

	// 回退策略：查找最长文本块
	var maxLength int
	var mainContent string
	doc.Find("div, section").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if len(text) > maxLength {
			maxLength = len(text)
			mainContent = text
		}
	})
	return mainContent
}

func extractBySelector(doc *goquery.Document, selector string) string {
	var content strings.Builder
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		s.Find("p, li, pre").Each(func(_ int, el *goquery.Selection) {
			text := strings.TrimSpace(el.Text())
			if len(text) > 50 {
				content.WriteString(text + "\n")
			}
		})
	})
	return content.String()
}
