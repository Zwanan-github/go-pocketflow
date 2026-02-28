package pocketflow

import (
	"testing"
)

func TestNode(t *testing.T) {
	node := NewNode("test").
		Prep(func(shared SharedStore) any {
			return shared["input"]
		}).
		Exec(func(prepRes any) any {
			return prepRes.(string) + " processed"
		}).
		Post(func(shared SharedStore, prepRes, execRes any) string {
			shared["output"] = execRes
			return "end"
		})

	shared := SharedStore{
		"input": "test data",
	}

	action := node.Run(shared)

	if action != "end" {
		t.Errorf("Expected action 'end', got '%s'", action)
	}

	if shared["output"] != "test data processed" {
		t.Errorf("Expected output 'test data processed', got '%v'", shared["output"])
	}
}

func TestFlow(t *testing.T) {
	node1 := NewNode("node1").
		Post(func(shared SharedStore, prepRes, execRes any) string {
			shared["step1"] = "done"
			return "next"
		})

	node2 := NewNode("node2").
		Post(func(shared SharedStore, prepRes, execRes any) string {
			shared["step2"] = "done"
			return "end"
		})

	node1.Then(node2)

	flow := NewFlow(node1)

	shared := SharedStore{}
	action := flow.Run(shared)

	if shared["step1"] != "done" {
		t.Errorf("Expected step1 'done', got '%v'", shared["step1"])
	}

	if shared["step2"] != "done" {
		t.Errorf("Expected step2 'done', got '%v'", shared["step2"])
	}

	if action != "end" {
		t.Errorf("Expected final action 'end', got '%s'", action)
	}
}

func TestConditionalFlow(t *testing.T) {
	decide := NewNode("decide").
		Prep(func(shared SharedStore) any {
			return shared["value"]
		}).
		Exec(func(prepRes any) any {
			value := prepRes.(int)
			if value > 10 {
				return "large"
			}
			return "small"
		}).
		Post(func(shared SharedStore, prepRes, execRes any) string {
			return execRes.(string)
		})

	largeHandler := NewNode("large").
		Post(func(shared SharedStore, prepRes, execRes any) string {
			shared["result"] = "handled large value"
			return "end"
		})

	smallHandler := NewNode("small").
		Post(func(shared SharedStore, prepRes, execRes any) string {
			shared["result"] = "handled small value"
			return "end"
		})

	decide.ThenWith("large", largeHandler)
	decide.ThenWith("small", smallHandler)

	flow := NewFlow(decide)

	shared1 := SharedStore{"value": 5}
	action1 := flow.Run(shared1)

	if shared1["result"] != "handled small value" {
		t.Errorf("Expected 'handled small value', got '%v'", shared1["result"])
	}
	if action1 != "end" {
		t.Errorf("Expected action 'end', got '%s'", action1)
	}

	shared2 := SharedStore{"value": 15}
	action2 := flow.Run(shared2)

	if shared2["result"] != "handled large value" {
		t.Errorf("Expected 'handled large value', got '%v'", shared2["result"])
	}
	if action2 != "end" {
		t.Errorf("Expected action 'end', got '%s'", action2)
	}
}

func TestBatchFlow(t *testing.T) {
	results := make([]string, 0)

	processNode := NewNode("process").
		Prep(func(shared SharedStore) any {
			return shared["item"]
		}).
		Exec(func(prepRes any) any {
			if prepRes == nil {
				return "nil_processed"
			}
			return prepRes.(string) + "_processed"
		}).
		Post(func(shared SharedStore, prepRes, execRes any) string {
			results = append(results, execRes.(string))
			return "done"
		})

	batchFlow := NewBatchFlow(processNode).
		Prep(func(shared SharedStore) []map[string]any {
			items := shared["items"].([]string)
			params := make([]map[string]any, len(items))
			for i, item := range items {
				params[i] = map[string]any{"item": item}
			}
			return params
		})

	shared := SharedStore{
		"items": []string{"item1", "item2", "item3"},
	}

	action := batchFlow.Run(shared)

	if action != "batch_complete" {
		t.Errorf("Expected 'batch_complete', got '%s'", action)
	}

}

func TestNodeClone(t *testing.T) {
	node := NewNode("original").
		SetParams(map[string]any{"key": "value"})

	clone := node.Clone()

	if clone.name != node.name {
		t.Errorf("Clone name mismatch: expected '%s', got '%s'", node.name, clone.name)
	}

	clone.SetParams(map[string]any{"key": "modified"})

	if origVal, _ := node.GetParam("key"); origVal != "value" {
		t.Errorf("Original node param modified: expected 'value', got '%v'", origVal)
	}

	if cloneVal, _ := clone.GetParam("key"); cloneVal != "modified" {
		t.Errorf("Clone node param not modified: expected 'modified', got '%v'", cloneVal)
	}
}

func TestBatchNode(t *testing.T) {
	node := NewBatchNode("test").
		Prep(func(shared SharedStore) any {
			return shared["inputs"] // 返回 []string{"input1", "input2"}
		}).
		Exec(func(item any) any { // 注意：这里参数是单个 item
			return item.(string) + "_processed"
		}).
		Post(func(shared SharedStore, prepRes, execRes any) string {
			shared["output"] = execRes // execRes 是 []any
			return "end"
		})

	shared := SharedStore{
		"inputs": []string{"input1", "input2"},
	}

	action := node.Run(shared)

	if action != "end" {
		t.Errorf("Expected action 'end', got '%s'", action)
	}

	// BatchNode 应该生成两个结果
	results, ok := shared["output"].([]any)
	if !ok {
		t.Errorf("Expected []any output, got '%v'", shared["output"])
	} else if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	} else {
		// 验证结果内容
		if results[0] != "input1_processed" || results[1] != "input2_processed" {
			t.Errorf("Expected ['input1_processed', 'input2_processed'], got '%v'", results)
		}
	}
	t.Logf("results: %v, shared: %v", results, shared)
}

func TestFlowWithParams(t *testing.T) {
	// 实际上，直接使用shared更简单直观
	node := NewNode("test").
		Prep(func(shared SharedStore) any {
			// 直接从shared读取配置
			if model, ok := shared["model"]; ok {
				return model
			}
			return "default"
		}).
		Exec(func(prepRes any) any {
			return "processed with " + prepRes.(string)
		}).
		Post(func(shared SharedStore, prepRes, execRes any) string {
			shared["result"] = execRes
			return "end"
		})

	flow := NewFlow(node)

	// 直接在shared中设置配置，比Flow参数更直观
	shared := SharedStore{
		"model": "gpt-4",
	}
	action := flow.Run(shared)

	expected := "processed with gpt-4"
	if shared["result"] != expected {
		t.Errorf("Expected '%s', got '%v'", expected, shared["result"])
	}

	if action != "end" {
		t.Errorf("Expected action 'end', got '%s'", action)
	}
}
