// Package pocketflow provides a minimal framework for building LLM-powered applications.
// It offers a node-based architecture with support for retries, batching, async execution, and flow orchestration.
package pocketflow

import (
	"log"
)

type SharedStore map[string]any

type PrepFunc func(shared SharedStore) any

type ExecFunc func(prepRes any) any

type PostFunc func(shared SharedStore, prepRes, execRes any) string

type Node struct {
	name       string
	prep       PrepFunc
	exec       ExecFunc
	post       PostFunc
	successors map[string]*Node
	params     map[string]any
}

func NewNode(name string) *Node {
	return &Node{
		name:       name,
		successors: make(map[string]*Node),
		params:     make(map[string]any),
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
	log.Printf("[Node] Running node: %s", n.name)

	var prepRes any
	if n.prep != nil {
		log.Printf("[Node] %s - Prep phase", n.name)
		prepRes = n.prep(shared)
	}

	var execRes any
	if n.exec != nil {
		log.Printf("[Node] %s - Exec phase", n.name)
		execRes = n.exec(prepRes)
	}

	var action string
	if n.post != nil {
		log.Printf("[Node] %s - Post phase", n.name)
		action = n.post(shared, prepRes, execRes)
	} else {
		action = "default"
	}

	log.Printf("[Node] %s - Returning action: %s", n.name, action)
	return action
}

func (n *Node) Clone() *Node {
	clone := &Node{
		name:       n.name,
		prep:       n.prep,
		exec:       n.exec,
		post:       n.post,
		successors: make(map[string]*Node),
		params:     make(map[string]any),
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
		Node: NewNode(name),
	}
}

func (n *BatchNode) Run(shared SharedStore) string {
	log.Printf("[BatchNode] Running batch node: %s", n.name)

	var prepRes any
	if n.prep != nil {
		log.Printf("[BatchNode] %s - Prep phase", n.name)
		prepRes = n.prep(shared)
	}

	var execRes []any
	if n.exec != nil {
		log.Printf("[BatchNode] %s - Exec phase", n.name)
		for _, item := range prepRes.([]any) {
			execRes = append(execRes, n.exec(item))
		}
	}

	var action string
	if n.post != nil {
		log.Printf("[BatchNode] %s - Post phase", n.name)
		action = n.post(shared, prepRes, execRes)
	} else {
		action = "default"
	}

	log.Printf("[BatchNode] %s - Returning action: %s", n.name, action)
	return action
}

func (n *BatchNode) Clone() *BatchNode {
	return &BatchNode{
		Node: n.Node.Clone(),
	}
}
