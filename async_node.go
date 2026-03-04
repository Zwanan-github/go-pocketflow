package pocketflow

type AsyncResult struct {
	Action string
}

type AsyncNode struct {
	*Node
}

func NewAsyncNode(name string) *AsyncNode {
	return &AsyncNode{
		Node: NewNode(name),
	}
}

func (an *AsyncNode) Prep(prepFunc func(shared SharedStore) any) *AsyncNode {
	an.Node.Prep(prepFunc)
	return an
}

func (an *AsyncNode) Exec(execFunc func(prepRes any) any) *AsyncNode {
	an.Node.Exec(execFunc)
	return an
}

func (an *AsyncNode) Post(postFunc func(shared SharedStore, prepRes, execRes any) string) *AsyncNode {
	an.Node.Post(postFunc)
	return an
}

func (an *AsyncNode) Run(shared SharedStore) string {
	var prepRes any
	if an.prep != nil {
		prepRes = an.prep(shared)
	}

	var execRes any
	if an.exec != nil {
		execRes = an.exec(prepRes)
	}

	var action string
	if an.post != nil {
		action = an.post(shared, prepRes, execRes)
	} else {
		action = "default"
	}

	return action
}

func (an *AsyncNode) RunAsync(shared SharedStore) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- AsyncResult{Action: "panic-recovered"}
			}
		}()

		action := an.Run(shared)
		resultChan <- AsyncResult{Action: action}
	}()

	return resultChan
}

func (an *AsyncNode) Clone() *AsyncNode {
	return &AsyncNode{
		Node: an.Node.Clone(),
	}
}

type AsyncBatchNode struct {
	syncBatchNode *BatchNode
}

func NewAsyncBatchNode(syncBatchNode *BatchNode) *AsyncBatchNode {
	return &AsyncBatchNode{
		syncBatchNode: syncBatchNode,
	}
}

func (abn *AsyncBatchNode) RunAsync(shared SharedStore) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- AsyncResult{Action: "panic-recovered"}
			}
		}()

		var prepRes any
		if abn.syncBatchNode.prep != nil {
			prepRes = abn.syncBatchNode.prep(shared)
		}

		var execRes any
		if abn.syncBatchNode.exec != nil {
			if items := convertToAnySlice(prepRes); items != nil {
				results := make([]any, len(items))

				type result struct {
					index int
					value any
				}

				ch := make(chan result, len(items))

				for i, item := range items {
					go func(idx int, itm any) {
						defer func() {
							if r := recover(); r != nil {
								ch <- result{index: idx, value: "exec-panic-recovered"}
							}
						}()

						value := abn.syncBatchNode.exec(itm)
						ch <- result{index: idx, value: value}
					}(i, item)
				}

				for i := 0; i < len(items); i++ {
					res := <-ch
					results[res.index] = res.value
				}

				execRes = results
			} else {
				execRes = []any{abn.syncBatchNode.exec(prepRes)}
			}
		}

		var action string
		if abn.syncBatchNode.post != nil {
			action = abn.syncBatchNode.post(shared, prepRes, execRes)
		} else {
			action = "default"
		}

		resultChan <- AsyncResult{Action: action}
	}()

	return resultChan
}

func (abn *AsyncBatchNode) Clone() *AsyncBatchNode {
	return &AsyncBatchNode{
		syncBatchNode: abn.syncBatchNode.Clone(),
	}
}
