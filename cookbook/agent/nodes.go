package main

import (
	"fmt"
	"strings"

	pocketflow "github.com/zwanan-github/go-pocketflow"
)

// Decision 表示 LLM 的决策结果
type Decision struct {
	Action      string // "search" 或 "answer"
	Answer      string // action=answer 时的回答
	SearchQuery string // action=search 时的搜索词
}

// maxSearchRounds 最大搜索轮数，超过后强制回答
const maxSearchRounds = 3

// newDecideNode 创建决策节点
// 调用 LLM 判断下一步：搜索更多信息，还是直接回答
func newDecideNode() *pocketflow.Node {
	return pocketflow.NewNode("decide").
		Prep(func(shared pocketflow.SharedStore) any {
			question := shared["question"].(string)
			context, _ := shared["context"].(string)
			if context == "" {
				context = "暂无搜索记录"
			}
			// 统计搜索次数
			rounds, _ := shared["search_rounds"].(int)
			return []any{question, context, rounds}
		}).
		Exec(func(prepRes any) any {
			inputs := prepRes.([]any)
			question := inputs[0].(string)
			context := inputs[1].(string)
			rounds := inputs[2].(int)

			// 超过最大搜索次数，强制回答
			if rounds >= maxSearchRounds {
				fmt.Printf("已搜索 %d 次，根据已有信息回答\n", rounds)
				return Decision{Action: "answer"}
			}

			fmt.Println("正在思考下一步...")

			prompt := fmt.Sprintf(`### 上下文
你是一个研究助手，可以搜索网页获取信息。

问题: %s
已有研究: %s
已搜索次数: %d（最多 %d 次）

### 可用操作

[1] search
  描述: 在网上搜索更多信息
  参数: search_query - 搜索关键词

[2] answer
  描述: 根据已有信息回答问题
  参数: answer - 最终答案

### 决策
根据上下文和可用操作，决定下一步。
如果已有足够信息或搜索次数接近上限，请直接回答。
用以下格式回复（纯文本，不要用代码块包裹）：

action: search 或 answer
reason: 选择该操作的原因
answer: 如果 action 是 answer，写出完整回答
search_query: 如果 action 是 search，写出搜索关键词`, question, context, rounds, maxSearchRounds)

			response := callLLM(prompt)
			return parseDecision(response)
		}).
		Post(func(shared pocketflow.SharedStore, _, execRes any) string {
			d := execRes.(Decision)
			if d.Action == "search" {
				shared["search_query"] = d.SearchQuery
				fmt.Printf("决定搜索: %s\n", d.SearchQuery)
			} else {
				shared["context"] = d.Answer
				fmt.Println("决定直接回答")
			}
			return d.Action
		})
}

// newSearchNode 创建搜索节点
// 执行网页搜索，将结果追加到上下文，然后回到决策节点
func newSearchNode() *pocketflow.Node {
	return pocketflow.NewNode("search").
		Prep(func(shared pocketflow.SharedStore) any {
			return shared["search_query"].(string)
		}).
		Exec(func(prepRes any) any {
			query := prepRes.(string)
			fmt.Printf("正在搜索: %s\n", query)
			return searchWeb(query)
		}).
		Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
			query := prepRes.(string)
			results := execRes.(string)
			previous, _ := shared["context"].(string)
			shared["context"] = previous + "\n\n搜索: " + query + "\n结果: " + results
			// 递增搜索次数
			rounds, _ := shared["search_rounds"].(int)
			shared["search_rounds"] = rounds + 1
			fmt.Println("已获取搜索结果，继续分析...")
			// 回到决策节点
			return "decide"
		})
}

// newAnswerNode 创建回答节点
// 根据已有上下文，调用 LLM 生成最终回答
func newAnswerNode() *pocketflow.Node {
	return pocketflow.NewNode("answer").
		Prep(func(shared pocketflow.SharedStore) any {
			question := shared["question"].(string)
			context, _ := shared["context"].(string)
			return []string{question, context}
		}).
		Exec(func(prepRes any) any {
			inputs := prepRes.([]string)
			question, context := inputs[0], inputs[1]

			fmt.Println("正在生成最终回答...")

			prompt := fmt.Sprintf(`### 上下文
根据以下信息回答问题。

问题: %s
研究资料: %s

### 回答
请基于研究结果给出完整、准确的回答。`, question, context)

			return callLLM(prompt)
		}).
		Post(func(shared pocketflow.SharedStore, _, execRes any) string {
			shared["answer"] = execRes.(string)
			fmt.Println("回答生成完毕")
			return "done"
		})
}

// parseDecision 解析 LLM 返回的决策文本
func parseDecision(response string) Decision {
	d := Decision{Action: "answer"}
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "action:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "action:"))
			if val == "search" || val == "answer" {
				d.Action = val
			}
		} else if strings.HasPrefix(line, "answer:") {
			d.Answer = strings.TrimSpace(strings.TrimPrefix(line, "answer:"))
		} else if strings.HasPrefix(line, "search_query:") {
			d.SearchQuery = strings.TrimSpace(strings.TrimPrefix(line, "search_query:"))
		}
	}
	return d
}
