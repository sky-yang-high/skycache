package lru_test

import (
	"reflect"
	"skycache/lru"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	c := lru.New(0, nil)
	c.Set("1", String("val1"))
	if v, ok := c.Get("1"); !ok || string(v.(String)) != "val1" {
		t.Fatalf("cache hit key 1 failed, expecetd val1, got %s", v.(String))
	}

	if v, ok := c.Get("2"); ok {
		t.Fatalf("cache hit key 2 failed, expected nil, got %s", v.(String))
	}
}

func TestSet(t *testing.T) {
	c := lru.New(0, nil)
	key1, key2 := "k1", "k2"
	val1, val2, val3 := "val1", "val2", "val3"
	c.Set(key1, String(val1))
	c.Set(key2, String(val2))

	if v, ok := c.Get(key1); !ok || string(v.(String)) != val1 {
		t.Fatalf("cache hit key 1 failed, expecetd val1, got %s", v.(String))
	}

	//更新 key1
	c.Set(key1, String(val3))
	if v, ok := c.Get(key1); !ok || string(v.(String)) != val3 {
		t.Fatalf("cache hit key 1 failed, expecetd val3, got %s", v.(String))
	}
}

func TestRemoveOldest(t *testing.T) {
	key1, key2, key3 := "k1", "k2", "k3"
	val1, val2, val3 := "val1", "val2", "val3"
	size := len(key1) + len(key2) + len(val1) + len(val2)

	c := lru.New(int64(size), nil)
	c.Set(key1, String(val1))
	c.Set(key2, String(val2))

	// 溢出，应该 remove key1-val1
	c.Set(key3, String(val3))

	if v, ok := c.Get(key1); ok || c.Len() != 2 {
		t.Fatalf("cache remove failed, should remove k1 ,but got c[k1]=%s", v.(String))
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value lru.Value) {
		keys = append(keys, key)
	}

	c := lru.New(int64(10), callback)
	c.Set("key1", String("123456"))
	c.Set("k2", String("k2"))
	c.Set("k3", String("k3"))
	c.Set("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
