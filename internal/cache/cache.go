package cache

import (
	"container/list"
	"hash/fnv"
	"sync"
	"time"
)

type cacheItem[V any] struct {
	key       string
	value     V
	expiresAt time.Time
}

type shard[V any] struct {
	mu    sync.RWMutex
	items map[string]*list.Element
	lru   *list.List
}

type ShardedLRU[V any] struct {
	shards   []shard[V]
	capacity int
	ttl      time.Duration
}

func NewShardedLRU[V any](numShards int, capacity int, ttl time.Duration) *ShardedLRU[V] {
	if numShards <= 0 {
		numShards = 16
	}
	s := make([]shard[V], numShards)
	for i := range s {
		s[i] = shard[V]{items: make(map[string]*list.Element), lru: list.New()}
	}
	return &ShardedLRU[V]{shards: s, capacity: capacity / numShards, ttl: ttl}
}

func (c *ShardedLRU[V]) shardFor(key string) *shard[V] {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return &c.shards[int(h.Sum32())%len(c.shards)]
}

func (c *ShardedLRU[V]) Get(key string) (V, bool) {
	s := c.shardFor(key)
	s.mu.RLock()
	defer s.mu.RUnlock()

	if el, ok := s.items[key]; ok {
		it := el.Value.(cacheItem[V])
		if time.Now().After(it.expiresAt) {
			return *new(V), false
		}
		return it.value, true
	}
	return *new(V), false
}

func (c *ShardedLRU[V]) Set(key string, value V) {
	s := c.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	if el, ok := s.items[key]; ok {
		el.Value = cacheItem[V]{key: key, value: value, expiresAt: time.Now().Add(c.ttl)}
		s.lru.MoveToFront(el)
		return
	}

	if s.lru.Len() >= c.capacity && c.capacity > 0 {
		tail := s.lru.Back()
		if tail != nil {
			del := tail.Value.(cacheItem[V]).key
			s.lru.Remove(tail)
			delete(s.items, del)
		}
	}

	el := s.lru.PushFront(cacheItem[V]{key: key, value: value, expiresAt: time.Now().Add(c.ttl)})
	s.items[key] = el
}

func (c *ShardedLRU[V]) Delete(key string) {
	s := c.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	if el, ok := s.items[key]; ok {
		s.lru.Remove(el)
		delete(s.items, key)
	}
}
