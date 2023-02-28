package Cache

import (
	"Cache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_gocache/"
	defaultReplicas = 50
)

type HttpPool struct {
	self        string                 // 记录本地地址和端口
	basePath    string                 // 基础路径
	mu          sync.Mutex             // 保护peer和httpGetters
	peers       *consistenthash.Map    // 根据具体的 key 选择节点
	httpGetters map[string]*httpGetter // 每一个远程节点对应一个 httpGetter
}

// 首先需要定义一个 struct 实现 PeerGetter 接口
type httpGetter struct {
	baseURL string
}

func NewHTTPPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (hp *HttpPool) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// 1. 判断url路径中是否包含 basePath
	url := req.URL.Path
	if !strings.HasPrefix(url, hp.basePath) {
		panic("http request path is invalid")
	}

	// 2. url路径为/_groupcache/<groupname>/<key>  -> <groupname>/<key>
	path := url[len(hp.basePath):]

	// 3. 把<groupname>/<key>字符截断为groupname和key
	parts := strings.SplitN(path, "/", 2)
	groupName := parts[0]
	key := parts[1]

	// 4. 通过groupname获取Group对象
	group := GetGroup(groupName)
	if group == nil {
		http.Error(resp, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	// 5. 使用Group对象方法和key来获取key对应的缓存值
	byteview, err := group.Get(key)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	// 6. 将缓存值作为http body进行响应
	resp.Header().Set("Content-Type", "application/octet-stream")
	resp.Write(byteview.ByteSlice())
}

// Get 实现接口对应的方法, 这个接口能提供访问网络接口拿到缓存数据。
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	// 访问 http 接口的逻辑
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
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

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)

// Set 更新节点列表
func (p *HttpPool) Set(peers ...string) {
	// 因为 hash 环的 map 不是线程安全的,所以这里要加锁.
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	// 调用上一章节的方法, 在 hash 环上添加真实节点和虚拟节点
	p.peers.Add(peers...)
	// 存储远端节点信息
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 根据key选择一个节点
func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	// 因为 hash 环的 map 不是线程安全的,所以这里要加锁.
	p.mu.Lock()
	defer p.mu.Unlock()
	// p.peers 是个 哈希环, 通过调用它的 Get 方法拿到远端节点.
	// 这里的 peer 是个地址.
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HttpPool)(nil)
