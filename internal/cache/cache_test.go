package cache

// used for testing only, not exported
import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/b07505054/gocache/internal/peer"
)

type mockPeerGetter struct {
	value []byte
	err   error
}

func (m *mockPeerGetter) Get(ctx context.Context, key string) ([]byte, error) {
	return m.value, m.err
}

func (m *mockPeerGetter) OwnerGet(ctx context.Context, key string) ([]byte, error) {
	return m.value, m.err
}

type mockPeerPicker struct {
	getter peer.Getter
	ok     bool
}

func (m *mockPeerPicker) PickPeer(key string) (peer.Getter, bool) {
	return m.getter, m.ok
}

func TestSetGet(t *testing.T) {
	c := New(2)

	c.Set("a", []byte("1"), 0)

	val, ok := c.Get("a")
	if !ok {
		t.Fatalf("expected key a to exist")
	}
	if string(val) != "1" {
		t.Fatalf("expected value 1, got %s", val)
	}
}

func TestTTLExpiration(t *testing.T) {
	c := New(2)

	c.Set("a", []byte("1"), 100*time.Millisecond)

	time.Sleep(150 * time.Millisecond)

	_, ok := c.Get("a")
	if ok {
		t.Fatalf("expected key a to expire")
	}
}

func TestLRUEviction(t *testing.T) {
	c := New(2)

	c.Set("a", []byte("1"), 0)
	c.Set("b", []byte("2"), 0)

	// use a
	c.Get("a")

	// insert c
	c.Set("c", []byte("3"), 0)

	if _, ok := c.Get("b"); ok {
		t.Fatalf("expected b to be evicted")
	}

	if _, ok := c.Get("a"); !ok {
		t.Fatalf("expected a to still exist")
	}

	if _, ok := c.Get("c"); !ok {
		t.Fatalf("expected c to exist")
	}
}

func TestDelete(t *testing.T) {
	c := New(2)

	c.Set("a", []byte("1"), 0)
	c.Delete("a")

	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected a to be deleted")
	}
}
func TestGetOrLoadSingleflight(t *testing.T) {
	c := New(10)

	var loaderCalls int32
	var wg sync.WaitGroup

	n := 20
	results := make([][]byte, n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			val, err := c.GetOrLoad("hotkey", time.Minute, func(key string) ([]byte, error) {
				atomic.AddInt32(&loaderCalls, 1)
				time.Sleep(50 * time.Millisecond)
				return []byte("loaded-value"), nil
			})

			results[idx] = val
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	if loaderCalls != 1 {
		t.Fatalf("expected loader to be called once, got %d", loaderCalls)
	}

	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("unexpected error: %v", errs[i])
		}
		if string(results[i]) != "loaded-value" {
			t.Fatalf("unexpected value: %s", results[i])
		}
	}
}
func BenchmarkCacheGet(b *testing.B) {
	c := New(1000)
	c.Set("hotkey", []byte("value"), time.Minute)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("hotkey")
	}
}

func BenchmarkCacheGetParallel(b *testing.B) {
	c := New(1000)
	c.Set("hotkey", []byte("value"), time.Minute)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.Get("hotkey")
		}
	})
}
func TestGetOrLoadFromPeer(t *testing.T) {
	c := New(10)

	picker := &mockPeerPicker{
		getter: &mockPeerGetter{value: []byte("peer-value")},
		ok:     true,
	}
	c.RegisterPeers(picker)

	var loaderCalls int32

	val, err := c.GetOrLoad("remote-key", time.Minute, func(key string) ([]byte, error) {
		atomic.AddInt32(&loaderCalls, 1)
		return []byte("local-value"), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(val) != "peer-value" {
		t.Fatalf("expected peer-value, got %s", val)
	}

	if loaderCalls != 0 {
		t.Fatalf("expected local loader not to be called, got %d", loaderCalls)
	}
}

func TestGetOrLoadPeerFallbackToLoader(t *testing.T) {
	c := New(10)

	picker := &mockPeerPicker{
		getter: &mockPeerGetter{value: nil, err: context.DeadlineExceeded},
		ok:     true,
	}
	c.RegisterPeers(picker)

	var loaderCalls int32

	val, err := c.GetOrLoad("remote-key", time.Minute, func(key string) ([]byte, error) {
		atomic.AddInt32(&loaderCalls, 1)
		return []byte("local-value"), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(val) != "local-value" {
		t.Fatalf("expected local-value, got %s", val)
	}

	if loaderCalls != 1 {
		t.Fatalf("expected local loader to be called once, got %d", loaderCalls)
	}
}
func TestResolveAsOwnerLoadsAndCaches(t *testing.T) {
	c := New(10)

	var loaderCalls int32
	c.RegisterOwnerLoader(time.Minute, func(key string) ([]byte, error) {
		atomic.AddInt32(&loaderCalls, 1)
		return []byte("owner-loaded-value"), nil
	})

	val, err := c.ResolveAsOwner("owner-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "owner-loaded-value" {
		t.Fatalf("expected owner-loaded-value, got %s", val)
	}

	val, err = c.ResolveAsOwner("owner-key")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if string(val) != "owner-loaded-value" {
		t.Fatalf("expected owner-loaded-value on second call, got %s", val)
	}

	if loaderCalls != 1 {
		t.Fatalf("expected owner loader to be called once, got %d", loaderCalls)
	}
}

func TestResolveAsOwnerWithoutLoader(t *testing.T) {
	c := New(10)

	_, err := c.ResolveAsOwner("owner-key")
	if err == nil {
		t.Fatalf("expected error when owner loader is not registered")
	}
}
