package list

/*QuickList表现效果比List好
 */

//GoLang内置包
import "container/list"

const pageSize = 1024

type QuickList struct {
	data *list.List
	size int
}

// [-1, ql.Len()]之间
type iterator struct {
	node   *list.Element
	offset int
	ql     *QuickList
}

func NewQuickList() *QuickList {
	l := &QuickList{data: list.New()}
	return l
}

// 分页机制，其中每个页面的容量为 pageSize，并且在添加元素时自动创建新的页面
func (ql *QuickList) Add(val interface{}) {
	ql.size++
	if ql.data.Len() == 0 {
		page := make([]interface{}, 0, pageSize)
		page = append(page, val)
		ql.data.PushBack(page)
		return
	}
	backNode := ql.data.Back()
	backPage := backNode.Value.([]interface{})
	if len(backPage) == cap(backPage) {
		page := make([]interface{}, 0, pageSize)
		page = append(page, val)
		ql.data.PushBack(page)
		return
	}
	backPage = append(backPage, val)
	backNode.Value = backPage
}

// 根据页表管理找到index坐标所在的结点页数iterator
func (ql *QuickList) find(index int) *iterator {
	if ql == nil {
		panic("list is nil")
	}
	if index < 0 || index >= ql.size {
		panic("index out of bound")
	}
	var n *list.Element
	var page []interface{}
	var pageBeg int
	if index < ql.size/2 {
		n = ql.data.Front()
		pageBeg = 0
		for {
			page = n.Value.([]interface{})
			if pageBeg+len(page) > index {
				break
			}
			pageBeg += len(page)
			n = n.Next()
		}
	} else {
		n = ql.data.Back()
		pageBeg = ql.size
		for {
			page = n.Value.([]interface{})
			pageBeg -= len(page)
			if pageBeg <= index {
				break
			}
			n = n.Prev()
		}
	}
	pageOffset := index - pageBeg
	return &iterator{
		node:   n,
		offset: pageOffset,
		ql:     ql,
	}
}

func (iter *iterator) get() interface{} {
	return iter.page()[iter.offset]
}

func (iter *iterator) page() []interface{} {
	return iter.node.Value.([]interface{})
}

// 检查iter是否下一个越总list的边界
func (iter *iterator) next() bool {
	page := iter.page()
	if iter.offset < len(page)-1 {
		iter.offset++
		return true
	}
	if iter.node == iter.ql.data.Back() {
		iter.offset = len(page)
		return false
	}
	iter.offset = 0
	iter.node = iter.node.Next()
	return true
}

// 找到前一个，做边界检测
func (iter *iterator) prev() bool {
	if iter.offset > 0 {
		iter.offset--
		return true
	}
	if iter.node == iter.ql.data.Front() {
		iter.offset = -1
		return false
	}
	iter.node = iter.node.Prev()
	prevPage := iter.node.Value.([]interface{})
	iter.offset = len(prevPage) - 1
	return true
}

// 判断是否是最后一位
func (iter *iterator) atEnd() bool {
	if iter.ql.data.Len() == 0 {
		return true
	}
	if iter.node != iter.ql.data.Back() {
		return false
	}
	page := iter.page()
	return iter.offset == len(page)
}

// 判断是否是第一位
func (iter *iterator) atBegin() bool {
	if iter.ql.data.Len() == 0 {
		return true
	}
	if iter.node != iter.ql.data.Front() {
		return false
	}
	return iter.offset == -1
}

func (ql *QuickList) Get(index int) (val interface{}) {
	iter := ql.find(index)
	return iter.get()
}

func (iter *iterator) set(val interface{}) {
	page := iter.page()
	page[iter.offset] = val
}

func (ql *QuickList) Set(index int, val interface{}) {
	iter := ql.find(index)
	iter.set(val)
}

