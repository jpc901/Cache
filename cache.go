package Cache

import (
	"Cache/lru"
	"sync"
)

/*
主要是进行并发控制
*/

// 支持并发读写的lru缓存
type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

// 向缓存中添加数据
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 使用延迟初始化，将该对象的创建延迟至第一次使该对象
	// 主要用于提高性能，并且减少程序内存要求
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// 从缓存中读取数据
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
