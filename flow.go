package pocketflow

type Flow struct {
	startNode *Node
}

func NewFlow(start *Node) *Flow {
	return &Flow{
		startNode: start,
	}
}

func (f *Flow) Run(shared SharedStore) string {
	curr := f.startNode.Clone()
	var lastAction string
	for curr != nil {
		lastAction = curr.Run(shared)

		next := curr.GetNext(lastAction)
		if next != nil {
			curr = next.Clone()
		} else {
			curr = nil
		}
	}
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
	batchParams := bf.prep(shared)
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
	return "batch_complete"
}
