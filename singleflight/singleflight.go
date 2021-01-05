package singleflight

import "sync"

// 正在进行中，或已经结束的请求。使用 sync.WaitGroup 锁避免重入。
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mutex sync.Mutex // protects m
	m     map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mutex.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mutex.Unlock()
		c.wg.Wait()
		return c.val, nil
	}

	c := new(call)
	c.wg.Add(1) // 发起请求前加锁
	g.m[key] = c
	g.mutex.Unlock()

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束

	g.mutex.Lock()
	delete(g.m, key)
	g.mutex.Unlock()

	return c.val, c.err
}
