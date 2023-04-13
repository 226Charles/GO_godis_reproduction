package protocol

/*协议中一些常用的固定回复，提前写好
 */

// ping的pong响应
type PongReply struct{}

var pongBytes = []byte("+PONG\r\n")

func (r *PongReply) ToBytes() []byte {
	return pongBytes
}

// ok响应
type OKReply struct{}

var okBytes = []byte("+OK\r\n")

func (r *OKReply) ToBytes() []byte {
	return okBytes
}

var theOKReply = new(OKReply)

func MakeOKReply() *OKReply {
	return theOKReply
}

// -1 表示 NULL Bulk，用于表示不存在的键或者值为 NULL 的情况
var nullBulkBytes = []byte("$-1\r\n")

type NullBulkReply struct{}

func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

// 0 表示空列表，用于表示一个空的数组回复
var emptyMultiBulkBytes = []byte("*0\r\n")

type EmptyMultiBulkReply struct{}

func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}

// subscribe命令中可用的空回复
type NoReply struct{}

var noBytes = []byte("")

func (r *NoReply) ToBytes() []byte {
	return noBytes
}

// 加入事务队列
type QueuedReply struct{}

var queuedBytes = []byte("+QUEUED\r\n")

func (r *QueuedReply) ToBytes() []byte {
	return queuedBytes
}

var theQueuedReply = new(QueuedReply)

func MakeQueuedReply() *QueuedReply {
	return theQueuedReply
}
