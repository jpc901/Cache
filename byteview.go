package Cache

/*
缓存值的抽象与封装
*/

// ByteView 只读数据结构
type ByteView struct {
	b []byte // 存储真实缓存值
}

// Len 返回其所占的内存大小。
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 以字符串形式返回数据
func (v ByteView) String() string {
	return string(v.b)
}

// 拷贝
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
