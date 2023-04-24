package list

// 期望值是否相等
type Expected func(a interface{}) bool

// 接收两个参数 i 和 v，分别表示索引和值，并返回一个布尔值。它用于遍历列表中的每个元素
type Consumer func(i int, v interface{}) bool

// 链表接口类型
type List interface {
	Add(val interface{})
	Get(index int) (val interface{})
	Set(index int, val interface{})
	Insert(index int, val interface{})
	Remove(index int) (val interface{})
	RemoveLast() (val interface{})
	RemoveAllByVal(expected Expected) int
	ReverseRemoveByVal(expected Expected, count int) int
	Len() int
	ForEach(consumer Consumer)
	Contains(expected Expected) bool
	Range(start int, stop int) []interface{}
}
