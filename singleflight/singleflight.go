package singleflight

import "sync"

// 定义请求对象
type call struct {
	wg  sync.WaitGroup // 控制线程是否等待
	val interface{}    // 请求返回结果
	err error          // 错误信息
}

// Group 存储请求对象
type Group struct {
	mu sync.Mutex       // 线程安全必须上锁
	m  map[string]*call // server 端需要记录一下这次请求
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	if c, ok := g.m[key]; ok {
		c.wg.Wait()         // 如果请求正在进行中，则等待
		return c.val, c.err // 请求结束，返回结果
	}
	//创建一个请求对象
	c := new(call)
	// 发起请求前加锁
	c.wg.Add(1)
	// 添加到 g.m，表明 key 已经有对应的请求在处理
	g.m[key] = c

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束

	delete(g.m, key) // 更新 g.m

	return c.val, c.err // 返回结果
}
