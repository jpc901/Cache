package Cache

import (
	"net/http"
	"strings"
)

const defaultBasePath = "/_gocache/"

type HttpPool struct {
	self     string // 记录本地地址和端口
	basePath string // 基础路径
}

func NewHTTPPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
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
