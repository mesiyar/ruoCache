package ruoCache

import (
	"fmt"
	"log"
	pb "ruoCache/ruoCachePb"
	"ruoCache/singleflight"
	"sync"
)

// region Group
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker

	loader *singleflight.Group
}

//
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 确保并发情况下 同一个key只被调用一次
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[ruoCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{Group: g.name, Key: key}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

var (
	mutex  sync.RWMutex
	Groups = make(map[string]*Group)
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
		loader: &singleflight.Group{},
	}
	Groups[name] = g
	log.Printf("new Group %s", name)
	return g
}

// 获取一个分组
func GetGroup(name string) *Group {
	mutex.RLock()
	g := Groups[name]
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

func (g *Group) Set(key , value string)  {
	v := ByteView{b: cloneBytes([]byte(value))}
	g.mainCache.add(key, v)
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
