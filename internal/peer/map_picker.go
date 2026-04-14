package peer

import (
	"sync"

	"github.com/b07505054/gocache/internal/hash"
)

// MapPicker selects peers using consistent hashing.
type MapPicker struct {
	mu       sync.RWMutex
	self     string
	replicas int
	hashRing *hash.Map
	getters  map[string]Getter
}

// NewMapPicker creates a new MapPicker.
func NewMapPicker(self string, replicas int) *MapPicker {
	return &MapPicker{
		self:     self,
		replicas: replicas,
		hashRing: hash.New(replicas),
		getters:  make(map[string]Getter),
	}
}

// Set registers peers with prebuilt Getter instances.
func (p *MapPicker) Set(peers map[string]Getter) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.hashRing = hash.New(p.replicas)
	p.getters = make(map[string]Getter)

	nodes := make([]string, 0, len(peers))
	for node, getter := range peers {
		p.getters[node] = getter
		nodes = append(nodes, node)
	}
	p.hashRing.Add(nodes...)
}

// SetHTTPPeers registers peers using node name -> base URL mapping.
func (p *MapPicker) SetHTTPPeers(peers map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.hashRing = hash.New(p.replicas)
	p.getters = make(map[string]Getter)

	nodes := make([]string, 0, len(peers))
	for node, baseURL := range peers {
		p.getters[node] = NewHTTPGetter(baseURL)
		nodes = append(nodes, node)
	}
	p.hashRing.Add(nodes...)
}

// PickPeer selects the remote peer responsible for the key.
func (p *MapPicker) PickPeer(key string) (Getter, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.getters) == 0 {
		return nil, false
	}

	node := p.hashRing.Get(key)
	if node == "" || node == p.self {
		return nil, false
	}

	getter, ok := p.getters[node]
	return getter, ok
}
