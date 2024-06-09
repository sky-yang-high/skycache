package geecache

import (
	"errors"
	"log"
	"sync"
)

// Getter接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// 类似于 HTTP.HandleFunc，实现 Get 方法
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 创建新 group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := groups[name]; ok {
		panic("repeat group name")
	}

	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
			//mu:         sync.RWMutex{},
		},
	}
	groups[name] = g
	return g
}

// 返回 名为 name 的 group，若不存在则返回 nil
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

// 用户获取(和被动添加) key-value
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is must")
	}

	//尝试从 cache 中获取
	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[Cache %s] hits\n", g.name)
		return v, nil
	}

	//尝试另外两种方式
	return g.load(key)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}

	return g.getLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// 调用 getter 从用户处获取源数据
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneByte(bytes)}
	g.polulateCache(key, value)
	return value, nil
}

func (g *Group) polulateCache(key string, value ByteView) {
	g.mainCache.set(key, value)
}

// 主动添加数据
func (g *Group) Set(key string, value ByteView) {
	g.mainCache.set(key, value)
}
