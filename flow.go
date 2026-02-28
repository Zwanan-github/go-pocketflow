package pocketflow

import (
	"log"
)

type Flow struct {
	startNode *Node
}

func NewFlow(start *Node) *Flow {
	return &Flow{
		startNode: start,
	}
}

func (f *Flow) Run(shared SharedStore) string {
	log.Printf("[Flow] Starting flow, begin node: %s", f.startNode.name)

	curr := f.startNode.Clone()
	var lastAction string

	for curr != nil {
		log.Printf("[Flow] Running node: %s", curr.name)
		lastAction = curr.Run(shared)
		log.Printf("[Flow] Node %s returned action: %s", curr.name, lastAction)

		next := curr.GetNext(lastAction)
		if next != nil {
			log.Printf("[Flow] Next node: %s", next.name)
			curr = next.Clone()
		} else {
			log.Printf("[Flow] Flow ended, no successor for action '%s'", lastAction)
			curr = nil
		}
	}

	log.Printf("[Flow] Flow completed, final action: %s", lastAction)
	return lastAction
}

type BatchFlow struct {
	*Flow
	prep BatchPrepFunc
}

func NewBatchFlow(start *Node) *BatchFlow {
	return &BatchFlow{
		Flow: NewFlow(start),
	}
}

type BatchPrepFunc func(shared SharedStore) []map[string]any

func (bf *BatchFlow) Prep(fn BatchPrepFunc) *BatchFlow {
	bf.prep = fn
	return bf
}
func (bf *BatchFlow) Run(shared SharedStore) string {
	if bf.prep == nil {
		return bf.Flow.Run(shared)
	}

	log.Printf("[BatchFlow] Starting batch processing")
	batchParams := bf.prep(shared)
	log.Printf("[BatchFlow] Generated %d batch params", len(batchParams))

	for i, params := range batchParams {
		batchShared := make(SharedStore)
		for k, v := range shared {
			batchShared[k] = v
		}
		for k, v := range params {
			batchShared[k] = v
		}

		tempFlow := NewFlow(bf.startNode)
		tempFlow.Run(batchShared)

		_ = i
	}

	log.Printf("[BatchFlow] Batch processing completed")
	return "batch_complete"
}
