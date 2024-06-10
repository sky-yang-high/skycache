package singleflight

import "sync"

// 使用 sync.WaitGroup 来防止缓存穿透

// 一个 call 表示发起一次请求
type call struct {
	wg    sync.WaitGroup
	value interface{}
	err   error
}

type Group struct {
	mu sync.Mutex //包含m
	m  map[string]*call
}

func (g Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 当前时刻，已经发起了对 key 的请求
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() //任意时刻，wg 最多被 add 1
		//到这里，即请求完毕
		return c.value, c.err
	}

	// 第一次发起请求
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.value, c.err = fn()
	c.wg.Done()

	// 请求完毕，更新 g.m
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.value, c.err
}
