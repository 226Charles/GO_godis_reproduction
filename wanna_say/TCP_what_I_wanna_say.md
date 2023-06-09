TCP部分主要是包含echo和server两块部分，还额外包含了一部分基础组件

今天在复现的过程中，学到了很多以前没考虑或者没学过的知识，从大到小说一下吧

# 1、就是TCP服务中最基本需要包含哪些部分

服务器，客户端：服务器需要处理很多事情，和客户端建立连接、处理客户端相应的请求、服务器做出相应的处理、对异常事件的处理，客户端连接的关闭。这些事件都是需要不断监听处理的，也就是需要在堵塞条件下进行，不然很容易出现错乱。

# 2、做工程需要注意的问题

这次做的还是复现，还不是从头到尾进行实现的梳理，所以在实现的时候是想到哪写到哪，还没有建立起完整的框架，可见一个好的项目总管对整个项目梳理清楚分工明确的重要性，而不是我以为的实现哥echo，server就好，里面还包含了他包含的log部分、wait超时处理部分、自定义bool原子操作部分。所以在发现这个问题后我先梳理了一下导入的包，实现了相应的基础的包的内容，再来处理TCP相关的代码。

# 3、代码的一部分问题

很有意思一块，以前从来没注意过，之前我都是用c++或者java做本科的一些课设项目，对c++相对来说熟悉一些，一般都是class类里面，定义成员变量以及相应的函数，go语言有趣的地方在他这些相应结构的函数是在外部定义的 ，如下面的echo部分里面的服务器处理的Close函数，他在前面增加了(h *EchoHandler)，是只这个函数是针对先前创造的结构体进行相应的处理，也是新了解到的一部分内容。

```Go
type EchoHandler struct {
    activeConn sync.Map
    closing    atomic.Boolean
}
// 关闭所有客户端连接
func (h *EchoHandler) Close() error {
    logger.Info("handler shutting down...")
    h.closing.Set(true)
    h.activeConn.Range(func(key interface{}, value interface{}) bool {
       client := key.(*EchoClient)
       _ = client.Close()
       return true
    })
    return nil
}
```

# 4、下面开始分别介绍一下今天工作相关的代码

## 先介绍sync包

wait.go

其实GO内置的Wait这些组件已经实现了基本满足的阻塞相关控制，但是考虑到tcp长期挂着是相当占用资源的，所以要考虑到超时等待这一点，wait.go就是实现了超时等待的功能，其中的Add、Done、Wait都是直接使用了sync包下实现的功能，WaitWithTimeout函数通过设置大小为1的通道，进行正常的阻塞等待，核心部分就是select模块，如果c通道有数据了，说明等待结束进流程了，没超时，如果time.After(timeout)超过了设置时间，那就判断超时，很简单的逻辑和实现

bool.go

这个是利用了go中的原子操作，指在多线程或分布式环境下的操作，能够保证其操作的原子性，即在执行时不会被中断或干扰，保证操作的完整性和一致性，sync/atomic实现了，但是没有实现布尔类型的，这里就自己实现了布尔类型的原子操作，底层原理就是uint32的1or0，也很基础。

## 再介绍下logger包

files.go

最终目的就是文件的打开读入操作，但是这里面倒是让我狠狠的上了一下工程上面的一课，对于异常操作处理我还是没有把他时刻放在心上，里面进行了一些对异常处理的操作，在工程项目中应该是常见或者必须的，培养一下自己的理念，有这些函数：检查文件存在性、检查权限、递归mkdir（检查目录是否存在）、以及读操作，读操作就先检查目录权限、再检查目标存在性、在进行相应的读，代码不难，但是项目理解很重要。

logger.go

我还是第一次详细了解log日志的相关概念，以前只是懂个皮毛，现在？比皮毛稍微多一点点，其实也不是很懂，作者在这一部分是为每一个log输出的设置了相应的前缀，按照级别、路径、内容进行生成，在初始的设置中，按照名称、时间以及ext格式设置（ext设置我还从来没考虑过，还是项目做少了），学到了这种设置方式，里面也用到了之前files实现的东西，所以这工程是个紧密相关，层层递进又相连的过程，最开始的规划很重要，要把底层的东西、较为原子的操作功能先实现了。

## 然后是接口部分

暂时就写了handler的两个接口，可能是因为外部暴露的原因吧，所以写个接口，其实具体原因还不明白。

## 再是TCP的核心部分

echo_old.go

哈哈为什么加个old，因为是原作者写的，因为作者实现的tcp呢比较基础简单，以后计划对redis的tcp了解更深入以后写个更复杂的处理。

这一部分主要是两块，服务器和客户端连接、服务器的处理。首先利用还是那个面实现的wait超时处理了一下客户端连接超时，写了个客户端连接Close函数，然后就是这一部分的核心内容，服务器处理的Handle函数，也就是上面的接口，首先对closing，利用实现的原子布尔，判一下关不关连接（人家都不会上来先写业务，先处理和避免相关的异常，要考虑全面），然后新建个客户端连接，map里面存进去，然后读数据写数据，这个map是sync.Map，再并发里很好用，切记写数据时候加个阻塞，最后就是服务器处理端的大Close函数，给原子boolean设真，然后便利map里面所有客户端连接全关了。

server_old

是TCP的服务端，建了个TCP属性的结构，包含：地址、最大连接数和时间。有两块函数，先说子函数ListenAndServeWithSignal，建一下服务器关闭管道和中断信号管道，有中断就传给关闭管道关闭，完成后执行ListenAndServe函数，然后listener.Accept()不断监听连接请求，一直阻塞直到有新的连接请求到来，如果err就写到他设置的err管道内，成功就计数+1，阻塞，直到处理完成。当然对closechan和errch并发处理关闭了监听和处理，有意思的是这个errCh，我一般是以为err出现了直接log或者处理，但是想到TCP是不断建立连接的，不是单对一个用户的，这个有问题了还得层层关闭，不然数据丢失或者，关闭连接listener之后再层层关闭，这也是做TCP服务需要考虑到的。和作者说的一样很优雅

# 5、拆包与粘包问题

这个问题很有意思，我直接复制原作者的话吧。

[Golang 实现 Redis(1): Golang 编写 Tcp 服务器 - -Finley- - 博客园](https://www.cnblogs.com/Finley/p/11070669.html)这是链接

在上文的 Echo 服务器示例中我们用`\n`表示消息结束，从 read 函数读取的数据可能存在下列几种情况:

1. 收到两段数据: "abc", "def\n" 它们属于一条消息 "abcdef\n" 这是拆包的情况
2. 收到一段数据: "abc\ndef\n" 它们属于两条消息 "abc\n", "def\n" 这是粘包的情况

应用层协议通常采用下列几种思路之一来定义消息，以保证完整地进行读取:

- 定长消息
- 在消息尾部添加特殊分隔符，如示例中的Echo协议和FTP控制协议。bufio 标准库会缓存收到的数据直到遇到分隔符才会返回，它可以帮助我们正确地分割字节流。
- 将消息分为 header 和 body, 并在 header 中提供 body 总长度，这种分包方式被称为 LTV(length，type，value) 包。这是应用最广泛的策略，如HTTP协议。当从 header 中获得 body 长度后, io.ReadFull 函数会读取指定长度字节流，从而解析应用层消息。

# 6、最后

管道管道管道！异常异常异常！这是今天最直接的感悟。