package lfu

import (
	"testing"

	"github.com/matryer/is"
)

func TestSet(t *testing.T) {
	is := is.New(t)

	cache := New(24, nil)
	cache.DelOldest()
	cache.Set("k1", 1)

	v := cache.Get("k1")
	is.Equal(v, 1)

	cache.Del("k1")
	is.Equal(0, cache.Len())
}

func TestOnEvicted(t *testing.T) {
	is := is.New(t)

	keys := make([]string, 0, 8)
	onEvicted := func(key string, value interface{}) {
		keys = append(keys, key)
	}
	cache := New(8, onEvicted)

	cache.Set("k1", int32(1))
	cache.Set("k2", int32(2))
	cache.Get("k1")
	cache.Set("k3", int32(3))
	cache.Get("k3")
	cache.Set("k4", int32(4))

	expected := []string{"k2", "k1"}

	is.Equal(expected, keys)
	is.Equal(2, cache.Len())
}
