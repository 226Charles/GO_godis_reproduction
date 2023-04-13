package redis

/*客户端连接接口
 */

type Connection interface {
	Write([]byte) (int, error)
	Close() error

	SetPassword(string)
	GetPassword() string

	Subscribe(channel string)
	UnSubsrcibe(channel string)
	SubsCount() int
	GetChannels() []string

	InMultiState() bool
	SetMultiState(bool)
	GetQueuedCmdLine() [][][]byte
	EnqueueCmd([][]byte)
	ClearQueuedCmds()
	GetWatching() map[string]uint32
	AddTxError(err error)
	GetTxErrors() []error

	GetDBIndex() int
	SelectDB(int)

	SetSlave()
	IsSlave() bool

	SetMaster()
	IsMaster() bool

	Name() string
}
