# Godis-RESP协议

RESP协议是我第一次系统了解过的协议，也是非常简单的一种协议，在具体了解过后我用为数不多的工程经验觉得或许2个.go文件就能解决，作者用了x个来完成，可见我工程结构的逻辑性还是有很大欠缺。

简单介绍一下RESP协议，之后再分别介绍一下这部分内容实现的各个文件。

# 1.RESP协议基本内容

- `简单字符串`：Simple Strings，第一个字节响应 `+`
- `错误`：Errors，第一个字节响应 `-`
- `整型`：Integers，第一个字节响应 `:`
- `批量字符串`：Bulk Strings，第一个字节响应 `$`
- `数组`：Arrays，第一个字节响应 `*`

具体例子:

```Go
*4
$6
PSETEX
$24
test_redisson_batch_key8
$6
120000
$32
test_redisson_batch_key=>value:8
```

其中提到了二进制安全这个概念，再直接借用一下作者的链接：[Golang 实现 Redis(2): 实现 Redis 协议解析器 - -Finley- - 博客园](https://www.cnblogs.com/Finley/p/11923168.html)

> 二进制安全是指允许协议中出现任意字符而不会导致故障。比如 C 语言的字符串以 `\0` 作为结尾不允许字符串中间出现`\0`, 而 Go 语言的 string 则允许出现 `\0`，我们说 Go 语言的 string 是二进制安全的，而 C 语言字符串不是二进制安全的。

RESP 的二进制安全性允许我们在 key 或者 value 中包含 `\r` 或者 `\n` 这样的特殊字符。在使用 redis 存储 protobuf、msgpack 等二进制数据时，二进制安全性尤为重要。

在实现中，包含以下：

- 简单字符串(Simple String): 服务器用来返回简单的结果，比如"OK"。非二进制安全，且不允许换行。
- 错误信息(Error): 服务器用来返回简单的错误信息，比如"ERR Invalid Synatx"。非二进制安全，且不允许换行。
- 整数(Integer): llen、scard 等命令的返回值, 64位有符号整数
- 字符串(Bulk String): 二进制安全字符串, 比如 get 等命令的返回值
- 数组(Array, 又称 Multi Bulk Strings): Bulk String 数组，客户端发送指令以及 lrange 等命令响应的格式

# 2.协议实现具体内容

这次实现的代码包含以下几部分

/lib/utils 下的rand_string.go 以及utils.go两个帮助实现协议的工具包，再就是核心内容/redis/parser/parser.go

东西不多，这周做别的事去了，拖拉了很久，下面依次介绍一下

## rand_string.go

这是实现了一个利用时间种子随机生成字符串和十六进制字符串，rand.New(rand.NewSource(time.Now().UnixNano()))给出根据当前时间的时间种子，然后用RandString()以及RandHexString()分别生成全字母数字以及十六进制的随机字符串生成，比较简单

## utils.go

实现了三种cmd的转换，ToCmdLine()将字符串转换成[][]byte，ToCmdLine2()实现了包含命令名称、以及后续string内容的转换，成为[][]byte，ToCmdLine3()和2的区别在于内容不是string是[]byte

同时还实现了任意类型的判等，因为GO内部并没有实现对[]byte的判等，如果给出两种类型的判等的话，代表的是各自byte数组的地址，而非数组内的值，所以需要手工实现，逻辑比较简单，判空、判长度、依次判等

在Go中，可以使用`==`运算符对以下类型进行直接判等：

- 数值类型（如int，float等）
- 布尔类型（true和false）
- 字符串类型
- 指针类型
- channel类型
- 接口类型（除非其动态类型为函数类型）
- 数组类型（当且仅当数组元素类型可进行直接判等）
- 结构体类型（当且仅当结构体中的所有字段都可进行直接判等）

注意，切片类型、map类型、函数类型和复合类型（如结构体、数组、切片等）中的元素类型不一定支持直接判等

最后实现了redis数组下标的转换，因为在redis种是从末尾开始细数下标的，也就是-1、-2、-3...和Go我们实现的不一样，所以要分别对start、end我们进行下标转换，换成正数，同时在实现的过程中要注意转换可能产生的越界等问题详情可见代码

```Go
func ConvertRange(start int64, end int64, size int64) (int, int) {
    if start < -size {
       return -1, -1
    } else if start < 0 {
       start = size + start
    } else if start >= size {
       return -1, -1
    }
    if end < -size {
       return -1, -1
    } else if end < 0 {
       end = size + end + 1
    } else if end < size {
       end = end + 1
    } else {
       end = size
    }
    if start > end {
       return -1, -1
    }
    return int(start), int(end)
}
```

## parser.go

这是RESP协议的核心内容了，说一下核心流程。

首先定义了协议回复的结构体

```Go
type Payload struct {
    Date redis.Reply
    Err  error
}
```

包含两部分内容Date与Err，Date就是在interface/reply种实现的接口类型，我们在redis/protocol/reply.go中实现了各种类型的协议指令回复，在下面会具体分析，Err就是自定义报错，没用到我们自己实现的logger中的error

其次是对数据的处理，ParseStream函数

```Go
func ParseStream(reader io.Reader) <-chan *Payload {
    ch := make(chan *Payload)
    go parse0(reader, ch)
    return ch
}
```

可以看到，我们获取的数据是io.Reader，一次处理一组读取，所以我们建了个默认1大小的管道，调用我们的核心函数parse0，这个函数放在最后来说，因为他的实现包含了其他几块内容，最后返回我们上述定义的只写管道结构体

再是ParseBytes函数，阅读代码可以很容易发现是将[]byte数据全部进行协议处理

与之对应的ParseOne函数，是解析第一个协议回复，这两个函数都是用来测试parser功能的

parseBulkString函数用来处理协议中的字符串回复，要注意，包含两行，所以第一行先处理$xxxx长度从第二个字节开始处理转换成64位10进制，代码处理了$-1的无对应指令回复（实现与consts.go），之后读取下一行的内容+2长度（包含\r\n），使用了readfull就怕出现4\r\n564 吧中间的\r\n忽略掉，因为是二进制安全所以这种情况会出现必须要处理，然后构建字符串协议回复，代码如下：

```Go
func parseBulkString(header []byte, reader *bufio.Reader, ch chan<- *Payload) error {
    strLen, err := strconv.ParseInt(string(header[1:]), 10, 64)
    if err != nil || strLen < -1 {
       protocolError(ch, "illegal bulk string header: "+string(header))
       return nil
    } else if strLen == -1 {
       ch <- &Payload{Date: protocol.MakeNullBulkReply()}
       return nil
    }
    body := make([]byte, strLen+2)
    _, err = io.ReadFull(reader, body)
    if err != nil {
       return err
    }
    ch <- &Payload{Date: protocol.MakeBulkReply(body[:len(body)-2])}
    return nil
}
```

parseRDBBulkString函数 ，RDB持久化字符串处理 RDB和AOF持久化不包含CRLF 所以需要重写一个处理函数，在对第二行内容处理的时候就不需要len+2 加上\r\n了，因为格式中不再包含

parseArray 数组处理，阅读起来并不困难，读数组长度，异常处理，里面内部两行一循环处理，内部都是$字符串处理的方式

protocolError ，最开始的结构体中的Err报错处理

然后来到核心函数parse0

因为每次调用parse0都是go parse0采用并发，所以内部首先是先defer，做好并发的结束处理。然后直接for循环，因为io.reader读取的协议数量是未知的，一直处理到彻底结束，然后一如正常的第一行异常处理，然后分为五种类型+、这些进行处理：

首先是+的简单命令处理，正常MakeStatusReply，如果前缀是FULLRESYNC，说明要生成RDB持久化，所以就用之前实现的parseRDBBulkString来处理，这是不同之处，

再是-的处理，错误的处理直接MakeErrReply

然后是：的处理，整数的处理MakeIntReply

接着是$字符串的处理，parseBulkString

最后是*数组处理，parseArray

如果都不是，则内部' '空的问题，按' '空分割为多行，然后按照数组的MakeMultiBulkReply多组字符串进行处理

整体考虑是相当周全的