// 插入操作，如果当前结点page满了，拆成两半，插入后将两半的结点插入回原链表
func (ql *QuickList) Insert(index int, val interface{}) {
	if index == ql.size {
		ql.Add(val)
		return
	}
	iter := ql.find(index)
	page := iter.node.Value.([]interface{})
	if len(page) < pageSize {
		page = append(page[:iter.offset+1], page[iter.offset:]...)
		page[iter.offset] = val
		iter.node.Value = page
		ql.size++
		return
	}
	var nextPage []interface{}
	nextPage = append(nextPage, page[pageSize/2:]...)
	page = page[:pageSize/2]
	if iter.offset < len(page) {
		page = append(page[:iter.offset+1], page[iter.offset:]...)
		page[iter.offset] = val
	} else {
		i := iter.offset - pageSize/2
		nextPage = append(nextPage[:i+1], nextPage[i:]...)
		nextPage[i] = val
	}
	iter.node.Value = page
	ql.data.InsertAfter(nextPage, iter.node)
	ql.size++
}

// 删除元素，如果越界跳到下一个节点
func (iter *iterator) remove() interface{} {
	page := iter.page()
	val := page[iter.offset]
	page = append(page[:iter.offset], page[iter.offset+1:]...)
	if len(page) > 0 {
		iter.node.Value = page
		if iter.offset == len(page) {
			if iter.node != iter.ql.data.Back() {
				iter.node = iter.node.Next()
				iter.offset = 0
			}
		}
	} else {
		if iter.node == iter.ql.data.Back() {
			iter.ql.data.Remove(iter.node)
			iter.node = nil
			iter.offset = 0
		} else {
			nextNode := iter.node.Next()
			iter.ql.data.Remove(iter.node)
			iter.node = nextNode
			iter.offset = 0
		}
	}
	iter.ql.size--
	return val
}

func (ql *QuickList) Remove(index int) interface{} {
	iter := ql.find(index)
	return iter.remove()
}

func (ql *QuickList) Len() int {
	return ql.size
}

func (ql *QuickList) RemoveLast() interface{} {
	if ql.Len() == 0 {
		return nil
	}
	ql.size--
	lastNode := ql.data.Back()
	lastPage := lastNode.Value.([]interface{})
	if len(lastPage) == 1 {
		ql.data.Remove(lastNode)
		return lastPage[0]
	}
	val := lastPage[len(lastPage)-1]
	lastPage = lastPage[:len(lastPage)-1]
	lastNode.Value = lastPage
	return val
}

func (ql *QuickList) RemoveAllByVal(expected Expected) int {
	iter := ql.find(0)
	removed := 0
	for !iter.atEnd() {
		if expected(iter.get()) {
			iter.remove()
			removed++
		} else {
			iter.next()
		}
	}
	return removed
}

func (ql *QuickList) RemoveByVal(expected Expected, count int) int {
	if ql.size == 0 {
		return 0
	}
	iter := ql.find(0)
	removed := 0
	for !iter.atEnd() {
		if expected(iter.get()) {
			iter.remove()
			removed++
			if removed == count {
				break
			}
		} else {
			iter.next()
		}
	}
	return removed
}

func (ql *QuickList) ReverseRemoveByVal(expected Expected, count int) int {
	if ql.size == 0 {
		return 0
	}
	iter := ql.find(ql.size - 1)
	removed := 0
	for !iter.atBegin() {
		if expected(iter.get()) {
			iter.remove()
			removed++
			if removed == count {
				break
			}
		}
		iter.prev()
	}
	return removed
}

func (ql *QuickList) ForEach(consumer Consumer) {
	if ql == nil {
		panic("list is nil")
	}
	if ql.Len() == 0 {
		return
	}
	iter := ql.find(0)
	i := 0
	for {
		goNext := consumer(i, iter.get())
		if !goNext {
			break
		}
		i++
		if !iter.next() {
			break
		}
	}
}

func (ql *QuickList) Contains(expected Expected) bool {
	contains := false
	ql.ForEach(func(i int, v interface{}) bool {
		if expected(v) {
			contains = true
			return false
		}
		return true
	})
	return contains
}

func (ql *QuickList) Range(start int, stop int) []interface{} {
	if start < 0 || start >= ql.Len() {
		panic("`start` out of range")
	}
	if stop < start || stop > ql.Len() {
		panic("`stop` out of range")
	}
	sliceSize := stop - start
	slice := make([]interface{}, 0, sliceSize)
	iter := ql.find(start)
	i := 0
	for i < sliceSize {
		slice = append(slice, iter.get())
		iter.next()
		i++
	}
	return slice
}
