package geecache

import (
	"geecache/lru"
	"sync"
)

type cache struct {
	mu         sync.RWMutex //读写锁
	lru        *lru.Cache   //实际的 lru cache
	cacheBytes int64        //容量
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
	c.mu.RLock() //只 lock r，允许其他 读，但不允许 写
	defer c.mu.RUnlock()

	if c.lru == nil {
		// * 不要返回 nil, false，会出问题的
		return ByteView{}, false
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return ByteView{}, false
}
