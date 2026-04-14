package cache

import "sync"

// call represents an in-flight or completed request
type call struct {
	wg  sync.WaitGroup
	val []byte
	err error
}

// Group manages duplicate suppression for concurrent calls
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// NewGroup creates a new Group
func NewGroup() *Group {
	return &Group{
		m: make(map[string]*call),
	}
}

// Do ensures that only one execution is in-flight per key
func (g *Group) Do(key string, fn func() ([]byte, error)) ([]byte, error) {
	g.mu.Lock()

	// If request is already in-flight, wait for it
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	// Otherwise create a new call
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// Execute the function
	c.val, c.err = fn()
	c.wg.Done()

	// Clean up
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
