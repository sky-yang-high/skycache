package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type Hash func(data []byte) uint32

type HashMap struct {
	hash     Hash           //哈希函数
	replicas uint           //节点放大倍数，即一个节点 -> r 个虚拟节点
	keys     []int          //有序的虚拟节点 哈希值，定义为int 是方便后面 sort
	keyMap   map[int]string //虚拟节点 -> 真实节点
	mu       sync.Mutex     //锁，控制并发
}

// 返回一个 一致哈希结构，若 fn 为 nil，则使用 crc编码
func New(replicas uint, fn Hash) *HashMap {
	if fn == nil {
		fn = crc32.ChecksumIEEE
	}

	return &HashMap{
		hash:     fn,
		replicas: replicas,
		keyMap:   make(map[int]string),
	}
}

// 增加节点，keys是节点名称，暂时忽略 哈希冲突
func (h *HashMap) Add(keys ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, key := range keys {
		for i := 0; i < int(h.replicas); i++ {
			hashKey := h.hash([]byte(strconv.Itoa(i) + key))
			h.keys = append(h.keys, int(hashKey))
			h.keyMap[int(hashKey)] = key
		}
	}
	sort.Ints([]int(h.keys))
}

// 然后 哈希环上 key 对应的 节点名称，key 是资源名称
func (h *HashMap) Get(key string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	// ? maybe panic is better ?
	if len(h.keys) == 0 {
		panic("empty hashMap keys")
	}

	hashKey := h.hash([]byte(key))

	// * sort.Search 返回的 index 满足条件是在 i<index，f(i) 都是 false
	// * i>=index，f(i) 都是 true
	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= int(hashKey)
	})

	// 成环，要对 len(h.keys) 取模，即 idx = n 时，选中节点0
	return h.keyMap[h.keys[idx%len(h.keys)]]
}

// 去掉该节点
func (h *HashMap) Remove(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i := 0; i < int(h.replicas); i++ {
		hashKey := h.hash([]byte(strconv.Itoa(i) + key))
		idx := sort.SearchInts(h.keys, int(hashKey))
		//去掉该虚拟节点
		h.keys = append(h.keys[:idx], h.keys[idx+1:]...)
		delete(h.keyMap, int(hashKey))
	}

}
