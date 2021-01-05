// q: 当一个节点收到请求,此时该节点没有数据, 数据需要从哪儿获取
// 假设第一次随机选取了节点 1 ，节点 1 从数据源获取到数据的同时缓存该数据；那第二次，只有 1/10 的可能性再次选择节点 1,
// 有 9/10 的概率选择了其他节点，如果选择了其他节点，就意味着需要再一次从数据源获取数据，
// 一般来说，这个操作是很耗时的。这样做，一是缓存效率低，二是各个节点上存储着相同的数据，浪费了大量的存储空间。
//那有什么办法，对于给定的 key，每一次都选择同一个节点呢？使用 hash 算法也能够做到这一点。
//那把 key 的每一个字符的 ASCII 码加起来，再除以 10 取余数可以吗？当然可以，这可以认为是自定义的 hash 算法。
// q:当节点数量变化了怎么办?
// 简单求取 Hash 值解决了缓存性能的问题，但是没有考虑节点数量变化的场景。
// 假设，移除了其中一台节点，只剩下 9 个，那么之前 hash(key) % 10 变成了 hash(key) % 9，也就意味着几乎缓存值对应的节点都发生了改变。
// 即几乎所有的缓存值都失效了。节点在接收到对应的请求时，均需要重新去数据源获取数据，容易引起 [缓存雪崩]。
// 为了解决这个问题,可以使用一致性hash来处理
// 一致性hash算法原理:
// 一致性哈希算法将 key 映射到 2^32 的空间中，将这个数字首尾相连，形成一个环。
// 计算节点/机器(通常使用节点的名称、编号和 IP 地址)的哈希值，放置在环上
// 计算 key 的哈希值，放置在环上，顺时针寻找到的第一个节点，就是应选取的节点/机器。

package consistentHash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// map 存储所有的hash keys
type Map struct {
	hash     Hash
	replicas int   // 虚拟节点倍数 ? 解决数据偏移问题而引入
	keys     []int // Sorted
	hashMap  map[int]string
}

// 创建一个hash的实例
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE // 默认hash算法
	}
	return m
}

// 添加
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys) // 排序
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	// 第一步，计算 key 的哈希值。
	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点的下标 idx
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 通过 hashMap 映射得到真实的节点。
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
