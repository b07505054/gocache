package hash

import "testing"

func TestConsistentHashGet(t *testing.T) {
	m := New(3)
	m.Add("nodeA", "nodeB", "nodeC")

	node := m.Get("user:123")
	if node == "" {
		t.Fatalf("expected a node, got empty string")
	}
}

func TestConsistentHashEmpty(t *testing.T) {
	m := New(3)

	node := m.Get("user:123")
	if node != "" {
		t.Fatalf("expected empty string for empty ring, got %s", node)
	}
}

func TestConsistentHashDeterministic(t *testing.T) {
	m := New(10)
	m.Add("nodeA", "nodeB", "nodeC")

	node1 := m.Get("hotkey")
	node2 := m.Get("hotkey")

	if node1 != node2 {
		t.Fatalf("expected deterministic mapping, got %s and %s", node1, node2)
	}
}
