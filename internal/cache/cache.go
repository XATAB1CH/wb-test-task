package cache

import (
	"container/list"
	"sync"
	"time"
)

type LRUCache struct {
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	list     *list.List
	mu       sync.Mutex
}

type cacheItem struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

func (c *LRUCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.list.MoveToFront(elem)
		elem.Value.(*cacheItem).value = value
		elem.Value.(*cacheItem).expiresAt = time.Now().Add(c.ttl)
		return
	}

	if c.list.Len() >= c.capacity {
		oldest := c.list.Back()
		delete(c.items, oldest.Value.(*cacheItem).key)
		c.list.Remove(oldest)
	}

	item := &cacheItem{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	elem := c.list.PushFront(item)
	c.items[key] = elem
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(elem.Value.(*cacheItem).expiresAt) {
		c.list.Remove(elem)
		delete(c.items, key)
		return nil, false
	}

	c.list.MoveToFront(elem)
	return elem.Value.(*cacheItem).value, true
}