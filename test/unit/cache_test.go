package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wb-test-task/internal/cache"
)

func TestShardedLRU_SetGet(t *testing.T) {
	c := cache.NewShardedLRU[*int](4, 10, time.Minute)
	v := 42
	c.Set("k1", &v)
	got, ok := c.Get("k1")
	require.True(t, ok)
	require.NotNil(t, got)
	assert.Equal(t, 42, *got)
}

func TestShardedLRU_TTL(t *testing.T) {
	c := cache.NewShardedLRU[*int](2, 10, 10*time.Millisecond)
	v := 1
	c.Set("k", &v)
	time.Sleep(20 * time.Millisecond)
	_, ok := c.Get("k")
	assert.False(t, ok, "срок действия истечет за TTL")
}

func TestShardedLRU_Eviction(t *testing.T) {
	c := cache.NewShardedLRU[*int](2, 2, time.Minute)
	v1, v2, v3 := 1, 2, 3
	c.Set("a", &v1)
	c.Set("b", &v2)
	c.Set("c", &v3)

	cnt := 0
	for _, k := range []string{"a", "b", "c"} {
		if _, ok := c.Get(k); ok {
			cnt++
		}
	}
	assert.LessOrEqual(t, cnt, 2)
}

func TestShardedLRU_Delete(t *testing.T) {
	c := cache.NewShardedLRU[*int](2, 10, time.Minute)
	v := 7
	c.Set("x", &v)
	c.Delete("x")
	_, ok := c.Get("x")
	assert.False(t, ok)
}
