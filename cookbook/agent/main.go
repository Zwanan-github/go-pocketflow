package main

import (
	"fmt"
	"os"

	pocketflow "github.com/zwanan-github/go-pocketflow"
)

// 研究助手 Agent
//
// 使用 go-pocketflow 实现的 Agent 模式：
//
//	          ┌─ "search" → search ─("decide")─┐
//	decide ──┤                                  decide (循环)
//	          └─ "answer" → answer → (结束)
//
// 环境变量：
//   - LLM_API_KEY:  LLM API 密钥（必填）
//   - LLM_BASE_URL: API 地址（默认 https://api.openai.com/v1）
//   - LLM_MODEL:    模型名称（默认 gpt-4o-mini）

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run . <你的问题>")
		fmt.Println("示例: go run . \"Go 1.22 有哪些新特性？\"")
		os.Exit(1)
	}
	question := os.Args[1]

	// 构建节点
	decide := newDecideNode()
	search := newSearchNode()
	answer := newAnswerNode()

	// 连接节点
	decide.ThenWith("search", search) // 决策 → 搜索
	decide.ThenWith("answer", answer) // 决策 → 回答
	search.ThenWith("decide", decide) // 搜索 → 回到决策

	// 运行
	shared := pocketflow.SharedStore{
		"question": question,
	}

	fmt.Printf("问题: %s\n\n", question)

	flow := pocketflow.NewFlow(decide)
	flow.Run(shared)

	fmt.Printf("\n=== 最终回答 ===\n%s\n", shared["answer"])
}
