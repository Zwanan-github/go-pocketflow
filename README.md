# go-pocketflow

> Go 版本的 PocketFlow - 极简的 LLM 应用框架
>
> 参考: [PocketFlow](https://github.com/The-Pocket/PocketFlow)

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 特性

- 🎯 **极简设计** - 只有 Node 和 Flow 两个核心概念
- 🔗 **流畅 API** - 链式调用，编写体验优雅
- 🌿 **条件分支** - 轻松实现 Agent 模式
- 📦 **批量处理** - 内置 BatchNode / BatchFlow 支持
- ⚡ **异步执行** - AsyncNode / AsyncBatchNode 并行处理

## 快速开始

```go
package main

import "github.com/zwanan-github/go-pocketflow"

func main() {
    node := pocketflow.NewNode("test").
        Prep(func(shared pocketflow.SharedStore) any {
            return shared["input"]
        }).
        Exec(func(prepRes any) any {
            return prepRes.(string) + " processed"
        }).
        Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
            shared["output"] = execRes
            return "end"
        })

    shared := pocketflow.SharedStore{"input": "test data"}
    node.Run(shared)
}
```

## 核心概念

### Node（节点）

节点是流程的基本处理单元，包含三个生命周期阶段：

```go
type Node struct {
    name       string
    prep       PrepFunc              // 准备阶段：从 shared 读取数据
    exec       ExecFunc              // 执行阶段：处理业务逻辑
    post       PostFunc              // 后置阶段：保存结果，决定下一步
    successors map[string]*Node      // 后继节点映射
    params     map[string]any        // 节点参数
}
```

### Flow（流程）

流程将多个节点串联，按 post 返回的 action 决定下一个节点：

```go
node1 := pocketflow.NewNode("step1").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        return "next"
    })

node2 := pocketflow.NewNode("step2").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        return "end"
    })

node1.ThenWith("next", node2)

flow := pocketflow.NewFlow(node1)
flow.Run(shared)
```

### BatchNode（批量节点）

Prep 返回切片，Exec 对每个元素串行执行：

```go
batchNode := pocketflow.NewBatchNode("batch").
    Prep(func(shared pocketflow.SharedStore) any {
        return []string{"item1", "item2", "item3"}
    }).
    Exec(func(item any) any {
        return "processed-" + item.(string)
    }).
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        shared["results"] = execRes
        return "success"
    })
```

### BatchFlow（批量流程）

对同一个 Flow 使用不同参数批量执行：

```go
batchFlow := pocketflow.NewBatchFlow(startNode).
    Prep(func(shared pocketflow.SharedStore) []map[string]any {
        return []map[string]any{
            {"item": "item1"},
            {"item": "item2"},
        }
    })

batchFlow.Run(shared)
```

## 异步组件

异步组件在 goroutine 中执行，返回 `<-chan AsyncResult`，不阻塞调用者。

### AsyncNode

```go
asyncNode := pocketflow.NewAsyncNode("async-test").
    Prep(func(shared pocketflow.SharedStore) any {
        return shared["input"]
    }).
    Exec(func(prepRes any) any {
        time.Sleep(100 * time.Millisecond)
        return prepRes.(string) + " processed"
    }).
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        shared["output"] = execRes
        return "success"
    })

shared := pocketflow.SharedStore{"input": "test data"}
resultChan := asyncNode.RunAsync(shared)

select {
case result := <-resultChan:
    fmt.Printf("action: %s\n", result.Action)
case <-time.After(time.Second):
    fmt.Println("超时")
}
```

### AsyncBatchNode

Exec 阶段并行处理每个 item，结果保序：

```go
batchNode := pocketflow.NewBatchNode("batch").
    Prep(func(shared pocketflow.SharedStore) any {
        return []string{"item1", "item2", "item3"}
    }).
    Exec(func(item any) any {
        return "processed-" + item.(string)
    }).
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        shared["results"] = execRes
        return "success"
    })

asyncBatchNode := pocketflow.NewAsyncBatchNode(batchNode)
resultChan := asyncBatchNode.RunAsync(shared)
```

### AsyncFlow

```go
node1 := pocketflow.NewAsyncNode("step1").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        return "next"
    })

node2 := pocketflow.NewNode("step2").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        return "end"
    })

node1.ThenWith("next", node2)

asyncFlow := pocketflow.NewAsyncFlow(node1)
resultChan := asyncFlow.RunAsync(shared)
```

### AsyncBatchFlow

每个 batch 并行执行整个 Flow：

```go
asyncBatchFlow := pocketflow.NewAsyncBatchFlow(batchFlow)
resultChan := asyncBatchFlow.RunAsync(shared)
```

### 同步 vs 异步

| 同步 | 异步 | 区别 |
|------|------|------|
| `Node` | `AsyncNode` | goroutine 中执行 |
| `BatchNode` | `AsyncBatchNode` | Exec 阶段并行处理 |
| `Flow` | `AsyncFlow` | goroutine 中执行 |
| `BatchFlow` | `AsyncBatchFlow` | 每个 batch 并行执行 |

## 条件分支

通过 Post 返回不同 action 实现分支：

```go
decide := pocketflow.NewNode("decide").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        if needAction {
            return "action"
        }
        return "done"
    })

handler := pocketflow.NewNode("handler")
done := pocketflow.NewNode("done")

decide.ThenWith("action", handler)
decide.ThenWith("done", done)
handler.Then(decide) // 循环回 decide，实现 Agent 模式
```

## 常见模式

| 模式 | 实现方式 |
|------|----------|
| **Workflow** | `node1.Then(node2).Then(node3)` |
| **Agent** | 条件分支：`decide.ThenWith("action", handler)` |
| **RAG** | Prep 检索 → Exec 生成 → Post 存储 |
| **MapReduce** | BatchFlow(map) → Node(reduce) |

## 与 Python 版本对比

| Python | Go |
|--------|-----|
| `node1 >> node2` | `node1.Then(node2)` |
| `node1 - "action" >> node2` | `node1.ThenWith("action", node2)` |
| 类继承 | 函数式风格 |
| `def prep(self, shared)` | `node.Prep(func(shared) any { ... })` |

## 架构优势

- ✅ **极简主义** - 最少的抽象覆盖最多场景
- ✅ **类型安全** - Go 的静态类型保证
- ✅ **易测试** - 纯函数，依赖注入
- ✅ **Go 风格** - 符合 Go 语言习惯

## 测试

```bash
go test -v
```

## 许可证

MIT License