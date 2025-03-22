package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino/components/document"
	"log"
	"regexp"
	"strings"

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

	// 使用默认配置初始化加载器
	loader, err := url.NewLoader(ctx, nil)
	if err != nil {
		log.Printf("初始化加载器失败: %v", err)
		return "", err
	}

	for _, result := range searchResp.Results {
		docs, err := loader.Load(ctx, document.Source{
			URI: result.Link,
		})
		if err != nil {
			log.Printf("加载网页内容失败: %v", err)
			continue
		}
		for _, doc := range docs {
			content := doc.Content
			sources = append(sources, content)
		}
	}

	out := strings.Join(sources, "\n")
	out = cleanText(out)
	return out, nil
}

// Pre-compiled regular expressions for better performance
var (
	// HTML element removal patterns
	scriptPattern     = regexp.MustCompile(`(?i)<script\b[^>]*>[\s\S]*?</script>`)
	inlineJSPattern   = regexp.MustCompile(`(?i)on\w+\s*=\s*["'].*?["']`)
	jsFunctionPattern = regexp.MustCompile(`(?i)javascript:\s*\w+\([^)]*\)`)
	stylePattern      = regexp.MustCompile(`<style\b[^>]*>[\s\S]*?</style>`)
	iframePattern     = regexp.MustCompile(`<iframe\b[^>]*>[\s\S]*?</iframe>`)
	noscriptPattern   = regexp.MustCompile(`<noscript\b[^>]*>[\s\S]*?</noscript>`)
	footerPattern     = regexp.MustCompile(`<footer\b[^>]*>[\s\S]*?</footer>`)
	headerPattern     = regexp.MustCompile(`<header\b[^>]*>[\s\S]*?</header>`)
	navPattern        = regexp.MustCompile(`<nav\b[^>]*>[\s\S]*?</nav>`)
	divIDPattern      = regexp.MustCompile(`<div\b[^>]*id=["']?(ad|banner|footer|header|nav|sidebar|copyright|menu|comment)["']?[^>]*>[\s\S]*?</div>`)
	divClassPattern   = regexp.MustCompile(`<div\b[^>]*class=["']?(ad|banner|footer|header|nav|sidebar|copyright|menu|comment)["']?[^>]*>[\s\S]*?</div>`)

	// Metadata extraction patterns
	titlePattern    = regexp.MustCompile(`(?i)"title"\s*:\s*"([^"<]+)"|<title[^>]*>([^<]+)</title>`)
	timePattern     = regexp.MustCompile(`(?i)postTime\s*=\s*"([^"]+)"|datetime="([^"]+)"|pubdate="([^"]+)"`)
	authorPattern   = regexp.MustCompile(`(?i)author"\s*:\s*"([^"]+)"|发布\s*[：:]\s*([^\s<]+)|作者\s*[：:]\s*([^\s<]+)`)
	keywordsPattern = regexp.MustCompile(`(?i)keywords\s*=\s*\[([^\]]+)\]|<meta\s+name=["']keywords["']\s+content=["']([^"']+)`)

	// HTML tag and whitespace patterns
	htmlTagPattern    = regexp.MustCompile(`<[^>]*>`)
	whitespacePattern = regexp.MustCompile(`\s{2,}`)
	newlinePattern    = regexp.MustCompile(`\n{3,}`)

	// HTML entities map
	htmlEntities = map[string]string{
		"&nbsp;":   " ",
		"&amp;":    "&",
		"&lt;":     "<",
		"&gt;":     ">",
		"&quot;":   "\"",
		"&apos;":   "'",
		"&copy;":   "©",
		"&reg;":    "®",
		"&trade;":  "™",
		"&mdash;":  "—",
		"&ndash;":  "–",
		"&hellip;": "…",
	}
)

// cleanText removes unnecessary HTML elements and extracts useful content
func cleanText(input string) string {

	// 1. Remove irrelevant elements
	cleaned := removeIrrelevantElements(input)

	// 2. Extract metadata
	metadata := extractMetadata(cleaned)

	// 3. Remove HTML tags but keep content
	contentOnly := htmlTagPattern.ReplaceAllString(cleaned, " ")

	// 4. Process HTML entities and whitespace
	contentOnly = processEntitiesAndWhitespace(contentOnly)

	// 6. Format final output
	var result strings.Builder
	if len(metadata) > 0 {
		result.WriteString(metadata)
		result.WriteString("\n--- 内容摘要 ---\n")
	}
	result.WriteString(contentOnly)

	return result.String()
}

// removeIrrelevantElements removes scripts, ads, and other unnecessary HTML elements
func removeIrrelevantElements(text string) string {
	text = scriptPattern.ReplaceAllString(text, "")
	text = inlineJSPattern.ReplaceAllString(text, "")
	text = jsFunctionPattern.ReplaceAllString(text, "")
	text = stylePattern.ReplaceAllString(text, "")
	text = iframePattern.ReplaceAllString(text, "")
	text = noscriptPattern.ReplaceAllString(text, "")
	text = footerPattern.ReplaceAllString(text, "")
	text = headerPattern.ReplaceAllString(text, "")
	text = navPattern.ReplaceAllString(text, "")
	text = divIDPattern.ReplaceAllString(text, "")
	text = divClassPattern.ReplaceAllString(text, "")
	return text
}

// extractMetadata finds and formats title, author, time, keywords
func extractMetadata(text string) string {
	metadataPatterns := map[string]*regexp.Regexp{
		"title":    titlePattern,
		"time":     timePattern,
		"author":   authorPattern,
		"keywords": keywordsPattern,
	}

	var metadata strings.Builder
	for name, re := range metadataPatterns {
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			for i := 1; i < len(matches); i++ {
				if value := strings.TrimSpace(matches[i]); value != "" {
					metadata.WriteString(fmt.Sprintf("%s: %s\n", strings.Title(name), value))
					break
				}
			}
		}
	}

	return metadata.String()
}

// processEntitiesAndWhitespace replaces HTML entities and normalizes spacing
func processEntitiesAndWhitespace(text string) string {
	// Replace HTML entities
	for entity, replacement := range htmlEntities {
		text = strings.ReplaceAll(text, entity, replacement)
	}

	// Normalize whitespace
	text = whitespacePattern.ReplaceAllString(text, " ")
	text = newlinePattern.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}
