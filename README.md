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
- 📦 **批量处理** - 内置 BatchFlow 支持

## 快速开始

### 基础节点

```go
package main

import (
    "fmt"
    pocketflow "github.com/zwanan-github/go-pocketflow"
)

func main() {
    // 创建测试节点
    node := pocketflow.NewNode("test").
        Prep(func(shared pocketflow.SharedStore) any {
            return shared["input"]  // 准备：读取输入
        }).
        Exec(func(prepRes any) any {
            return prepRes.(string) + " processed"  // 执行：处理数据
        }).
        Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
            shared["output"] = execRes  // 后置：保存结果
            return "end"
        })

    // 运行节点
    shared := pocketflow.SharedStore{"input": "test data"}
    node.Run(shared)

    fmt.Println(shared["output"])  // 输出：test data processed
}
```

### 流程编排

```go
// 创建节点链
node1 := pocketflow.NewNode("node1").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        shared["step1"] = "done"
        return "next"
    })

node2 := pocketflow.NewNode("node2").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        shared["step2"] = "done"
        return "end"
    })

// 连接节点
node1.Then(node2)

// 创建并运行流程
flow := pocketflow.NewFlow(node1)
shared := pocketflow.SharedStore{}
flow.Run(shared)

fmt.Printf("Step1: %v, Step2: %v\n", shared["step1"], shared["step2"])
```

### Agent 模式（条件分支）

```go
// 决策节点
decide := pocketflow.NewNode("decide").
    Prep(func(shared pocketflow.SharedStore) any {
        return shared["value"]
    }).
    Exec(func(prepRes any) any {
        value := prepRes.(int)
        if value > 10 {
            return "large"
        }
        return "small"
    }).
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        return execRes.(string)
    })

// 处理节点
largeHandler := pocketflow.NewNode("large").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        shared["result"] = "handled large value"
        return "end"
    })

smallHandler := pocketflow.NewNode("small").
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        shared["result"] = "handled small value"
        return "end"
    })

// 设置分支
decide.ThenWith("large", largeHandler)
decide.ThenWith("small", smallHandler)

// 创建 Agent 流程
agentFlow := pocketflow.NewFlow(decide)

// 测试小值
shared1 := pocketflow.SharedStore{"value": 5}
agentFlow.Run(shared1)
fmt.Println(shared1["result"])  // 输出：handled small value

// 测试大值
shared2 := pocketflow.SharedStore{"value": 15}
agentFlow.Run(shared2)
fmt.Println(shared2["result"])  // 输出：handled large value
```

### 批量处理

```go
var results []string

// 处理节点
processNode := pocketflow.NewNode("process").
    Prep(func(shared pocketflow.SharedStore) any {
        return shared["item"]
    }).
    Exec(func(prepRes any) any {
        return prepRes.(string) + "_processed"
    }).
    Post(func(shared pocketflow.SharedStore, prepRes, execRes any) string {
        results = append(results, execRes.(string))
        return "done"
    })

// 批量流程
batchFlow := pocketflow.NewBatchFlow(processNode).
    Prep(func(shared pocketflow.SharedStore) []map[string]any {
        items := shared["items"].([]string)
        params := make([]map[string]any, len(items))
        for i, item := range items {
            params[i] = map[string]any{"item": item}
        }
        return params
    })

// 运行批量处理
batchShared := pocketflow.SharedStore{
    "items": []string{"item1", "item2", "item3"},
}
batchFlow.Run(batchShared)

fmt.Println(results)  // 输出：[item1_processed item2_processed item3_processed]
```

## 核心概念

### Node（节点）

节点是流程的基本处理单元，包含三个生命周期阶段：

```go
type Node struct {
    name       string
    prep       PrepFunc     // 准备阶段：从 shared 读取数据
    exec       ExecFunc     // 执行阶段：处理业务逻辑
    post       PostFunc     // 后置阶段：保存结果，决定下一步
    successors map[string]*Node  // 后继节点映射
    params     map[string]any     // 节点参数
}
```

- **Prep**：准备阶段，从 SharedStore 读取数据
- **Exec**：执行阶段，处理业务逻辑
- **Post**：后置阶段，保存结果并决定下一步动作

### Flow（流程）

流程负责编排节点的执行顺序：

```go
type Flow struct {
    startNode *Node  // 起始节点
}
```

### SharedStore（共享存储）

节点间通过 SharedStore 共享数据：

```go
type SharedStore map[string]any
```

### BatchFlow（批量流程）

批量流程支持对多个数据集执行相同的节点逻辑：

```go
type BatchFlow struct {
    *Flow
    prep BatchPrepFunc  // 批量准备函数
}
```

## API 参考

### Node 方法

| 方法 | 描述 |
|------|------|
| `NewNode(name string) *Node` | 创建新节点 |
| `Prep(fn PrepFunc) *Node` | 设置准备函数 |
| `Exec(fn ExecFunc) *Node` | 设置执行函数 |
| `Post(fn PostFunc) *Node` | 设置后置函数 |
| `Then(next *Node) *Node` | 连接后继节点（默认 action） |
| `ThenWith(action string, next *Node) *Node` | 连接带特定 action 的后继节点 |
| `SetParams(params map[string]any) *Node` | 设置节点参数 |
| `GetParam(key string) (any, bool)` | 获取节点参数 |
| `Clone() *Node` | 克隆节点（状态隔离） |

### Flow 方法

| 方法 | 描述 |
|------|------|
| `NewFlow(start *Node) *Flow` | 创建新流程 |
| `Run(shared SharedStore) string` | 运行流程 |

### BatchFlow 方法

| 方法 | 描述 |
|------|------|
| `NewBatchFlow(start *Node) *BatchFlow` | 创建批量流程 |
| `Prep(fn BatchPrepFunc) *BatchFlow` | 设置批量准备函数 |
| `Run(shared SharedStore) string` | 运行批量流程 |

## 设计模式

| 模式 | 实现方式 |
|------|----------|
| **Workflow** | `node1.Then(node2).Then(node3)` |
| **Agent** | 条件分支：`decide.ThenWith("action", handler)` |
| **RAG** | Prep 检索 → Exec 生成 → Post 存储 |
| **MapReduce** | BatchFlow(map) → Node(reduce) |
| **批量处理** | BatchFlow 支持并行/串行批量执行 |

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
- ✅ **高性能** - 无反射，编译优化
- ✅ **易测试** - 纯函数，依赖注入
- ✅ **Go 风格** - 符合 Go 语言习惯
- ✅ **状态隔离** - 每次运行独立，避免副作用

## 文件结构

```
go-pocketflow/
├── node.go           # Node 核心实现
├── flow.go           # Flow 和 BatchFlow 实现
├── pocketflow_test.go # 测试文件
└── README.md         # 文档
```

## 测试

运行测试：

```bash
go test -v
```

测试覆盖：
- Node 基础功能
- Flow 流程编排
- 条件分支（Agent 模式）
- 批量处理
- 节点克隆
- 参数传递

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License