package captcha

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type memoryCache struct {
	mu     sync.RWMutex
	values map[string]any
}

func newMemoryCache() *memoryCache {
	return &memoryCache{values: make(map[string]any)}
}

func (c *memoryCache) Set(key string, value any, _ time.Duration) {
	c.mu.Lock()
	c.values[key] = value
	c.mu.Unlock()
}

func (c *memoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	value, ok := c.values[key]
	c.mu.RUnlock()
	return value, ok
}

func (c *memoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.values, key)
	c.mu.Unlock()
}

func (c *memoryCache) Increment(string, int64) error        { return nil }
func (c *memoryCache) Add(string, any, time.Duration) error { return nil }

func TestStoreVerifyConsumesOnlySuccessfulAnswer(t *testing.T) {
	store := NewStore(newMemoryCache())
	store.Set("id", "1234")

	if store.Verify("id", "wrong", true) {
		t.Fatal("wrong answer passed verification")
	}
	if !store.Verify("id", "1234", false) {
		t.Fatal("correct answer failed after a wrong attempt")
	}
	if store.Verify("id", "1234", false) {
		t.Fatal("successful answer was not consumed")
	}
}

func TestStoreUnexpectedValueTypeDoesNotPanic(t *testing.T) {
	backend := newMemoryCache()
	backend.Set(cachePrefix+"id", 1234, defaultTTL)
	store := NewStore(backend)

	if value := store.Get("id", false); value != "" {
		t.Fatalf("unexpected string value: %q", value)
	}
	if store.Verify("id", "1234", false) {
		t.Fatal("non-string cache value passed verification")
	}
}

func TestStoreConcurrentReplayAllowsOneSuccess(t *testing.T) {
	store := NewStore(newMemoryCache())
	store.Set("id", "1234")

	const workers = 64
	start := make(chan struct{})
	var successes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(workers)
	for range workers {
		go func() {
			defer wait.Done()
			<-start
			if store.Verify("id", "1234", false) {
				successes.Add(1)
			}
		}()
	}
	close(start)
	wait.Wait()

	if got := successes.Load(); got != 1 {
		t.Fatalf("successful replays = %d, want 1", got)
	}
}

func TestHMACSHA256(t *testing.T) {
	const want = "5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0"
	if got := hmacSHA256("key", "data"); got != want {
		t.Fatalf("hmacSHA256() = %q, want %q", got, want)
	}
}
