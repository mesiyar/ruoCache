package lru

import "container/list"

type Cache struct {
	maxBytes int64 //允许使用的最大内存，
	nbytes   int64 // 当前已使用的内存

	ll    *list.List // 双向列表
	cache map[string]*list.Element

	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// 方便实例化cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 如果键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 能获取到缓存 则将其移动到队尾
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 移出元素
func (c *Cache) RemoveOldest() {
	// 如果列表为空，则返回 / 返回列表l或nil的最后一个元素。
	ele := c.ll.Back()

	if ele != nil {
		c.ll.Remove(ele) // 移出元素
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                                // 从缓存中删除
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 重新计算当前已用内存
		// 如果回调函数 OnEvicted 不为 nil，则调用回调函数。
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 添加缓存
func (c *Cache) Add(key string, value Value) {
	// 元素已存在 则更新
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 当内存不足时, 检测并移出没有使用的
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}
