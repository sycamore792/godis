package atomic

import "sync/atomic"

// 模拟一个锁？ 通过原子操作保证不变，用uint32是因为atomic没有针对bool的 ，且最小支持uint32。
type Boolean uint32

// 只要拿到非0值就是true？   是一个cnt的功能吗？
// 用1表示关闭 0表示开启？ 使用它的是一个isClosing 因此1表示已关闭，0表示没关闭
func (b *Boolean) Get() bool {
	return atomic.LoadUint32((*uint32)(b)) != 0
}

// 是用false表示开启 用true表示关闭？
func (b *Boolean) Set(v bool) {
	if v {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}
