package skycache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// 为 cache 实现访问远程节点的能力

var _ PeerGetter = (*Client)(nil)

// 实现 Getter 接口，使用 HTTP 访问 url 获得对应的资源
type Client struct {
	target string // ip:port + basePath
}

// 和 远程节点通信，远程节点进入 ServeHTTP 响应
func (h *Client) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.target,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}
