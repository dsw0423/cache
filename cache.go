package cache

import (
	"sync"

	"github.com/dsw0423/cache/lru"
)

type cache struct {
	mu       sync.Mutex
	ca       *lru.Cache
	maxBytes uint64
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ca == nil {
		return
	}

	if v, ok := c.ca.Get(key); ok {
		return v.(ByteView), true
	}

	return
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ca == nil {
		c.ca = lru.New(c.maxBytes, nil)
	}

	c.ca.Add(key, value)
}
