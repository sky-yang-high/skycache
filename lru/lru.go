package lru

import "container/list"

//LRU Cache
type Cache struct {
	maxBytes  int64                         //最大容量，为 0 则无上限(表现在不会主动淘汰记录)
	nBytes    int64                         //当前容量
	ll        *list.List                    //存储value
	cache     map[string]*list.Element      //存储key-value对
	OnEvicted func(key string, value Value) //回调函数，可选
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, OnEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

// 查询，视为一次使用，更新其位置
func (c *Cache) Get(key string) (Value, bool) {
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e) //移到队首
		kv := e.Value.(*entry)
		return kv.value, ok
	}
	return nil, false
}

//写入，可能是新增/删除，同样需要更新位置
func (c *Cache) Set(key string, value Value) {
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		kv.value = value
		c.nBytes += (int64(value.Len()) - int64(kv.value.Len()))
	} else {
		// ! 记得和上面保持一致，e.Value 是 *entry 类型
		e := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = e

		// ? maybe += int64(len(e)) is better
		c.nBytes += int64(len(key) + value.Len())
	}
	//写入后，可能短时间内占用空间溢出，清除
	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.removeOldest()
	}
}

//根据 lru 策略，淘汰最近最少使用的一项
func (c *Cache) removeOldest() {
	e := c.ll.Back()
	if e != nil {
		c.ll.Remove(e)
		kv := e.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key) + kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Remove(key string) {
	if e, ok := c.cache[key]; ok {
		kv := e.Value.(*entry)
		c.ll.Remove(e)
		delete(c.cache, key)
		c.nBytes -= int64(len(kv.key) + kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
