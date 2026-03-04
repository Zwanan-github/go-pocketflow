package pocketflow

import "sync"

type AsyncFlow struct {
	syncFlow *Flow
}

func NewAsyncFlow(asyncNode *AsyncNode) *AsyncFlow {
	return &AsyncFlow{
		syncFlow: NewFlow(asyncNode.Node),
	}
}

func (af *AsyncFlow) RunAsync(shared SharedStore) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- AsyncResult{Action: "panic-recovered"}
			}
		}()

		action := af.syncFlow.Run(shared)
		resultChan <- AsyncResult{Action: action}
	}()
	return resultChan
}

type AsyncBatchFlow struct {
	syncBatchFlow *BatchFlow
}

func NewAsyncBatchFlow(syncBatchFlow *BatchFlow) *AsyncBatchFlow {
	return &AsyncBatchFlow{
		syncBatchFlow: syncBatchFlow,
	}
}

func (abf *AsyncBatchFlow) RunAsync(shared SharedStore) <-chan AsyncResult {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- AsyncResult{Action: "panic-recovered"}
			}
		}()

		if abf.syncBatchFlow.prep == nil {
			action := abf.syncBatchFlow.Flow.Run(shared)
			resultChan <- AsyncResult{Action: action}
			return
		}

		batchParams := abf.syncBatchFlow.prep(shared)

		var wg sync.WaitGroup
		wg.Add(len(batchParams))

		for _, params := range batchParams {
			go func(p map[string]any) {
				defer func() {
					if r := recover(); r != nil {
						// 单个 batch panic 不影响其他
					}
					wg.Done()
				}()

				batchShared := make(SharedStore)
				for k, v := range shared {
					batchShared[k] = v
				}
				for k, v := range p {
					batchShared[k] = v
				}
				tempFlow := NewFlow(abf.syncBatchFlow.startNode)
				tempFlow.Run(batchShared)
			}(params)
		}

		wg.Wait()
		resultChan <- AsyncResult{Action: "batch_complete"}
	}()

	return resultChan
}
