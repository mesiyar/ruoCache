package ruoCache

import (
	"ruoCache/lru"
	"sync"
)

//region cache

type cache struct {
	mutex      sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

// 添加缓存
func (c *cache) add(key string, value ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock() // 自动解锁
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// 获取缓存
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}

//endregion cache

// region getter
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// endregion
