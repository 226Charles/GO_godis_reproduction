package lock

/* 弥补concurrent的单key锁不足，可满足多锁，如MSETNX操作
在进行读写锁处理时，先进行顺序排序，然后依次按序进行处理，保证了不会冲突 但时间消耗增加
*/

import (
	"sort"
	"sync"
)

const prime32 = uint32(16777619)

// 读写锁
type Locks struct {
	table []*sync.RWMutex
}

func Make(tableSize int) *Locks {
	table := make([]*sync.RWMutex, tableSize)
	for i := 0; i < tableSize; i++ {
		table[i] = &sync.RWMutex{}
	}
	return &Locks{table: table}
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func (locks *Locks) spread(hashCode uint32) uint32 {
	if locks == nil {
		panic("dict is nil")
	}
	tableSize := uint32(len(locks.table))
	return (tableSize - 1) & uint32(hashCode)
}

func (locks *Locks) Lock(key string) {
	index := locks.spread(fnv32(key))
	mu := locks.table[index]
	mu.Lock()
}

func (locks *Locks) RLock(key string) {
	index := locks.spread(fnv32(key))
	mu := locks.table[index]
	mu.RLock()
}

func (locks *Locks) UnLock(key string) {
	index := locks.spread(fnv32(key))
	mu := locks.table[index]
	mu.Unlock()
}

func (locks *Locks) RUnLock(key string) {
	index := locks.spread(fnv32(key))
	mu := locks.table[index]
	mu.RUnlock()
}

func (locks *Locks) toLockIndices(keys []string, reverse bool) []uint32 {
	indexMap := make(map[uint32]bool)
	for _, key := range keys {
		index := locks.spread(fnv32(key))
		indexMap[index] = true
	}
	indices := make([]uint32, 0, len(indexMap))
	for index := range indexMap {
		indices = append(indices, index)
	}
	sort.Slice(indices, func(i, j int) bool {
		if !reverse {
			return indices[i] < indices[j]
		}
		return indices[i] > indices[j]
	})
	return indices
}

func (locks *Locks) Locks(keys ...string) {
	indices := locks.toLockIndices(keys, false)
	for _, index := range indices {
		mu := locks.table[index]
		mu.Lock()
	}
}

func (locks *Locks) RLocks(keys ...string) {
	indices := locks.toLockIndices(keys, false)
	for _, index := range indices {
		mu := locks.table[index]
		mu.RLock()
	}
}

func (locks *Locks) URLocks(keys ...string) {
	indices := locks.toLockIndices(keys, false)
	for _, index := range indices {
		mu := locks.table[index]
		mu.Unlock()
	}
}

func (locks *Locks) URnLocks(keys ...string) {
	indices := locks.toLockIndices(keys, false)
	for _, index := range indices {
		mu := locks.table[index]
		mu.RUnlock()
	}
}

// 区分读和写锁
func (locks *Locks) RWLocks(writeKeys []string, readKeys []string) {
	keys := append(writeKeys, readKeys...)
	indices := locks.toLockIndices(keys, false)
	writeIndexSet := make(map[uint32]struct{})
	for _, wKey := range writeKeys {
		idx := locks.spread(fnv32(wKey))
		writeIndexSet[idx] = struct{}{}
	}
	for _, index := range indices {
		_, w := writeIndexSet[index]
		mu := locks.table[index]
		if w {
			mu.Lock()
		} else {
			mu.RLock()
		}
	}
}

func (locks *Locks) RWUnLocks(writeKeys []string, readKeys []string) {
	keys := append(writeKeys, readKeys...)
	indices := locks.toLockIndices(keys, false)
	writeIndexSet := make(map[uint32]struct{})
	for _, wKey := range writeKeys {
		idx := locks.spread(fnv32(wKey))
		writeIndexSet[idx] = struct{}{}
	}
	for _, index := range indices {
		_, w := writeIndexSet[index]
		mu := locks.table[index]
		if w {
			mu.Unlock()
		} else {
			mu.RUnlock()
		}
	}
}
