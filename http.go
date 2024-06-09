package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache/"

type HTTPPool struct {
	self     string //自身的URL
	basePath string //节点通信的 前缀URL
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// 实现节点间的通信，在自己的 groups 下查找对应的 group 和 key
// 返回 err(如果未找到) 或 value
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/octet-stream")
	log.Println(view)
	w.Write(view.ByteSlice())
}
