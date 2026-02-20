package infra

import (
	"sync"
	"testing"
	"time"
)

func TestCacheSetWithTTLExpiresOnGetAll(t *testing.T) {
	c := NewCache()
	c.Set(1, "value", 10*time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	vals, err := c.GetAll()
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(vals) != 0 {
		t.Fatalf("expected no values after expiration, got %d", len(vals))
	}
}

func TestCacheWithoutTTLDoesNotExpire(t *testing.T) {
	c := NewCache()
	c.Set(1, "value", 0)

	time.Sleep(20 * time.Millisecond)

	got, err := c.Get(1)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "value" {
		t.Fatalf("unexpected value: got %v want %v", got, "value")
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	c := NewCache()

	const workers = 12
	const opsPerWorker = 300

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(worker int) {
			defer wg.Done()

			base := int64(worker * 10_000)
			for i := 0; i < opsPerWorker; i++ {
				id := base + int64(i)
				c.Set(id, i, 0)

				got, err := c.Get(id)
				if err != nil {
					t.Errorf("get failed: %v", err)
					return
				}
				if got == nil {
					t.Errorf("expected non-nil value for id=%d", id)
					return
				}

				if _, err = c.GetAll(); err != nil {
					t.Errorf("get all failed: %v", err)
					return
				}
			}
		}(w)
	}

	wg.Wait()
}

func TestCacheConcurrentAccessWithTTL(t *testing.T) {
	c := NewCache()

	const writers = 8
	const readers = 8
	const opsPerWriter = 200
	ttl := 5 * time.Millisecond

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for w := 0; w < writers; w++ {
		go func(worker int) {
			defer wg.Done()

			base := int64(worker * 10_000)
			for i := 0; i < opsPerWriter; i++ {
				id := base + int64(i)
				c.Set(id, i, ttl)
				time.Sleep(time.Microsecond * 200)
			}
		}(w)
	}

	for r := 0; r < readers; r++ {
		go func(reader int) {
			defer wg.Done()

			base := int64((reader % writers) * 10_000)
			for i := 0; i < opsPerWriter; i++ {
				id := base + int64(i)
				_, err := c.Get(id)
				if err != nil {
					t.Errorf("get failed: %v", err)
					return
				}
				if _, err = c.GetAll(); err != nil {
					t.Errorf("get all failed: %v", err)
					return
				}
				time.Sleep(time.Microsecond * 150)
			}
		}(r)
	}

	wg.Wait()

	// give remaining keys time to expire, then verify lazy cleanup on read path
	time.Sleep(ttl * 3)
	vals, err := c.GetAll()
	if err != nil {
		t.Fatalf("get all after ttl: %v", err)
	}
	if len(vals) != 0 {
		t.Fatalf("expected all values to expire, got %d", len(vals))
	}
}
