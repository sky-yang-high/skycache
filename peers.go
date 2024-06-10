package skycache

// 实现 根据 key,使用一致性哈希 选择相应的节点 的能力
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 实现 访问 group 和 key 获取对应的 value 的能力
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
