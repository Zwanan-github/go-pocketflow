## Why

go-pocketflow 当前只支持同步执行模式，限制了处理并发任务和非阻塞 I/O 操作的能力。在现代 LLM 应用中，异步执行对于提升性能和用户体验至关重要，特别是在需要处理多个并发请求或长时间运行的 LLM 调用时。

## What Changes

- **新增 AsyncNode 结构体**：支持异步执行的节点类型
- **新增 AsyncFlow 结构体**：支持异步流程编排
- **扩展 SharedStore**：添加异步状态管理和信号量机制
- **新增并发控制**：支持限制并发数量和超时控制
- **新增错误处理**：异步执行的错误传播和重试机制
- **保持向后兼容**：现有的同步 Node 和 Flow API 保持不变

## Capabilities

### New Capabilities
- `async-execution`: 异步节点执行核心功能，包括并发控制、错误处理和状态管理
- `async-flow`: 异步流程编排，支持节点间的异步协调和依赖管理
- `concurrent-batch`: 并发批量处理，提升 BatchFlow 的性能

### Modified Capabilities
- (无现有能力需要修改，这是纯新增功能)

## Impact

**受影响的代码文件：**
- `node.go` - 需要添加 AsyncNode 相关实现
- `flow.go` - 需要添加 AsyncFlow 和并发控制
- `pocketflow_test.go` - 需要添加异步测试用例

**API 变更：**
- 新增公共 API：`NewAsyncNode()`, `NewAsyncFlow()`, `WithConcurrency()` 等
- 现有 API 保持不变，确保向后兼容性

**依赖关系：**
- 可能需要引入 `context` 包用于超时控制
- 可能需要引入 `sync` 包用于并发同步

**系统影响：**
- 提升处理大规模并发任务的能力
- 增强框架在生产环境的适用性
- 为后续的流式处理和实时响应奠定基础