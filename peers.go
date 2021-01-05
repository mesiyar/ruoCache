package ruoCache

// 节点选择器
type PeerPicker interface {
	//根据传入的 key 选择相应节点
	PickPeer(key string) (peer PeerGetter, ok bool)
}


type PeerGetter interface {
	// 从对应 group 查找缓存值
	Get(group, key string) ([]byte, error)
}
