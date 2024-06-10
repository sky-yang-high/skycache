package skycache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	pb "skycache/skycachepb"

	"google.golang.org/protobuf/proto"
)

// 为 cache 实现访问远程节点的能力

var _ PeerGetter = (*Client)(nil)

// 实现 Getter 接口，使用 HTTP 访问 url 获得对应的资源
type Client struct {
	target string // ip:port + basePath
}

// 和 远程节点通信，远程节点进入 ServeHTTP 响应
func (h *Client) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.target,
		url.QueryEscape(in.Group),
		url.QueryEscape(in.Key),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}
