package ruoCache

import (
	"fmt"
	"log"
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

// region Group
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

// 新建一个分组的实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}

	mutex.Lock()
	defer mutex.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
	}
	groups[name] = g
	return g
}

// 获取一个分组
func GetGroup(name string) *Group {
	mutex.RLock()
	g := groups[name]
	mutex.RUnlock()
	return g
}

// 从 mainCache 中查找缓存，如果存在则返回缓存值。
//
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[RuoCache] hit")
		return v, nil
	}
	// 缓存不存在，则调用 load 方法
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// endregion
