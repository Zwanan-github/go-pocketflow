package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func init() {
	// 从 .env 文件加载配置，不覆盖已有环境变量
	_ = godotenv.Load()
}

// callLLM 调用 OpenAI 兼容的 LLM API
// 配置项从 .env 文件或环境变量读取：
//   - LLM_BASE_URL: API 地址（默认 http://napi.zwanan.top/v1）
//   - LLM_API_KEY:  API 密钥
//   - LLM_MODEL:    模型名称（默认 gpt-4o-mini）
func callLLM(prompt string) string {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		panic("请在 .env 文件中设置 LLM_API_KEY")
	}

	config := openai.DefaultConfig(apiKey)
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "http://napi.zwanan.top/v1"
	}
	config.BaseURL = baseURL

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = openai.GPT4oMini
	}

	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("LLM 请求失败: %v", err))
	}
	if len(resp.Choices) == 0 {
		panic(fmt.Sprintf("LLM 返回空响应, model=%s, base_url=%s", model, baseURL))
	}
	return resp.Choices[0].Message.Content
}

// searchWeb 使用 Tavily 搜索 API
// 配置项：TAVILY_API_KEY（从 https://tavily.com 获取）
func searchWeb(query string) string {
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		panic("请在 .env 文件中设置 TAVILY_API_KEY")
	}

	reqBody, _ := json.Marshal(map[string]any{
		"query":          query,
		"api_key":        apiKey,
		"max_results":    5,
		"include_answer": true,
	})

	resp, err := http.Post("https://api.tavily.com/search", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Sprintf("搜索失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Answer  string `json:"answer"`
		Results []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Sprintf("搜索结果解析失败: %v", err)
	}

	var sb strings.Builder
	// Tavily 自带的 AI 摘要
	if result.Answer != "" {
		sb.WriteString("摘要: " + result.Answer + "\n\n")
	}
	for i, r := range result.Results {
		sb.WriteString(fmt.Sprintf("[%d] %s\n%s\n\n", i+1, r.Title, r.Content))
	}

	if sb.Len() == 0 {
		return "未找到相关搜索结果"
	}
	return sb.String()
}
