package lru

import "container/list"

// Cache 是LRU缓存，对于并发访问是不安全的。
type Cache struct {
	maxBytes int64 // 允许使用的最大内存
	nbytes   int64 // 已经使用的最大内存
	ll       *list.List
	cache    map[string]*list.Element // 键是字符串，值是双向链表中节点的指针
	// 某条记录被移除时的回调函数，可为nil。
	OnEvicted func(key string, value Value)
}

// 节点的数据类型
type entry struct {
	key   string // 淘汰队首节点时，用key删除对应映射
	value Value
}

// Value 用Len来计算它需要多少字节（返回值所占内存大小）
type Value interface {
	Len() int
}

// New 实例化Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

/*
1. 找到字典中对应的双向链表的节点。
2. 将该节点移动到队尾。
*/

// Get 查找功能
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 如果对应的链表存在，将链表中的节点ele移动到队尾
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 删除，缓存淘汰，队头元素就是最近最少访问的节点。
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 队头元素
	if ele != nil {
		c.ll.Remove(ele)                                       // 链表删除队头元素
		kv := ele.Value.(*entry)                               // 元素实体
		delete(c.cache, kv.key)                                // 删除map中的key,value
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 已经使用字节大小
		if c.OnEvicted != nil {                                // OnEvicted不为nil，调用回调函数
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 向缓存中添加一个值。
func (c *Cache) Add(key string, value Value) {
	// 查看是否在字典中
	if ele, ok := c.cache[key]; ok { // 在字典中
		c.ll.MoveToFront(ele) // 移动到队尾
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value // 更新value
	} else { // 不在字典中
		ele := c.ll.PushFront(&entry{key, value}) // 加到队尾
		c.cache[key] = ele                        // 存入字典
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点。
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len 缓存项的个数
func (c *Cache) Len() int {
	return c.ll.Len()
}
