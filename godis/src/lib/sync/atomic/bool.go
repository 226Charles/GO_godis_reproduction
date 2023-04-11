package atomic

/*手工制作一个原子操作的bool类型，
利用原子操作的不会因为多个 goroutine 同时访问内存而产生竞争条件
*/

import "sync/atomic"

type Boolean uint32

func (b *Boolean) Get() bool {
	return atomic.LoadUint32((*uint32)(b)) != 0
}

func (b *Boolean) Set(v bool) {
	if v {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}
