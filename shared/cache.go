package shared

import (
	"sync"
	"time"
)

type Cache struct {
	Mu         sync.Mutex
	CacheItems map[string]cacheEntry
}

type cacheEntry struct {
	createdAt time.Time
	Val       []byte
}

func (c *Cache) Add(key string, value []byte) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	c.CacheItems[key] = cacheEntry{
		createdAt: time.Now(),
		Val:       value,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	item, ok := c.CacheItems[key]
	if !ok {
		return []byte{}, false
	}

	return item.Val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for t := range ticker.C {
		for key, entry := range c.CacheItems {
			delTime := entry.createdAt.Add(interval)
			if delTime.Compare(t) == -1 {
				c.Mu.Lock()
				delete(c.CacheItems, key)
				c.Mu.Unlock()
			}
		}
	}
}

func NewCache(interval time.Duration) *Cache {
	c := Cache{
		CacheItems: make(map[string]cacheEntry),
		Mu:         sync.Mutex{},
	}
	go c.reapLoop(interval)
	return &c
}
