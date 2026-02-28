package pocketflow

import (
	"reflect"
)

type SharedStore map[string]any

type PrepFunc func(shared SharedStore) any

type ExecFunc func(prepRes any) any

type PostFunc func(shared SharedStore, prepRes, execRes any) string

type NodeRunner interface {
	Run(n *Node, shared SharedStore) string
}

func convertToAnySlice(value any) []any {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice {
		return nil
	}

	result := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}
	return result
}

type Node struct {
	name       string
	prep       PrepFunc
	exec       ExecFunc
	post       PostFunc
	successors map[string]*Node
	params     map[string]any
	runner     NodeRunner
}

type SingleRunner struct{}

func (r *SingleRunner) Run(n *Node, shared SharedStore) string {
	var prepRes any
	if n.prep != nil {
		prepRes = n.prep(shared)
	}

	var execRes any
	if n.exec != nil {
		execRes = n.exec(prepRes)
	}

	var action string
	if n.post != nil {
		action = n.post(shared, prepRes, execRes)
	} else {
		action = "default"
	}

	return action
}

type BatchRunner struct{}

func (r *BatchRunner) Run(n *Node, shared SharedStore) string {
	var prepRes any
	if n.prep != nil {
		prepRes = n.prep(shared)
	}

	var execRes any
	if n.exec != nil {
		if items := convertToAnySlice(prepRes); items != nil {
			results := make([]any, len(items))
			for i, item := range items {
				results[i] = n.exec(item)
			}
			execRes = results
		} else {
			execRes = []any{n.exec(prepRes)}
		}
	}

	var action string
	if n.post != nil {
		action = n.post(shared, prepRes, execRes)
	} else {
		action = "default"
	}

	return action
}

func NewNode(name string) *Node {
	return &Node{
		name:       name,
		successors: make(map[string]*Node),
		params:     make(map[string]any),
		runner:     &SingleRunner{},
	}
}

func (n *Node) Prep(fn PrepFunc) *Node {
	n.prep = fn
	return n
}

func (n *Node) Exec(fn ExecFunc) *Node {
	n.exec = fn
	return n
}

func (n *Node) Post(fn PostFunc) *Node {
	n.post = fn
	return n
}

func (n *Node) Then(next *Node) *Node {
	return n.ThenWith("default", next)
}

func (n *Node) ThenWith(action string, next *Node) *Node {
	n.successors[action] = next
	return n
}

func (n *Node) SetParams(params map[string]any) *Node {
	for k, v := range params {
		n.params[k] = v
	}
	return n
}

func (n *Node) GetParam(key string) (any, bool) {
	v, ok := n.params[key]
	return v, ok
}

func (n *Node) GetNext(action string) *Node {
	if next, ok := n.successors[action]; ok {
		return next
	}
	return n.successors["default"]
}

func (n *Node) Run(shared SharedStore) string {
	return n.runner.Run(n, shared)
}

func (n *Node) Clone() *Node {
	clone := &Node{
		name:       n.name,
		prep:       n.prep,
		exec:       n.exec,
		post:       n.post,
		successors: make(map[string]*Node),
		params:     make(map[string]any),
		runner:     n.runner,
	}

	for k, v := range n.params {
		clone.params[k] = v
	}

	for action, node := range n.successors {
		clone.successors[action] = node
	}

	return clone
}

func (n *Node) String() string {
	return n.name
}

type BatchNode struct {
	*Node
}

func NewBatchNode(name string) *BatchNode {
	return &BatchNode{
		Node: &Node{
			name:       name,
			successors: make(map[string]*Node),
			params:     make(map[string]any),
			runner:     &BatchRunner{},
		},
	}
}

func (n *BatchNode) Clone() *BatchNode {
	return &BatchNode{
		Node: n.Node.Clone(),
	}
}
