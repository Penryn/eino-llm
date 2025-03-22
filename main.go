package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	ctx := context.Background()
	for {
		input := getMultilineInput()
		if input == "" {
			return
		}
		if input == "exit" {
			return
		}

		fmt.Println("\n分析中...")

		_, err := runAgent(ctx, input)
		if err != nil {
			log.Printf("[Chat] Error running agent: %v\n", err)
			return
		}
	}

}

func runAgent(ctx context.Context, msg string) (string, error) {
	runner, err := buildeinoLLM(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to build agent graph: %w", err)
	}

	sr, err := runner.Stream(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("failed to stream: %w", err)
	}
	defer sr.Close()

	var fullResponse strings.Builder
	for {
		resp, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			History = append(History, schema.AssistantMessage(fullResponse.String(), nil))
			return fullResponse.String(), nil
		}
		if err != nil {
			return "", fmt.Errorf("流式处理异常: %w", err)
		}
		fmt.Print(resp.Content)
		fullResponse.WriteString(resp.Content)
	}

}
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
