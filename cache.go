package geecache

import (
	"geecache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex //锁
	lru        *lru.Cache //实际的 lru cache
	cacheBytes int64      //容量
}

func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	//延迟初始化
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Set(key, value)
}

func (c *cache) get(key string) (ByteView, bool) {
	// ! mu 不能只 lock 读锁，因为 lru.Get 存在移动链表的操作，会修改它
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		// * 不要返回 nil, false，会出问题的
		return ByteView{}, false
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return ByteView{}, false
}
