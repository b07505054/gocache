package cache

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/b07505054/gocache/internal/peer"
)

var ErrNoOwnerLoader = errors.New("owner loader is not registered")

type entry struct {
	key       string
	value     []byte
	expiresAt time.Time
}

type Cache struct {
	mu          sync.RWMutex
	ll          *list.List
	items       map[string]*list.Element
	maxEntries  int
	sf          *Group
	picker      peer.Picker
	ownerLoader func(string) ([]byte, error)
	ownerTTL    time.Duration
}

func New(maxEntries int) *Cache {
	return &Cache{
		ll:         list.New(),
		items:      make(map[string]*list.Element),
		maxEntries: maxEntries,
		sf:         NewGroup(),
	}
}

func (c *Cache) RegisterOwnerLoader(ttl time.Duration, loader func(string) ([]byte, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ownerTTL = ttl
	c.ownerLoader = loader
}

func (c *Cache) RegisterPeers(p peer.Picker) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.picker = p
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		ent := ele.Value.(*entry)
		ent.value = value
		if ttl > 0 {
			ent.expiresAt = time.Now().Add(ttl)
		} else {
			ent.expiresAt = time.Time{}
		}
		c.ll.MoveToFront(ele)
		return
	}

	ent := &entry{
		key:   key,
		value: value,
	}
	if ttl > 0 {
		ent.expiresAt = time.Now().Add(ttl)
	}

	ele := c.ll.PushFront(ent)
	c.items[key] = ele

	if c.maxEntries > 0 && c.ll.Len() > c.maxEntries {
		c.removeOldest()
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ele, ok := c.items[key]
	if !ok {
		return nil, false
	}

	ent := ele.Value.(*entry)

	if !ent.expiresAt.IsZero() && time.Now().After(ent.expiresAt) {
		c.removeElement(ele)
		return nil, false
	}

	c.ll.MoveToFront(ele)
	return ent.value, true
}

func (c *Cache) GetOrLoad(key string, ttl time.Duration, loader func(string) ([]byte, error)) ([]byte, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}

	return c.sf.Do(key, func() ([]byte, error) {
		if val, ok := c.Get(key); ok {
			return val, nil
		}

		c.mu.RLock()
		p := c.picker
		c.mu.RUnlock()

		if p != nil {
			if getter, ok := p.PickPeer(key); ok {
				if ownerGetter, ok := getter.(peer.OwnerGetter); ok {
					val, err := ownerGetter.OwnerGet(context.Background(), key)
					if err == nil {
						c.Set(key, val, ttl)
						return val, nil
					}
				}
			}
		}

		val, err := loader(key)
		if err != nil {
			return nil, err
		}

		c.Set(key, val, ttl)
		return val, nil
	})
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		c.removeElement(ele)
	}
}

func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ll.Len()
}

func (c *Cache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(ele *list.Element) {
	c.ll.Remove(ele)
	ent := ele.Value.(*entry)
	delete(c.items, ent.key)
}
func (c *Cache) ResolveAsOwner(key string) ([]byte, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}

	c.mu.RLock()
	loader := c.ownerLoader
	ttl := c.ownerTTL
	c.mu.RUnlock()

	if loader == nil {
		return nil, ErrNoOwnerLoader
	}

	return c.sf.Do("owner:"+key, func() ([]byte, error) {
		if val, ok := c.Get(key); ok {
			return val, nil
		}

		val, err := loader(key)
		if err != nil {
			return nil, err
		}

		c.Set(key, val, ttl)
		return val, nil
	})
}
