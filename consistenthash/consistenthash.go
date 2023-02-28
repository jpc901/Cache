package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash hash函数类型
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // hash函数，crc32哈希
	replicas int            // 虚拟节点背书
	keys     []int          // 哈希环，环中元素是已经排序的
	hashMap  map[int]string // 虚拟节点和真实节点的映射关系
}

// New 创建一个Map实例
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string), // 初始化map函数，map必须要进行初始化
	}

	// 默认用crc32
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 可以一次性添加多个缓存服务器
func (m *Map) Add(keys ...string) {
	for _, key := range keys {

		// 虚拟节点名称为序号+key名称
		for i := 0; i < m.replicas; i++ {

			// 计算虚拟节点的哈希值，进行类型转换
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))

			// 添加到哈希环中
			m.keys = append(m.keys, hash)

			// 维护虚拟节点到真实节点之间的距离
			m.hashMap[hash] = key
		}
	}
	// 对哈希环进行排序
	sort.Ints(m.keys)
}

// Get 选择缓存节点函数
func (m *Map) Get(key string) string {
	// 校验数据有效性
	if len(m.keys) == 0 {
		return ""
	}

	// 计算key的哈希值
	hash := int(m.hash([]byte(key)))
	// 顺时针找到哈希环上的第一个匹配的虚拟节点的下标
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 通过hashMap定位到真实的缓存节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
