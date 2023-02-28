package Cache

import (
	"fmt"
	"log"
	"sync"
)

/*
负责与外部交互，控制缓存存储和获取的主流程
设计了一个回调函数，在缓存不存在时，调用这个函数，得到源数据。
*/

// Getter 通过回调函数加载数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 用一个函数实现Getter。
type GetterFunc func(key string) ([]byte, error)

// Get 实现 Getter 接口方法
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 是一个缓存的命名空间
type Group struct {
	name      string
	getter    Getter // 缓存未命中时获取源数据的回调(callback)
	mainCache cache  // 并发缓存
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建Group实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil { // 缓存未命中
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup 返回之前用NewGroup创建的命名组，如果没有这样的组则为nil。
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 从缓存中获取一个键的值
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 命中会直接返回
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GoCache] hit")
		return v, nil
	}

	// 未命中
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

// 获取本地数据源
func (g *Group) getLocally(key string) (ByteView, error) {

	//  调用用户回调函数 g.getter.Get() 获取源数据
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}

	// 将源数据添加到缓存 mainCache 中
	g.populateCache(key, value)
	return value, nil
}

// 将源数据添加到缓存 mainCache 中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
