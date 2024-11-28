package sflight

import (
	"sync"
	"time"
)

type call[T any] struct {
	wg     sync.WaitGroup
	val    T
	err    error
	execAt time.Time
}

type Group[K comparable, T any] struct {
	mu      sync.Mutex
	m       map[K]*call[T]
	expires time.Duration
	scanAt  time.Time
}

func New[K comparable, T any](expires time.Duration) *Group[K, T] {
	return &Group[K, T]{
		expires: expires,
	}
}
func (g *Group[K, T]) Do(k K, fn func() (T, error)) (T, error, bool) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[K]*call[T])
	}
	now := time.Now()
	if c, ok := g.m[k]; ok && now.Sub(c.execAt) < g.expires {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err, false
	}
	c := &call[T]{
		wg:     sync.WaitGroup{},
		execAt: now,
	}
	c.wg.Add(1)
	g.m[k] = c
	if now.Sub(g.scanAt) >= g.expires*2 {
		g.scanAt = now
		go g.clearExpired(now)
	}
	g.mu.Unlock()
	g.doCall(c, fn)

	return c.val, c.err, true
}

func (g *Group[K, T]) Forget(k K) {
	g.mu.Lock()
	delete(g.m, k)
	g.mu.Unlock()
}

func (g *Group[K, T]) doCall(c *call[T], fn func() (T, error)) {
	defer c.wg.Done()
	c.val, c.err = fn()
}

func (g *Group[K, T]) clearExpired(t time.Time) {
	g.mu.Lock()
	for k, c := range g.m {
		if t.Sub(c.execAt) >= g.expires {
			delete(g.m, k)
		}
	}
	g.mu.Unlock()
}
