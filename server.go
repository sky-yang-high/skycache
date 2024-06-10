package skycache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"skycache/consistenthash"
	pb "skycache/skycachepb"

	"google.golang.org/protobuf/proto"
)

// 为 geecache 之间提供通信能力
// 这样部署在其他机器上的 cache 可以通过访问server获取缓存

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// 确保 实现了对应的接口
var _ PeerPicker = (*Server)(nil)

// 和Group 解耦合，实现
type Server struct {
	addr        string     // ip:port 的形式
	basePath    string     //节点通信的 前缀URL
	mu          sync.Mutex // guards peers and httpGetters
	peers       *consistenthash.HashMap
	httpGetters map[string]*Client // keyed by e.g. "http://10.0.0.2:8008"
}

func NewHTTPPool(self string) *Server {
	return &Server{
		addr:     self,
		basePath: defaultBasePath,
	}
}

func (p *Server) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.addr, fmt.Sprintf(format, v...))
}

// 实现节点间的通信，在自己的 groups 下查找对应的 group 和 key
// 返回 err(如果未找到) 或 value
func (p *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		// ? should i panic in this case?
		log.Println(r.URL.Path, p.basePath)
		http.Error(w, "no such source", http.StatusNotFound)
		return
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	//should be url like <basePath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupname, key := parts[0], parts[1]

	group := groups[groupname]
	if group == nil {
		http.Error(w, "no such group", http.StatusInternalServerError)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//改为使用 protobuf 传输数据
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	//log.Println(view)
	w.Write(body)
}

// Set updates the pool's list of peers.
func (p *Server) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*Client, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &Client{target: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
func (p *Server) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.addr {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
