package peer

import (
	"strconv"
	"testing"
)

func TestPickPeerEmpty(t *testing.T) {
	p := NewMapPicker("self", 10)

	getter, ok := p.PickPeer("user:1")
	if ok || getter != nil {
		t.Fatalf("expected no peer for empty picker")
	}
}

func TestPickPeerDeterministic(t *testing.T) {
	p := NewMapPicker("self", 10)

	peers := map[string]Getter{
		"nodeA": &mockGetter{value: []byte("A")},
		"nodeB": &mockGetter{value: []byte("B")},
		"nodeC": &mockGetter{value: []byte("C")},
	}
	p.Set(peers)

	g1, ok1 := p.PickPeer("hotkey")
	g2, ok2 := p.PickPeer("hotkey")

	if ok1 != ok2 {
		t.Fatalf("expected deterministic peer selection")
	}
	if ok1 && g1 != g2 {
		t.Fatalf("expected same getter instance for same key")
	}
}
func TestPickPeerSkipsSelf(t *testing.T) {
	p := NewMapPicker("nodeA", 100)

	peers := map[string]Getter{
		"nodeA": &mockGetter{value: []byte("A")},
		"nodeB": &mockGetter{value: []byte("B")},
	}
	p.Set(peers)

	for i := 0; i < 100; i++ {
		getter, ok := p.PickPeer("key-" + strconv.Itoa(i))
		if ok && getter == peers["nodeA"] {
			t.Fatalf("expected self node to be skipped")
		}
	}
}
func TestSetHTTPPeers(t *testing.T) {
	p := NewMapPicker("self", 10)

	p.SetHTTPPeers(map[string]string{
		"nodeA": "http://localhost:8081",
		"nodeB": "http://localhost:8082",
	})

	getter, ok := p.PickPeer("some-key")
	if ok && getter == nil {
		t.Fatalf("expected getter to be non-nil when peer is selected")
	}
}
