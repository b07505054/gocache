package hash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Map represents a consistent hash ring.
type Map struct {
	replicas int
	keys     []int
	hashMap  map[int]string
}

// New creates a new consistent hash ring.
func New(replicas int) *Map {
	return &Map{
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
}

// hashKey computes a hash for a given string key.
func hashKey(key string) int {
	return int(crc32.ChecksumIEEE([]byte(key)))
}

// Add inserts one or more real nodes into the hash ring,
// along with their virtual nodes.
func (m *Map) Add(nodes ...string) {
	for _, node := range nodes {
		for i := 0; i < m.replicas; i++ {
			virtualKey := strconv.Itoa(i) + node
			h := hashKey(virtualKey)
			m.keys = append(m.keys, h)
			m.hashMap[h] = node
		}
	}
	sort.Ints(m.keys)
}

// Get returns the node responsible for the given key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	h := hashKey(key)

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= h
	})

	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}
