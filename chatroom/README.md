# 4.1 基于 TCP 的聊天室

本节通过命令行来模拟基于 TCP 的简单聊天室。

本程序可以将用户发送的文本消息广播给该聊天室内的所有其他用户。该服务端程序中有四种 goroutine：main goroutine 和 广播消息的 goroutine，以及每一个客户端连接都会有一对读和写的 goroutine。

先在本地创建一个项目（若为 Windows 系统，可根据实际情况自行调整项目的路径）：

```bash
$ mkdir -p $HOME/go-programming-tour-book/chatroom
$ cd $HOME/go-programming-tour-book/chatroom
$ go mod init github.com/go-programming-tour-book/chatroom
$ mkdir -p cmd/tcp
```

## 4.1.1 一步步代码实现

在项目 cmd/tcp 目录下创建文件：server.go，入口 main 函数的代码如下：

```go
func main() {
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

// broadcaster 用于记录聊天室用户，并进行消息广播：
// 1. 新用户进来；2. 用户普通消息；3. 用户离开
func broadcaster() {
}
```

以上代码基本上是一个 TCP Server 的通用代码，这就是 Go 语言进行 Socket 编程的框架，是不是超级简单？！

> 注意，在 `listen` 时没有指定 IP，表示绑定到当前机器的所有 IP 上。根据具体情况可以限制绑定具体的 IP，比如只绑定在 127.0.0.1 上：net.Listen(“tcp”, “127.0.0.1:2020”)

代码中 `go broadcaster` 这句用于广播消息，后文会实现、讲解。

然后看 handleConn 函数的实现：

```go
func handleConn(conn net.Conn) {
	defer conn.Close()

	// 1. 新用户进来，构建该用户的实例
	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}

	// 2. 当前在一个新的 goroutine 中，用来进行读操作，因此需要开一个 goroutine 用于写操作
	// 读写 goroutine 之间可以通过 channel 进行通信
	go sendMessage(conn, user.MessageChannel)

	// 3. 给当前用户发送欢迎信息；给所有用户告知新用户到来
	user.MessageChannel <- "Welcome, " + user.String()
	messageChannel <- "user:`" + strconv.Itoa(user.ID) + "` has enter"

	// 4. 将该记录到全局的用户列表中，避免用锁
	enteringChannel <- user

	// 5. 循环读取用户的输入
	input := bufio.NewScanner(conn)
	for input.Scan() {
			messageChannel <- strconv.Itoa(user.ID) + ":" + input.Text()
	}
  
  if err := input.Err(); err != nil {
		log.Println("读取错误：", err)
	}

	// 6. 用户离开
	leavingChannel <- user
	messageChannel <- "user:`" + strconv.Itoa(user.ID) + "` has left"
}
```

以上代码有详细的注释说明。这里说下思路：

1）新用户到来，生成一个 User 的实例，代表该用户。User 结构体声明如下：

```go
type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan string
}
```

- ID 是用户唯一标识，通过 `GenUserID` 函数生成；
- Addr 是用户的 IP 地址和端口；
- EnterAt 是用户进入时间；
- MessageChannel 是当前用户发送消息的通道；

2）新开一个 goroutine 用于给用户发送消息：

```go
func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
```

结合 User 结构的 MessageChannel，很容易知道，需要给某个用户发送消息，只需要往该用户的 MessageChannel 中写入消息即可。这里需要特别提醒下，因为 sendMessage 在一个新 goroutine 中，如果函数里的 `ch` 不关闭，该 goroutine 是不会退出的，因此需要注意不关闭 ch 导致的 goroutine 泄露问题。

**这里有一个语法，有些人可能没见过：ch <-chan string。简单介绍下它的含义和作用：**

> channel 实际上有三种类型，大部分时候，我们只用了其中一种，就是正常的既能发送也能接收的 channel。除此之外还有单向的 channel：只能接收（<-chan，only receive）和只能发送（chan<-， only send）。它们没法直接创建，而是通过正常（双向）channel 转换而来（会自动隐式转换）。它们存在的价值，主要是避免 channel 被乱用。上面代码中 ch <-chan string 就是为了限制在 sendMessage 函数中只从 channel 读数据，不允许往里写数据。

3）给当前用户发送欢迎信息，同时给聊天室所有用户发送有新用户到来的提醒；

4）将该新用户写入全局用户列表，也就是聊天室用户列表，这里通过 channel 来写入，避免了锁。注意，这里和 3）的顺序不能反，否则自己会收到自己到来的消息提醒；（当然，我们也可以做消息过滤处理）

5）读取用户的输入，并将用户信息发送给其他用户。这里简单介绍一下 bufio 包的 Scanner：

> 在 bufio 包中有多种方式获取文本输入，ReadBytes、ReadString 和独特的 ReadLine，对于简单的目的这些都有些过于复杂了。在 Go 1.1 中，添加了一个新类型，Scanner，以便更容易的处理如按行读取输入序列或空格分隔单词等这类简单的任务。它终结了如输入一个很长的有问题的行这样的输入错误，并且提供了简单的默认行为：基于行的输入，每行都剔除分隔标识。

6）用户离开，需要做登记，并给聊天室其他用户发通知；

接下来我们实现 broadcaster 函数，该方法的主要用于记录聊天室用户，并进行消息广播，代码如下：

```go
// broadcaster 用于记录聊天室用户，并进行消息广播：
// 1. 新用户进来；2. 用户普通消息；3. 用户离开
func broadcaster() {
	users := make(map[*User]struct{})

	for {
		select {
		case user := <-enteringChannel:
			// 新用户进入
			users[user] = struct{}{}
		case user := <-leavingChannel:
			// 用户离开
			delete(users, user)
			// 避免 goroutine 泄露
			close(user.MessageChannel)
		case msg := <-messageChannel:
			// 给所有在线用户发送消息
			for user := range users {
				user.MessageChannel <- msg
			}
		}
	}
}
```

这里关键有 3 点：

- 负责登记/注销用户，通过 map 存储在线用户；
- 用户登记、注销，使用专门的 channel。在注销时，除了从 map 中删除用户，还将 user 的 MessageChannel 关闭，避免上文提到的 goroutine 泄露问题；
- 全局的 messageChannel 用来给聊天室所有用户广播消息；

可见 broadcaster 函数关键在于 goroutine 和 channel 的运用，很好的践行了 Go 的理念：通过通信来共享内存。它里面三个 channel 的定义如下：

```go
var (
	// 新用户到来，通过该 channel 进行登记
	enteringChannel = make(chan *User)
	// 用户离开，通过该 channel 进行登记
	leavingChannel = make(chan *User)
	// 广播专用的用户普通消息 channel，缓冲是尽可能避免出现异常情况堵塞，这里简单给了 8，具体值根据情况调整
	messageChannel = make(chan string, 8)
)
```

## 4.1.2 简单客户端

客户端的实现直接采用 《The Go Programming Language》一书对应的示例源码：ch8/netcat3/netcat.go 。我们将代码拷贝放入项目的 cmd/tcp/client.go 文件中，代码如下：

```go
func main() {
	conn, err := net.Dial("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn) // NOTE: ignoring errors
		log.Println("done")
		done <- struct{}{} // signal the main goroutine
	}()

	mustCopy(conn, os.Stdin)
	conn.Close()
	<-done
}

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}
```

- 新开了一个 goroutine 用于接收消息；
- 通过 io.Copy 来操作 IO，包括从标准输入读取数据写入 TCP 连接中，以及从 TCP 连接中读取数据写入标准输出；
- 新开的 goroutine 通过一个 channel 来和 main goroutine 通讯；

## 4.1.3 演示

在终端启动服务端：

```bash
$ cd $HOME/go-programming-tour-book/chatroom
$ go run cmd/tcp/server.go
```

在若干终端启动客户端：（这里启动 3 个）

```bash
$ go run cmd/tcp/client.go
Welcome, 127.0.0.1:49777, UID:1, Enter At:2020-01-31 16:15:24+8000
user:`2` has enter
user:`3` has enter

$ go run cmd/tcp/client.go
Welcome, 127.0.0.1:49781, UID:2, Enter At:2020-01-31 16:15:35+8000
user:`3` has enter

$ go run cmd/tcp/client.go
Welcome, 127.0.0.1:49784, UID:3, Enter At:2020-01-31 16:15:44+8000
```

接着，在第一个客户端输入：`hello, I am first user`，发现大家都收到了该消息。

## 4.1.4 改进

以上的聊天室比较简单、粗糙，存在几个比较明显的问题：

- 自己发的消息自己收到了；
- 客户端长时间没有发送任何消息（例如超过 5 分钟），应该自动将其踢出聊天室；

那么怎么解决呢？

要过滤自己的消息，可以在发送消息时加上发送者。我们可以定义一个 Message 类型：

```go
// 给用户发送的消息
type Message struct {
	OwnerID int
	Content string
}
```

将之前使用普通文本的地方全部替换为 Message 类型，之后将 broadcaster 函数中相关代码修改如下：

```go
case msg := <-messageChannel:
  // 给所有在线用户发送消息
  for user := range users {
    if user.ID == msg.OwnerID {
      continue
    }
    user.MessageChannel <- msg.Content
  }
}
```

这样就达到了过滤自己发送的消息的目的。

至于自动踢出未活跃用户，在 handleConn 函数第 4 步之后新增如下代码：

```go
// 控制超时用户踢出
var userActive = make(chan struct{})
go func() {
  d := 5 * time.Minute
  timer := time.NewTimer(d)
  for {
    select {
    case <-timer.C:
      conn.Close()
    case <-userActive:
      timer.Reset(d)
    }
  }
}()
```

同时，每次接收到用户的消息后，往 userActive 中写入消息，继续向 handleConn 函数的第 5 步中 for 循环的最后增加 `userActive <- struct{}{}` 这行代码：

```go
// 5. 循环读取用户的输入
input := bufio.NewScanner(conn)
for input.Scan() {
  msg.Content = strconv.Itoa(user.ID) + ":" + input.Text()
  messageChannel <- msg

  // 用户活跃
  userActive <- struct{}{}
}
```

如果用户 5 分钟内未发送任何消息，服务端会将连接断开，这样 input.Scan 会返回 false，handleConn 结束，客户退出。

## 4.1.5 小结

基于 TCP 的聊天室已经完成，这是一个简单的开始，基于这个简单的模型，可以演进出强大的系统。

**思考题：**

如果有大量用户，广播消息是否会存在延迟，导致消息堵塞？

# 4.2 WebSocket 介绍、握手协议和细节

基于 WebSocket 的聊天室是本章的重点。先一起认识下 WebSocket。

## 4.2.1 WebSocket 介绍

来自维基百科的解释：

> WebSocket 是一种网络传输协议，可在单个 TCP 连接上进行全双工通信，位于 OSI 模型的应用层。 WebSocket 协议在 2011 年由 IETF 标准化为 RFC 6455，后由 RFC 7936 补充规范。Web IDL 中的 WebSocket API 由 W3C 标准化。
>
> WebSocket 使得客户端和服务器之间的数据交换变得更加简单，允许服务端主动向客户端推送数据。在 WebSocket API 中，浏览器和服务器只需要完成一次握手，两者之间就可以创建持久性的连接，并进行双向数据传输。

WebSocket 是一种与 HTTP 不同的协议。两者都位于 OSI 模型的应用层，并且都依赖于传输层的 TCP 协议。虽然它们不同，但 RFC 6455 规定：“WebSocket 设计为通过 80 和 443 端口工作，以及支持 HTTP 代理和中介”，从而使其与 HTTP 协议兼容。 为了实现兼容性，WebSocket 握手使用 HTTP Upgrade 头从 HTTP 协议更改为 WebSocket 协议。

WebSocket 协议支持 Web 浏览器（或其他客户端应用程序）与 Web 服务器之间的交互，具有较低的开销，便于实现客户端与服务器的实时数据传输。 服务器可以通过标准化的方式来实现，而无需客户端首先请求内容，并允许消息在保持连接打开的同时来回传递。通过这种方式，可以在客户端和服务器之间进行双向持续交互。 通信默认通过 TCP 端口 80 或 443 完成。

大多数浏览器都支持该协议，包括 Google Chrome、Firefox、Safari、Microsoft Edge、Internet Explorer 和 Opera。

与 HTTP 不同，WebSocket 提供全双工通信。此外，WebSocket 还可以在 TCP 之上启用消息流。TCP 单独处理字节流，没有固有的消息概念。 在 WebSocket 之前，使用 Comet 可以实现全双工通信。但是 Comet 存在 TCP 握手和 HTTP 头的开销，因此对于小消息来说效率很低。WebSocket 协议旨在解决这些问题。

WebSocket 协议规范将 ws（WebSocket）和 wss（WebSocket Secure）定义为两个新的统一资源标识符（URI）方案，分别对应明文和加密连接。除了方案名称和片段 ID（不支持#）之外，其余的 URI 组件都被定义为此 URI 的通用语法。

## 4.2.2 WebSocket 的优点

了解了 WebSocket 是什么，那 WebSocket 有哪些优点？这里总结如下：

- 较少的控制开销。在连接创建后，服务器和客户端之间交换数据时，用于协议控制的数据包头部相对较小。在不包含扩展的情况下，对于服务器到客户端的内容，此头部大小只有 2 至 10 字节（和数据包长度有关）；对于客户端到服务器的内容，此头部还需要加上额外的 4 字节的掩码。相对于 HTTP 请求每次都要携带完整的头部，此项开销显著减少了。
- 更强的实时性。由于协议是全双工的，所以服务器可以随时主动给客户端下发数据。相对于 HTTP 请求需要等待客户端发起请求服务端才能响应，延迟明显更少；即使是和 Comet 等类似的长轮询比较，其也能在短时间内更多次地传递数据。
- 保持连接状态。Websocket 是一种有状态的协议，通信前需要先创建连接，之后的通信就可以省略部分状态信息了。而 HTTP 请求可能需要在每个请求都携带状态信息（如身份认证等）。
- 更好的二进制支持。Websocket 定义了二进制帧，相对 HTTP，可以更轻松地处理二进制内容。
- 可以支持扩展。Websocket 定义了扩展，用户可以扩展协议、实现部分自定义的子协议。如部分浏览器支持压缩等。
- 更好的压缩效果。相对于 HTTP 压缩，Websocket 在适当的扩展支持下，可以沿用之前内容的上下文，在传递类似的数据时，可以显著地提高压缩率。

## 4.2.3 选择一个合适的库

从前面介绍可知，WebSocket 是独立的、创建在 TCP 上的协议。Websocket 通过 HTTP/1.1 协议的 101 状态码进行握手。

为了创建 Websocket 连接，需要通过浏览器发出请求，之后服务器进行回应，这个过程通常称为“握手”（handshaking）。

本节我们自己实现一个 WebSocket 服务端和客户端来学习下这个握手协议。

对于 Web 客户端（浏览器），直接使用 HTML 5 的 WebSocket API 即可。对于服务端，我们肯定需要选择一个 WebSocket 库。一般来说，一种技术对应的库，很少只有一个，在众多库中怎么选择？

首先，我们需要找到目标库。打开 pkg.go.dev，输入 websocket 搜索：

![img](https://golang2.eddycjy.com/images/ch4/pkg-search.png)

图中前 3 个中，大家对于 gorilla 应该挺熟悉，有挺多优秀 Go 库，这个 WebSocket 的库也是最早的。而第 2 个库很明显是官方维护的（域名 golang.org/x），但点开这个库，发现它说 gorilla/websocket 和 nhooyr.io/websocket 这两个库实现的更全，维护也更活跃。（为了简便，后文称 gorilla/websocket 为 gorilla，nhooyr.io/websocket 为 nhooyr）

![image](https://golang2.eddycjy.com/images/ch4/websocket-doc.png)

我们着重看看 nhooyr 库，它是 2019 年初实现的，它的文档有这几个库的对比。

> nhooyr 作者指出，gorilla/websocket 和 gobwas/ws 在正确实现 WebSocket 协议方面都非常有用，这要归功于它们的作者。特别是，他查看了 gorilla/websocket 的 issue 跟踪，以确保理解了实现的细节和人们如何在生产环境中使用 WebSocket。

### 与 gorilla/websocket 比较

该库已有 6 年的历史。因此，与 nhooyr 相比，它被广泛使用并且非常成熟。

通过对比 gorilla 和 nhooyr 发现，nhooyr 的 API 只提供一种处理方式，这使得它很容使用，它的 API 不仅更简单，而且实现只有 2200 行代码，而 gorilla 有 3500 行。更多的代码意味着要更多的维护，更多的测试、文档以及更多的漏洞。

而且，nhooyr 支持更新的 Go 习惯用法，例如 context.Context。它还将 net/http 的 Client 和 ResponseWriter 直接用于 WebSocket 握手。

nhooyr 的其他一些优点是，它支持并发写入，并且非常容易在关闭连接时提供状态码和原因。实际上，它甚至为你实现了完整的 WebSocket 关闭方法。

nhooyr 的 ping API 也更好。gorilla 需要在 Conn 上注册一个 pong 处理程序，这会导致笨拙的控制流。通过 nhooyr，您可以在 Conn 上使用 ping 方法，该方法发送 ping 并等待 pong。

此外，nhooyr 可以针对浏览器编译为 Wasm。

在性能方面，差异主要取决于你的应用程序代码。如果你使用 wsjson 和 wspb 子包，则 nhooyr 可以立即使用消息缓冲区。如上所述，nhooyr 还支持并发写。

仅使用并发安全 Go 时，nhooyr 使用的 WebSocket 掩码算法也比 gorilla 或 gobwas/ws 快 1.75 倍。

nhooyr 的唯一性能缺陷是它使用了一个额外的 goroutine 来支持 context.Context 的取消。这将花费 2 KB 的内存，与这些优点相比很廉价。

### 与 golang.org/x/net/websocket 比较

该库基本不维护，而且 API 也没有很好的反映出 WebSocket 的语义，建议永远别用。

### 与 gobwas/ws 的比较

该库具有非常灵活的 API，但这是以可用性和清晰性为代价的。

该库在性能方面很棒，作者为确保其速度付出了巨大的努力。nhooyr 库的作者已将尽可能多的优化应用到了 nhooyr.io/websocket 中。

如果你想要一个可以完全控制所有内容的库，那么可以使用该库。但是对于 99.9％ 场景，nhooyr 会更适合。它有几乎相近的性能，但更易于使用。

### 小结

通过上面的介绍，选择哪个库，大家心里应该有数了。在后面实现聊天室时，我们将用 nhooyr 这个库，它更易使用、性能更好，也更符合 Go 的风格。在本节最后会附上这个小节试验 Demo 对应的 gorilla 版本。

## 4.2.4 nhooyr.io/websocket 的介绍和使用

该库的核心特色：

- 很小、符合 Go 习惯的 API
- 核心代码 2200 行
- context.Context 支持
- 全面测试覆盖
- 不依赖任何第三方库
- 支持 JSON 和 ProtoBuf
- 默认情况就具有高性能
- 开箱即用的并发支持
- 全面的 Wasm 支持
- 完整的 Close handshake 支持

Conn，Dial 和 Accept 是此包的主要入口点。使用 Dial 连接 WebSocket 服务器，Accept 接受 WebSocket 客户端连接请求，然后使用 Conn 与生成的 WebSocket 连接进行交互。

### 服务端实现

在 chatroom 目录下执行如下操作：

```
$ mkdir -p cmd/websocket
```

在 cmd/websocket 目录下创建 server.go 文件。为了验证握手过程，服务端简单实现如下：

```go
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w,"HTTP, Hello")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		conn, err := websocket.Accept(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close(websocket.StatusInternalError, "内部出错了！")

		ctx, cancel := context.WithTimeout(req.Context(), time.Second*10)
		defer cancel()

		var v interface{}
		err = wsjson.Read(ctx, conn, &v)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("接收到客户端：%v\n", v)

		err = wsjson.Write(ctx, conn, "Hello WebSocket Client")
		if err != nil {
			log.Println(err)
			return
		}

		conn.Close(websocket.StatusNormalClosure, "")
	})

	log.Fatal(http.ListenAndServe(":2021", nil))
}
```

在同一个端口（2021），提供 HTTP 和 WebSocket 两种协议的服务。对端点 `/` 的请求，走 HTTP；对端点 `/ws` 的请求走 WebSocket。

这段代码，后面会专门介绍，这里主要学习握手过程。

### 客户端实现

nhooyr 库提供了简便的方式实现一个 WebSocket 客户端（cmd/websocket/client.go）。

```go
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://localhost:2021/ws", nil)
	if err != nil {
		panic(err)
	}
	defer c.Close(websocket.StatusInternalError, "内部错误！")

	err = wsjson.Write(ctx, c, "Hello WebSocket Server")
	if err != nil {
		panic(err)
	}

	var v interface{}
	err = wsjson.Read(ctx, c, &v)
	if err != nil {
		panic(err)
	}
	fmt.Printf("接收到服务端响应：%v\n", v)

	c.Close(websocket.StatusNormalClosure, "")
}
```

## 4.2.5 抓包分析协议

执行如下命令启动服务端：

```bash
$ cd $HOME/go-programming-tour-book/chatroom
$ go run cmd/websocket/server.go
```

打开 Chrome 浏览器，访问 http://localhost:2021 ，通过浏览器的开发者工具抓包如下：

![image](https://golang2.eddycjy.com/images/ch4/http.png)

接着启动 WebSocket 客户端，同时开启 Wireshark 抓包工具（如果没有，请安装），并监控 loopback(lo0) 网络接口。

```bash
$ go run cmd/websocket/client.go
```

服务端和客户端都能按照预期打印 Hello 字样。这里我们主要关注抓包情况。

![image](https://golang2.eddycjy.com/images/ch4/wireshark.png)

TCP 的三次握手和四次挥手很容易找到。我们选中序号是 5 的这条：GET /ws HTTP/1.1。

![image](https://golang2.eddycjy.com/images/ch4/websocket-request.png)

请求头和上面访问 http://localhost:2021 的请求头对比，发现有如下不同（主要关注 WebSocket 相关的点）：

| 请求头                | HTTP       | WebSocket                | WebSocket 值说明                                             |
| --------------------- | ---------- | ------------------------ | ------------------------------------------------------------ |
| Connection            | keep-alive | Upgrade                  | Connection 必须设置为 Upgrade，表示客户端希望连接升级。      |
| Upgrade               | -          | websocket                | Upgrade 字段必须设置 websocket，表示希望升级到 Websocket 协议。 |
| Sec-Websocket-Key     | -          | cNV8eLOBYxq9MQir9FjCgw== | 随机的字符串，服务器端会用这些数据来构造出一个 SHA-1 的信息摘要。把 “Sec-WebSocket-Key” 加上一个特殊字符串 “258EAFA5-E914-47DA-95CA-C5AB0DC85B11”，然后计算 SHA-1 摘要，之后进行 Base64 编码，将结果做为 “Sec-WebSocket-Accept” 头的值，返回给客户端。如此操作，可以尽量避免普通 HTTP 请求被误认为 Websocket 协议。 |
| Sec-Websocket-Version | -          | 13                       | 表示支持的 Websocket 版本。RFC6455 要求使用的版本是 13，之前草案的版本均应当弃用。 |

此外，如果浏览器中发起 WebSocket 请求，可能会有可选的 Origin 头，用来表示在浏览器中发起此 Websocket 连接所在的页面，类似于 Referer。但是，与 Referer 不同的是，Origin 只包含了协议和主机名称。其他的 HTTP 头，如 Cookie 等，也可以用于 WebSocket。

接下来看看针对该请求的响应。选中序号 7 的这条：HTTP/1.1 101 Switching Protocols 。

![image](https://golang2.eddycjy.com/images/ch4/websocket-response.png)

响应头和上面访问 http://localhost:2021 的响应头对比，发现有如下不同（主要关注 WebSocket 相关的点）：其中响应行 HTTP 是 200，而 WebSocket 是 101，表示切换协议。

| 响应头                 | HTTP | WebSocket                    | WebSocket 值说明                                |
| ---------------------- | ---- | ---------------------------- | ----------------------------------------------- |
| Connection             | -    | Upgrade                      | 和请求头一致                                    |
| Upgrade                | -    | websocket                    | 和请求头一致                                    |
| Sec-WebSocket-Accept   | -    | 0Nkd0hCFLtHwsRyX/mmlM7ulNGI= | 计算出来的，见请求头 Sec-Websocket-Key 的说明   |
| Sec-WebSocket-Location | -    | 可选                         | 与 Host 字段对应，表示请求 WebSocket 协议的地址 |

以上就是 WebSocket 协议的握手过程，之后就是实际的数据传输。这也验证了 Websocket 协议本质上是一个基于 TCP 的协议。建立连接需要握手，客户端首先向服务器发起一条特殊的 http 请求，服务器解析后生成应答到客户端，这样子一个 websocket 连接就建立了，直到某一方关闭连接。

读者可以接着看抓包中的四个 Protocol 为 WebSocket 的条目：它们分别是客户端发送消息、服务端发送消息、服务端关闭连接和客户端关闭连接。（这里提示一点，客户端通过 WebSocket 发送的数据会进行掩码处理，见抓包中的 MASKED 标记）

## 4.2.6 小结

本节简单介绍了什么是 WebSocket，有哪些优点，对比了 Go 中几个 WebSocket 库，最后提供一个简单的 WebSocket 客户端和服务端实现，用 Wireshark 抓包分析 WebSocket 协议。

通过本节的学习，希望读者掌握 nhooyr.io/websocket 包的使用。下一节起，我们会使用该包来构建一个 WebSocket 聊天室。

附 gorilla/websocket 对应的服务端代码：

```go
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "HTTP, Hello")
	})

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		// 就做一次读写
		var v interface{}
		err = conn.ReadJSON(&v)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("接收到客户端：%s\n", v)

		if err := conn.WriteJSON("Hello WebSocket Client"); err != nil {
			log.Println(err)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":2021", nil))
}
```

# 4.3 聊天室需求分析和设计

聊天室是一种人们可以在线交谈的网络论坛，在同一聊天室的人们通过广播消息进行实时交谈。现在几乎所有的社交类的网站，都或多或少有聊天室的影子，特别突出的就是直播了。

## 4.3.1 聊天室主要需求

一般来说，聊天室的功能主要有：

- 聊天室在线人数、在线列表
- 昵称、头像设置
- 发送/接收消息
- 进入、退出聊天室
- 不同消息类型的支持：文字、表情、图片、语音、视频等
- 敏感词过滤
- 黑名单/禁言
- 聊天室信息
- 创建、修改聊天室
- 聊天室信息存储和历史消息查询
- 。。。

功能还可以有更多，读者可以放下本书，自己想想怎么实现。

因为篇幅原因，本书的聊天室不会包含上面所有内容，主要集中在基本功能的实现，掌握如何通过 Go 语言实现一个聊天室。

该聊天室的界面：

![img](https://golang2.eddycjy.com/images/ch4/chat-ui.png)

## 4.3.2 技术选择

聊天室有一个特点：页面不需要刷新，所有界面的变化是通过长连接读写数据实现的。这很适合使用 Vue 之类的前端 MVVM 框架来做客户端，但本书主要面向服务端 Go 开发人员，前端的内容相对会比较浅（但建议你还是需要学习一些前端的知识），作者也只是略懂一二，如果有不对之处还请指正。

服务端自然是使用 Go，WebSocket 的库使用 nhooyr.io/websocket，本周最后会介绍如何部署、配置环境，包括用 Nginx 代理 WebSocket 以及 HTTPS。

\##4.3.3 总体设计思路和流程

聊天室的交互流程一般是这样的：

用户通过浏览器访问首页，比如 http://localhost:20222 ，这时使用的是 HTTP 协议，服务端进行首页的 HTML 渲染，返回给用户浏览器，也就是聊天室页面。

浏览器加载完聊天室页面，用户输入自己的昵称，点击进入聊天室，发起 ws 协议请求，比如 ws://localhost:2022/ws 。按前面的讲解，这时走的依然是 HTTP 协议，进行 WebSocket 握手，协议切换，返回客户端 101 Switching Protocols。这里说明下，HTTP 协议得是 1.1 版本，1.0 版本会报错。

之后进入长链接的读写流程，分别在各自的 goroutine 进行。同时，广播消息在单独的 goroutine 中进行，它们之间通过 channel 进行通讯。流程示意图如下：

![image](https://golang2.eddycjy.com/images/ch4/chatroom-design.png)

> 注意：限于篇幅，本书的聊天室，聊天内容在服务端不存储，所以服务端没有引入存储服务。

现在让我们开始 WebSocket 聊天室之旅吧！

# 4.4 实现聊天室：项目组织和基础代码框架

一个项目，目录结构如何组织，各个语言似乎有自己的一套约定成俗的东西，比如了解 Java 的应该知道，Java Web 几乎是固定的目录组织方式。Go 语言经过这几年的发展，慢慢的也会有自己的一些目录结构组织方式。

## 4.4.1 聊天室项目的组织方式

本书的聊天室项目不复杂，所以项目的组织结构也比较简单，目录结构如下：（读者在本地创建类似的目录结构，方便跟着动手实现）

```
├── README.md
├── cmd
│   ├── chatroom
│       └── main.go
├── go.mod
├── go.sum
├── logic
│   ├── broadcast.go
│   ├── message.go
│   └── user.go
├── server
│   ├── handle.go
│   ├── home.go
│   └── websocket.go
└── template
    └── home.html
```

相关目录说明如下：

- cmd：该目录几乎是 Go 圈约定俗成的，Go 官方以及开源界推荐的方式，用于存放 main.main；
- logic：用于存放项目核心业务逻辑代码，和 service 目录是类似的作用；
- server：存放 server 相关代码，虽然这是 WebSocket 项目，但也可以看成是 Web 项目，因此可以理解成存放类似 controller 的代码；
- template：存放静态模板文件；

关于 main.main，即包含 main 包 和 main 函数的文件（一般是 main.go）放在哪里，目前一般有两种做法：

1）放在项目根目录下。这样放有一个好处，那就是可以方便的通过 go get 进行安装。比如 github.com/polaris1119/golangclub ，按这样的方式安装：

```bash
$ go get github.com/polaris1119/golangclub
```

成功后在 `$GOBIN`（未设置时取 `$GOPATH[0]/bin` ）目录下会找到 golangclub 可执行文件。但如果你的项目不止一个可执行文件，也就是会存在多个 main.go，这种方式显然没法满足需求。

2）创建一个 cmd 目录，专门放置 main.main，有些可能会直接将 main.go 放在 cmd 下，但这又回到了上面的方式，而且还没上面的方式方便。一般建议项目存在多个可执行文件时，在 cmd 下创建对应的目录。因为前面章节的需要，在项目 chatroom 中，cmd 下有了三个目录：tcp、websocket 和 chatroom。对于这种方式，通过 go get 可以这样安装：

```bash
$ go get -v github.com/go-programming-tour-book/chatroom/cmd/...
```

为了演示方便，我们的 tcp 和 websocket 同时包含了 server 和 client，相当于一个目录下有两个 main.main，所以用这种方式安装会报错，错误信息类似这样：

```bash
../../../../go/pkg/mod/github.com/go-programming-tour-book/chatroom@v0.0.0-20200412113309-9f22642e72e5/cmd/tcp/server.go:16:6: main redeclared in this block
	previous declaration at ../../../../go/pkg/mod/github.com/go-programming-tour-book/chatroom@v0.0.0-20200412113309-9f22642e72e5/cmd/tcp/client.go:13:6
# github.com/go-programming-tour-book/chatroom/cmd/websocket
	previous declaration at ../../../../go/pkg/mod/github.com/go-programming-tour-book/chatroom@v0.0.0-20200412113309-9f22642e72e5/cmd/websocket/client.go:12:6
```

所以，我们这个聊天室项目，可以用下面这种方式安装：

```bash
$ go get -v github.com/go-programming-tour-book/chatroom/cmd/chatroom
```

## 4.4.2 基础代码框架

接下来看看具体的代码实现。

1、main.go 的代码如下：

```go
var (
	addr   = ":2022"
	banner = `
    ____              _____
   |    |    |   /\     |
   |    |____|  /  \    | 
   |    |    | /----\   |
   |____|    |/      \  |

Go 语言编程之旅 —— 一起用 Go 做项目：ChatRoom，start on：%s
`
)

func main() {
	fmt.Printf(banner+"\n", addr)

	server.RegisterHandle()

	log.Fatal(http.ListenAndServe(addr, nil))
}
```

该项目直接使用标准库 net/http 来启动 HTTP 服务，Handle 的注册统一在 server 包中进行。

> 大家以后项目中，可以试试 banner 的打印，感觉挺酷的。

2、server.RegisterHandle

在 server/handle.go 中，加上如下代码：

```go
func RegisterHandle() {
	inferRootDir()

	// 广播消息处理
	go logic.Broadcaster.Start()

	http.HandleFunc("/", homeHandleFunc)
	http.HandleFunc("/ws", WebSocketHandleFunc)
}
```

该函数内的四行代码，前两行其实并非是 Handle 的注册。

一般来说，项目中会需要读文件，比如读模板文件、读配置文件、数据文件等。为了能够准确的找到文件所在路径，在程序中应该尽早推断出项目的根目录，之后读其他文件，通过该根目录拼接绝对路径读取。inferRootDir 函数就是负责推断出项目根目录。具体看看推断的逻辑：

```go
var rootDir string

// inferRootDir 推断出项目根目录
func inferRootDir() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var infer func(d string) string
	infer = func(d string) string {
    // 这里要确保项目根目录下存在 template 目录
		if exists(d + "/template") {
			return d
		}

		return infer(filepath.Dir(d))
	}

	rootDir = infer(cwd)
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
```

- 通过 os.Getwd() 获取当前工作目录；
- infer 被递归调用，判断目录 d 下面是否存在 template 目录（只要是项目根目录下存在的目录即可，并非一定是 template）；
- 如果 d 中不存在，则在其上级目录递归查找；

在项目中任意一个目录执行编译然后运行或直接 go run，该函数都能正确找到项目的根目录。

`go logic.Broadcaster.Start()` 启动一个 goroutine 进行广播消息的处理，具体内容后文再讲。

最后两行代码：

```go
http.HandleFunc("/", homeHandleFunc)
http.HandleFunc("/ws", WebSocketHandleFunc)
```

用于注册 “/” 和 “/ws” 两个路由，其中 “/” 代表首页，"/ws” 用来服务 WebSocket 长连接。

至此咱们代码的基础框架或者说项目启动涉及到的流程就基本完成了。下节将讲解聊天室的核心处理流程。

# 4.5 实现聊天室：核心流程

本节我们讲解聊天室的核心流程的实现。

## 4.5.1 前端关键代码

在项目中的 template/home.html 文件中增加 html 相关代码：（考虑篇幅，只保留主要的 html 部分，完整代码可通过 `git clone https://github.com/go-programming-tour-book/chatroom` 获取）

```html
<div class="container" id="app">
    <div class="row">
        <div class="col-md-12">
            <div class="page-header">
                <h2 class="text-center"> 欢迎来到《Go 语言编程之旅：一起用 Go 做项目》聊天室 </h2>
            </div>
        </div>
    </div>
    <div class="row">
        <div class="col-md-1"></div>
        <div class="col-md-6">
            <div> 聊天内容 </div>
            <div class="msg-list" id="msg-list">
                <div class="message"
                    v-for="msg in msglist"
                    v-bind:class="{ system: msg.type==1, myself: msg.user.nickname==curUser.nickname }"
                    >
                    <div class="meta" v-if="msg.user.nickname"><span class="author">${ msg.user.nickname }</span> at ${ formatDate(msg.msg_time) }</div>
                    <div>
                        <span class="content" style="white-space: pre-wrap;">${ msg.content }</span>
                    </div>
                </div>
            </div>
        </div>
        <div class="col-md-4">
            <div> 当前在线用户数：<font color="red">${ onlineUserNum }</font></div>
            <div class="user-list">
                <div class="user" v-for="user in users">
                    用户：@${ user.nickname } 加入时间：${ formatDate(user.enter_at) }
                </div>
            </div>
        </div>
    </div>
    <div class="row">
        <div class="col-md-1"></div>
        <div class="col-md-10">
            <div class="user-input">
                <div class="usertip text-center">${ usertip }</div>
                <div class="form-inline has-success text-center" style="margin-bottom: 10px;">
                    <div class="input-group">
                        <span class="input-group-addon"> 您的昵称 </span>
                        <input type="text" v-model="curUser.nickname" v-bind:disabled="joined" class="form-control" aria-describedby="inputGroupSuccess1Status">
                    </div>
                    <input type="submit" class="form-control btn-primary text-center" v-on:click="leavechat" v-if="joined" value="离开聊天室">
                    <input type="submit" class="form-control btn-primary text-center" v-on:click="joinchat" v-else="joined" value="进入聊天室">
                </div>
                <textarea id="chat-content" rows="3" class="form-control" v-model="content"
                          @keydown.enter.prevent.exact="sendChatContent"
                          @keydown.meta.enter="lineFeed"
                          @keydown.ctrl.enter="lineFeed"
                          placeholder="在此收入聊天内容。ctrl/command+enter 换行，enter 发送"></textarea>&nbsp;
                <input type="button" value="发送(Enter)" class="btn-primary form-control" v-on:click="sendChatContent">
            </div>
        </div>
    </div>
</div>
```

之后打开终端，启动聊天室。打开浏览器访问 localhost:2022，出现如下界面：

![image](https://golang2.eddycjy.com/images/ch4/start-ui.png)

根据前面的讲解知道，这是通过 HTTP 请求了 `/` 这个路由，对应到如下 handle 的代码：

```go
// server/home.go
func homeHandleFunc(w http.ResponseWriter, req *http.Request) {
	tpl, err := template.ParseFiles(rootDir + "/template/home.html")
	if err != nil {
		fmt.Fprint(w, "模板解析错误！")
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		fmt.Fprint(w, "模板执行错误！")
		return
	}
}
```

代码只是简单的渲染页面。

> 小提示：因为模板中不涉及到任何服务端渲染，所以，在部署时，如果使用 Nginx 这样的 WebServer，完全可以直接将 index 指向 home.html，而不经过 Go 渲染。

我们的前端使用了 Vue，如果你对 Vue 完全不了解，建议你可以到 Vue 的官网学习一下，它是国人开发的，中文文档很友好。

在看到的页面中，在「您的昵称」处输入：polaris，点击「进入聊天室」。

![image](https://golang2.eddycjy.com/images/ch4/first-enter.png)

这个过程涉及到的网络环节前面已经抓包讲解过，这里主要看下前端 JS 部分的实现。

```js
// 只保留了 WebSocket 相关的核心代码
if ("WebSocket" in window) {
    let host = location.host;
    // 打开一个 websocket 连接
    gWS = new WebSocket("ws://"+host+"/ws?nickname="+this.nickname);

    gWS.onopen = function () {
        // WebSocket 已连接上的回调
    };

    gWS.onmessage = function (evt) {
        let data = JSON.parse(evt.data);
        if (data.type == 2) {
            that.usertip = data.content;
            that.joined = false;
        } else if (data.type == 3) {
            // 用户列表
            that.users.splice(0);
            for (let nickname in data.users) {
                that.users.push(data.users[nickname]);
            }
        } else {
            that.addMsg2List(data);
        }
    };

    gWS.onerror = function(evt) {
        console.log("发生错误：");
        console.log(evt);
    };

    gWS.onclose = function () {
        console.log("连接已关闭...");
    };

} else {
    alert("您的浏览器不支持 WebSocket!");
}
```

前端 WebSocket 的核心是构造函数和几个回调函数。

- new WebSocket：创建一个 WebSocket 实例，提供服务端的 ws 地址，地址可以跟 HTTP 协议一样，加上请求参数。注意，如果你使用 HTTPS 协议，相应的 WebSocket 地址协议要改为 wss；
- WebSocket.onopen：用于指定连接成功后的回调函数；
- WebSocket.onerror：用于指定连接失败后的回调函数；
- WebSocket.onmessage：用于指定当从服务器接收到信息时的回调函数；
- WebSocket.onclose：用于指定连接关闭后的回调函数；

在用户点击进入聊天室时，根据 Vue 绑定的事件，会执行上面的代码，发起 WebSocket 连接，服务端会将相关信息通过 WebSocket 长连接返回给客户端，客户端通过 `WebSocket.onmessage` 回调进行处理。

得益于 Vue 的双向绑定，在数据显示、事件绑定等方面，处理起来很方便。

关于前端的实现，这里有几点提醒下读者：

- Vue 默认的分隔符是 `{{}}`，和 Go 的一样，避免冲突进行了修改；
- ctrl/command+enter 换行，enter 发送 的事件绑定需要留意下；
- 因为我们没有实现注册登录的功能，为了方便，做了自动记住上次昵称的处理，存入 localStorage 中；
- 通过 setInterval 来自动重连；
- 注意用户列表的处理：`that.users.splice(0)` ，如果 `that.users = []` 是不行的，这涉及到 Vue 怎么监听数据的问题；
- WebSocket 有两个方法：send 和 close，一个用来发送消息，一个用于主动断开链接；
- WebSocket 有一个属性 readyState 可以判定当前连接的状态；

## 4.5.2 后端流程关键代码

后端关键流程和本章第 1 节的关键流程是类似的。（为了方便，我们给涉及到的几个 goroutine 进行命名：运行 WebSocketHandleFunc 的 goroutine 叫 conn goroutine，也可以称为 read goroutine；给用户发送消息的 goroutine 叫 write goroutine；广播器所在 goroutine 叫 broadcaster goroutine）。

```go
// server/websocket.go
func WebSocketHandleFunc(w http.ResponseWriter, req *http.Request) {
	// Accept 从客户端接收 WebSocket 握手，并将连接升级到 WebSocket。
	// 如果 Origin 域与主机不同，Accept 将拒绝握手，除非设置了 InsecureSkipVerify 选项（通过第三个参数 AcceptOptions 设置）。
	// 换句话说，默认情况下，它不允许跨源请求。如果发生错误，Accept 将始终写入适当的响应
	conn, err := websocket.Accept(w, req, nil)
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}

	// 1. 新用户进来，构建该用户的实例
	nickname := req.FormValue("nickname")
	if l := len(nickname); l < 2 || l > 20 {
		log.Println("nickname illegal: ", nickname)
		wsjson.Write(req.Context(), conn, logic.NewErrorMessage("非法昵称，昵称长度：4-20"))
		conn.Close(websocket.StatusUnsupportedData, "nickname illegal!")
		return
	}
	if !logic.Broadcaster.CanEnterRoom(nickname) {
		log.Println("昵称已经存在：", nickname)
		wsjson.Write(req.Context(), conn, logic.NewErrorMessage("该昵称已经已存在！"))
		conn.Close(websocket.StatusUnsupportedData, "nickname exists!")
		return
	}

	user := logic.NewUser(conn, nickname, req.RemoteAddr)

	// 2. 开启给用户发送消息的 goroutine
	go user.SendMessage(req.Context())

	// 3. 给当前用户发送欢迎信息
	user.MessageChannel <- logic.NewWelcomeMessage(nickname)

	// 给所有用户告知新用户到来
	msg := logic.NewNoticeMessage(nickname + " 加入了聊天室")
	logic.Broadcaster.Broadcast(msg)

	// 4. 将该用户加入广播器的用户列表中
	logic.Broadcaster.UserEntering(user)
	log.Println("user:", nickname, "joins chat")

	// 5. 接收用户消息
	err = user.ReceiveMessage(req.Context())

	// 6. 用户离开
	logic.Broadcaster.UserLeaving(user)
	msg = logic.NewNoticeMessage(user.NickName + " 离开了聊天室")
	logic.Broadcaster.Broadcast(msg)
	log.Println("user:", nickname, "leaves chat")

	// 根据读取时的错误执行不同的 Close
	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "")
	} else {
		log.Println("read from client error:", err)
		conn.Close(websocket.StatusInternalError, "Read from client error")
	}
}
```

根据注释，我们就关键流程步骤一一讲解。

### 1、新用户进来，创建一个代表该用户的 User 实例

该聊天室没有实现注册登录功能，为了方便识别谁是谁，我们简单要求输入昵称。昵称在建立 WebSocket 连接时，通过 HTTP 协议传递，因此可以通过 http.Request 获取到，即：`req.FormValue("nickname")`。虽然没有注册功能，但依然要解决昵称重复的问题。这里必须引出 Broadcaster 了。

**广播器 broadcaster**

聊天室，顾名思义，消息要进行广播。broadcaster 就是一个广播器，负责将用户发送的消息广播给聊天室里的其他人。先看看广播器的定义。

```go
// logic/broadcast.go
// broadcaster 广播器
type broadcaster struct {
  // 所有聊天室用户
	users map[string]*User

	// 所有 channel 统一管理，可以避免外部乱用

	enteringChannel chan *User
	leavingChannel  chan *User
	messageChannel  chan *Message

	// 判断该昵称用户是否可进入聊天室（重复与否）：true 能，false 不能
	checkUserChannel      chan string
	checkUserCanInChannel chan bool
}
```

这里使用了“单例模式”，在 broadcat.go 中实例化一个广播器实例：Broadcaster，方便外部使用。

因为 Broadcaster.Broadcast() 在一个单独的 goroutine 中运行，按照 Go 语言的原则，应该通过通信来共享内存。因此，我们定义了 5 个 channel，用于和其他 goroutine 进行通信。

- enteringChannel：用户进入聊天室时，通过该 channel 告知 Broadcaster，即将该用户加入 Broadcaster 的 users 中；
- leavingChannel：用户离开聊天室时，通过该 channel 告知 Broadcaster，即将该用户从 Broadcaster 的 users 中删除，同时需要关闭该用户对应的 messageChannel，避免 goroutine 泄露，后文会讲到；
- messageChannel：用户发送的消息，通过该 channel 告知 Broadcaster，之后 Broadcaster 将它发送给 users 中的用户；
- checkUserChannel：用来接收用户昵称，方便 Broadcaster 所在 goroutine 能够无锁判断昵称是否存在；
- checkUserCanInChannel：用来回传该用户昵称是否已经存在；

判断用户是否存在时，利用了上面提到的两个 channel，看看具体的实现：

```go
func (b *broadcaster) CanEnterRoom(nickname string) bool {
	b.checkUserChannel <- nickname

	return <-b.checkUserCanInChannel
}
```

![image](https://golang2.eddycjy.com/images/ch4/user-exists-goroutine.png)

如上图所示，两个 goroutine 通过两个 channel 进行通讯，因为 conn goroutine（代表用户连接 goroutine）可能很多，通过这种方式，避免了使用锁。

> 虽然没有显示使用锁，但这里要求 checkUserChannel 必须是无缓冲的，否则判断可能会出错。

如果用户已存在，连接会断开；否则创建该用户的实例：

```go
user := logic.NewUser(conn, nickname, req.RemoteAddr)
```

这里又引出了 User 类型。

```go
// logic/user.go
type User struct {
	UID            int           `json:"uid"`
	NickName       string        `json:"nickname"`
	EnterAt        time.Time     `json:"enter_at"`
	Addr           string        `json:"addr"`
	MessageChannel chan *Message `json:"-"`

	conn *websocket.Conn
}
```

一个 User 代表一个进入了聊天室的用户。

### 2、开启给用户发送消息的 goroutine

服务一个用户（一个连接），至少需要两个 goroutine：一个读用户发送的消息，一个给用户发送消息。

```go
go user.SendMessage(req.Context())

// logic/user.go
func (u *User) SendMessage(ctx context.Context) {
	for msg := range u.MessageChannel {
		wsjson.Write(ctx, u.conn, msg)
	}
}
```

当前连接已经在一个新的 goroutine 中了，我们用来做消息读取用，同时新开一个 goroutine 用来给用户发送消息。

具体的消息发送是，通过 for-range 从当前用户的 MessageChannel 中读取消息，然后通过 `nhooyr.io/websocket/wsjson` 包的 Write 方法发送给浏览器，该库会自动做 JSON 编码。

前文提到过，这里是一个长期运行的 goroutine，存在泄露的风险。当用户退出时，一定要让给 goroutine 退出，退出方法就是关闭 u.MessageChannel 这个 channel。

### 3、新用户进入，给用户发消息

```go
// 给当前用户发送欢迎信息
user.MessageChannel <- logic.NewWelcomeMessage(nickname)

// 给所有用户告知新用户到来
msg := logic.NewNoticeMessage(nickname + " 加入了聊天室")
logic.Broadcaster.Broadcast(msg)
```

新用户进入，一方面给 TA 发送欢迎的消息，另一方面需要通知聊天室的其他人，有新用户进来了。

这里又引出了第三个类型：Message。

```go
// 给用户发送的消息
type Message struct {
  // 哪个用户发送的消息
	User    *User     `json:"user"`
	Type    int       `json:"type"`
	Content string    `json:"content"`
	MsgTime time.Time `json:"msg_time"`

	Users map[string]*User `json:"users"`
}
```

这里着重需要关注的是 Type 字段，用它来判定消息在客户端如何显示。有如下几种类型的消息：

```go
const (
	MsgTypeNormal   = iota // 普通 用户消息
	MsgTypeSystem          // 系统消息
	MsgTypeError           // 错误消息
	MsgTypeUserList        // 发送当前用户列表
)
```

消息一共分成三大类：1）在聊天室窗口显示；2）页面错误提示（比如昵称已存在）；3）当前聊天室用户列表。其中，在聊天室窗口显示，又分为用户消息和系统消息。

Message 结构中几个字段的意思就清楚了，特别说明的是，字段 User 代表该消息的属主：普通用户还是系统。所以，特别实例化了一个系统用户：

```go
// 系统用户，代表是系统主动发送的消息
var System = &User{}
```

它的 UID 是 0。

接下来看看发送消息的过程，发送消息分两情况，它们的处理方式有些差异：

- 给单个用户（当前）用户发送消息
- 给聊天室其他用户广播消息

用两个图来来表示这两种情况。

![image](https://golang2.eddycjy.com/images/ch4/send-message-single.png)

给当前用户发送消息的情况比较简单：conn goroutine 通过用户实例（User）的字段 MessageChannel 将 Message 发送给 write goroutine。

![image](https://golang2.eddycjy.com/images/ch4/send-message-broadcast.png)

给聊天室其他用户广播消息自然需要通过 broadcaster goroutine 来实现：conn goroutine 通过 Broadcaster 的 MessageChannel 将 Message 发送出去，broadcaster goroutine 遍历自己维护的聊天室用户列表，通过 User 实例的 MessageChannel 将消息发送给 write goroutine。

> 提示：细心的读者可能会想到 broadcaster 这里可能会成为瓶颈，用户量大时，可能会有消息挤压，这一点后续讨论。

### 4. 将该用户加入广播器的用户列表中

这个过程很简单，一行代码，最终通过 channel 发送到 Broadcaster 中。

```go
logic.Broadcaster.UserEntering(user)
```

### 5. 接收用户消息

跟给用户发送消息类似，调用的是 user 的方法：

```go
err = user.ReceiveMessage(req.Context())
```

该方法的实现如下：

```go
// logic/user.go
func (u *User) ReceiveMessage(ctx context.Context) error {
	var (
		receiveMsg map[string]string
		err        error
	)
	for {
		err = wsjson.Read(ctx, u.conn, &receiveMsg)
		if err != nil {
			// 判定连接是否关闭了，正常关闭，不认为是错误
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				return nil
			}

			return err
		}

		// 内容发送到聊天室
		sendMsg := NewMessage(u, receiveMsg["content"])
		Broadcaster.Broadcast(sendMsg)
	}
}
```

逻辑较简单，即通过 `nhooyr.io/websocket/wsjson` 包读取用户输入数据，构造出 Message 实例，广播出去。

这里特别提一下 Go1.13 中 errors 包的新功能，实际项目中可能大家还没有用到。

```go
var closeErr websocket.CloseError
if errors.As(err, &closeErr) {
  return nil
}
```

当用户主动退出聊天室时，`wsjson.Read` 会返回错，除此之外，可能还有其他原因导致返回错误。这两种情况应该加以区分。这得益于 Go1.13 errors 包的新功能和 nhooyr.io/websocket 包对该新功能的支持，我们可以通过 As 来判定错误是不是连接关闭导致的。

### 6. 用户离开

用户可以主动或由于其他原因离开聊天室，这时候 user.ReceiveMessage 方法会返回，执行下面的代码：

```go
// 6. 用户离开
logic.Broadcaster.UserLeaving(user)
msg = logic.NewNoticeMessage(user.NickName + " 离开了聊天室")
logic.Broadcaster.Broadcast(msg)
log.Println("user:", nickname, "leaves chat")

// 根据读取时的错误执行不同的 Close
if err == nil {
  conn.Close(websocket.StatusNormalClosure, "")
} else {
  log.Println("read from client error:", err)
  conn.Close(websocket.StatusTryAgainLater, "Read from client error")
}
```

这里主要做了三件事情：

- 在 Broadcaster 中注销该用户；
- 给聊天室中其他还在线的用户发送通知，告知该用户已离开；
- 根据 err 处理不同的 Close 行为。关于 Close 的 Status 可以参考 rfc6455 的 第 7.4 节；

## 4.5.3 小结

到这里我们把最核心的流程讲解完了。但我们略过了 broadcaster 中的关键代码，下节我们主要讲解广播器：broadcaster。

# 4.6 实现聊天室：广播器

上一节介绍了聊天室的核心流程，其中多次提到了 Broadcaster，但没有过多涉及到其中的细节。本节我们详细介绍它的实现：广播器，这是聊天室的一个核心模块。

## 4.6.1 单例模式

Go 不是完全面向对象的语言，只支持部分面向对象的特性。面向对象中的单例模式是一个常见、简单的模式。前文提到，广播器中我们应用了单例模式，这里进行必要的讲解。

### 4.6.1.1 简介

英文名称：Singleton Pattern，该模式规定一个类只允许有一个实例，而且自行实例化并向整个系统提供这个实例。因此单例模式的要点有：1）只有一个实例；2）必须自行创建；3）必须自行向整个系统提供这个实例。

单例模式主要避免一个全局使用的类频繁地创建与销毁。当你想控制实例的数量，或有时候不允许存在多实例时，单例模式就派上用场了。

为了更好的讲解单例模式，我们先使用 Java 来描述它，之后回到 Go 中来。

![image](https://golang2.eddycjy.com/images/ch4/singleton.png)

通过该类图我们可以看出，实现一个单例模式有如下要求：

- 私有、静态的类实例变量；
- 构造函数私有化；
- 静态工厂方法，返回此类的唯一实例；

根据实例化的时机，单例模式一般分成饿汉式和懒汉式。

- 饿汉式：在定义 instance 时直接实例化，private static Singleton instance = new Singleton();
- 懒汉式：在 getInstance 方法中进行实例化；

那两者有什么区别或优缺点？饿汉式单例类在自己被加载时就将自己实例化。即便加载器是静态的，饿汉式单例类被加载时仍会将自己实例化。单从资源利用率角度讲，这个比懒汉式单例类稍差些。从速度和反应时间角度讲，则比懒汉式单例类稍好些。然而，懒汉式单例类在实例化时，必须处理好在多个线程同时首次引用此类时的访问限制问题，特别是当单例类作为资源控制器在实例化时必须涉及资源初始化，而资源初始化很有可能耗费时间。这意味着出现多线程同时首次引用此类的几率变得较大。

### 4.6.1.2 单例模式的 Java 实现

结合上面的讲解，以一个计数器为例，我们看看 Java 中饿汉式的实现：

```java
public class Singleton {
  private static final Singleton instance = new Singleton();
  private int count = 0;
  private Singleton() {}
  public static Singleton getInstance() {
    return instance;
  }
  public int Add() int {
    this.count++;
    return this.count;
  }
}
```

代码很简单，不过多解释。直接看懒汉式的实现：

```java
public class Singleton {
  private static Singleton instance = null;
  private int count = 0;
  private Singleton() {}
  public static synchronized Singleton getInstance() {
    if (instance == null) {
      instance = new Singleton();
    }
    return instance;
  }
  public int Add() int {
    this.count++;
    return this.count;
  }
}
```

主要区别在于 getInstance 的实现，要注意 synchronized ，避免多线程时出现问题。

### 4.6.1.3 单例模式的 Go 实现

回到 Go 语言，看看 Go 语言如何实现单例。

```go
// 饿汉式单例模式
package singleton

type singleton struct {
  count int
}

var Instance = new(singleton)

func (s *singleton) Add() int {
  s.count++
  return s.count
}
```

前面说了，Go 只支持部分面向对象的特性，因此看起来有点不太一样：

- 类（结构体 singleton）本身非公开（小写字母开头，非导出）;
- 没有提供导出的 GetInstance 工厂方法（Go 没有静态方法），而是直接提供包级导出变量 Instance；

这样使用：

```go
c := singleton.Instance.Add()
```

看看懒汉式单例模式在 Go 中如何实现：

```go
// 懒汉式单例模式
package singleton

import (
	"sync"
)

type singleton struct {
  count int
}

var (
  instance *singleton
  mutex sync.Mutex
)

func New() *singleton {
  mutex.Lock()
  if instance == nil {
    instance = new(singleton)
  }
  mutex.Unlock()
  
  return instance
}

func (s *singleton) Add() int {
  s.count++
  return s.count
}
```

代码多了不少：

- 包级变量变成非导出（instance），注意这里类型应该用指针，因为结构体的默认值不是 nil；
- 提供了工厂方法，按照 Go 的惯例，我们命名为 New()；
- 多 goroutine 保护，对应 Java 的 synchronized，Go 使用 sync.Mutex；

关于懒汉式有一个“双重检查”，这是 C 语言的一种代码模式。

在上面 New() 函数中，同步化（锁保护）实际上只在 instance 变量第一次被赋值之前才有用。在 instance 变量有了值之后，同步化实际上变成了一个不必要的瓶颈。如果能够有一个方法去掉这个小小的额外开销，不是更加完美吗？因此出现了“双重检查”。看看 Go 如何实现“双重检查”，只看 New() 代码：

```go
func New() *singleton {
  if instance == nil {	// 第一次检查（①）
    // 这里可能有多于一个 goroutine 同时达到（②）
    mutex.Lock()
    // 这里每个时刻只会有一个 goroutine（③）
    if instance == nil {	// 第二次检查（④）
      instance = new(singleton)
    }
    mutex.Unlock()
  }
  
  return instance
}
```

有读者可能看不懂上面代码的意思，这里详细解释下。假设 goroutine X 和 Y 作为第一批调用者同时或几乎同时调用 New 函数。

1. 因为 goroutine X 和 Y 是第一批调用者，因此，当它们进入此函数时，instance 变量是 nil。因此 goroutine X 和 Y 会同时或几乎同时到达位置 ①；
2. 假设 goroutine X 会先达到位置 ②，并进入 mutex.Lock() 达到位置 ③。这时，由于 mutex.Lock 的同步限制，goroutine Y 无法到达位置 ③，而只能在位置 ② 等候；
3. goroutine X 执行 instance = new(singleton) 语句，使得 instance 变量得到一个值，即对 singleton 实例的引用。此时，goroutine Y 只能继续在位置 ② 等候；
4. goroutine X 释放锁，返回 instance，退出 New 函数；
5. goroutine Y 进入 mutex.Lock()，到达位置 ③，进而到达位置 ④。由于 instance 变量已经不是 nil，因此 goroutine Y 释放锁，返回 instance 所引用的 singleton 实例（也就是 goroutine X 锁创建的 singleton 实例），退出 New 函数；

到这里，goroutine X 和 Y 得到了同一个 singleton 实例。可见上面的 New 函数中，锁仅用来避免多个 goroutine 同时实例化 singleton。

相比前面的版本，双重检查版本，只要 instance 实例化后，锁永远不会执行了，而前面版本每次调用 New 获取实例都需要执行锁。性能很显然，我们可以基准测试来验证：（双重检查版本 New 重命名为 New2）

```go
package singleton_test

import (
	"testing"

	"github.com/go-programming-tour-book/go-demo/singleton"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		singleton.New()
	}
}

func BenchmarkNew2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		singleton.New2()
	}
}
```

因为是单例，所以两个基准测试需要分别执行。

New1 的结果：

```
$ go test -benchmem -bench ^BenchmarkNew$ github.com/go-programming-tour-book/go-demo/singleton
goos: darwin
goarch: amd64
pkg: github.com/go-programming-tour-book/go-demo/singleton
BenchmarkNew-8   	80470467	        14.0 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/go-programming-tour-book/go-demo/singleton	1.151s
```

New2 的结果：

```
$ go test -benchmem -bench ^BenchmarkNew2$ github.com/go-programming-tour-book/go-demo/singleton
goos: darwin
goarch: amd64
pkg: github.com/go-programming-tour-book/go-demo/singleton
BenchmarkNew2-8   	658810392	         1.80 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/go-programming-tour-book/go-demo/singleton	1.380s
```

New2 快十几倍。

Go 语言单例模式，推荐一般优先考虑使用饿汉式。

## 4.6.2 广播器的实现

本章第 6 节我们看过广播器结构的定义：

```go
// broadcaster 广播器
type broadcaster struct {
	// 所有聊天室用户
	users map[string]*User

	// 所有 channel 统一管理，可以避免外部乱用

	enteringChannel chan *User
	leavingChannel  chan *User
	messageChannel  chan *Message

	// 判断该昵称用户是否可进入聊天室（重复与否）：true 能，false 不能
	checkUserChannel      chan string
	checkUserCanInChannel chan bool
}
```

很显然，广播器全局应该只有一个，所以是典型的单例。我们使用饿汉式实现。

```go
var Broadcaster = &broadcaster{
	users: make(map[string]*User),

	enteringChannel: make(chan *User),
	leavingChannel:  make(chan *User),
	messageChannel:  make(chan *Message, MessageQueueLen),

	checkUserChannel:      make(chan string),
	checkUserCanInChannel: make(chan bool),
}
```

导出的 Broadcaster 代表广播器的唯一实例，通过 logic.Broadcaster 来使用这个单例。

在本章第 4 节时提到了通过如下语句启动广播器：

```go
go logic.Broadcaster.Start()
```

现在看看 Start 的具体实现：

```go
// logic/broadcast.go

// Start 启动广播器
// 需要在一个新 goroutine 中运行，因为它不会返回
func (b *broadcaster) Start() {
	for {
		select {
		case user := <-b.enteringChannel:
			// 新用户进入
			b.users[user.NickName] = user

			b.sendUserList()
		case user := <-b.leavingChannel:
			// 用户离开
			delete(b.users, user.NickName)
			// 避免 goroutine 泄露
			user.CloseMessageChannel()

			b.sendUserList()
		case msg := <-b.messageChannel:
			// 给所有在线用户发送消息
			for _, user := range b.users {
				if user.UID == msg.User.UID {
					continue
				}
				user.MessageChannel <- msg
			}
		case nickname := <-b.checkUserChannel:
			if _, ok := b.users[nickname]; ok {
				b.checkUserCanInChannel <- false
			} else {
				b.checkUserCanInChannel <- true
			}
		}
	}
}
```

核心关注的知识点：

- 需要在一个新 goroutine 中进行，因为它不会返回。注意这里并非说，只要不会返回的函数/方法就应该在新的 goroutine 中运行，虽然大部分情况是这样；
- Go 有一个最佳实践：应该让调用者决定并发（启动新 goroutine），这样它清楚自己在干什么。Start 的设计遵循了这一实践，没有自己内部开启新的 goroutine；
- for + select 形式，是 Go 中一种较常用的编程模式，可以不断监听各种 channel 的状态，有点类似 Unix 系统的 select 系统调用；
- 每新开一个 goroutine，你必须知道它什么时候会停止。这一句 `user.CloseMessageChannel()` 就涉及到 goroutine 的停止，避免泄露；

### 4.6.2.1 select-case 结构

Go 中有一个专门为 channel 设计的 select-case 分支流程控制语法。 此语法和 switch-case 分支流程控制语法很相似。 比如，select-case 流程控制代码块中也可以有若干 case 分支和最多一个 default 分支。 但是，这两种流程控制也有很多不同点。在一个 select-case 流程控制中：

- select 关键字和 { 之间不允许存在任何表达式和语句；
- fallthrough 语句不能使用；
- 每个 case 关键字后必须跟随一个 channel 接收数据操作或者一个 channel 发送数据操作，所以叫做专门为 channel 设计的；
- 所有的非阻塞 case 操作中将有一个被随机选择执行（而不是按照从上到下的顺序），然后执行此操作对应的 case 分支代码块；
- 在所有的 case 操作均阻塞的情况下，如果 default 分支存在，则 default 分支代码块将得到执行； 否则，当前 goroutine 进入阻塞状态；

所以，广播器的 Start 方法中，当所有 case 操作都阻塞时，Start 方法所在的 goroutine 进入阻塞状态。

另外，根据以上规则，一个不含任何分支的 select-case 代码块 select{} 将使当前 goroutine 处于永久阻塞状态，这可以用于一些服务开发中，如果你见到了 select{} 这样的写法不要惊讶了。比如：

```go
func main() {
  go func() {
    // 该函数不会退出
    for {
      // 省略代码
    }
  } ()
  
  select {}
}
```

这样保证 main goroutine 永远阻塞，让其他 goroutine 运行。但如果除了当前因为 select{} 阻塞的 goroutine 外，没有其他可运行的 goroutine，会导致死锁。因此下面的代码会死锁：

```go
func main() {
  select {}
}
```

运行报错：

> fatal error: all goroutines are asleep - deadlock!

### 4.6.2.2 goroutine 泄露

在 Go 中，goroutine 的创建成本低廉且调度效率高。Go 运行时能很好的支持具有成千上万个 goroutine 的程序运行，数十万个也并不意外。但是，goroutine 在内存占用方面却需要谨慎，内存资源是有限的，因此你不能创建无限的 goroutine。

每当你在程序中使用 go 关键字启动 goroutine 时，你必须知道该 goroutine 将在何时何地退出。如果你不知道答案，那可能会内存泄漏。

我们回过头梳理下聊天室项目有哪些新启动的 goroutine。

1）启动广播器

```go
// 广播消息处理
go logic.Broadcaster.Start()
```

我们很清楚，该广播器的生命周期是和程序生命周期一致的，因此它不应该结束。

2）负责给用户发送消息的 goroutine

在 WebSocketHandleFunc 函数中：

```go
// 2. 开启给用户发送消息的 goroutine
go user.SendMessage(req.Context())
```

user.SendMessage 的具体实现是：

```go
func (u *User) SendMessage(ctx context.Context) {
	for msg := range u.MessageChannel {
		wsjson.Write(ctx, u.conn, msg)
	}
}
```

根据 for-range 用于 channel 的语法，默认情况下，for-range 不会退出。很显然，如果我们不做特殊处理，这里的 goroutine 会一直存在。而实际上，当用户离开聊天室时，它对应连接的写 goroutine 应该终止。这也就是上面 Start 方法中，在用户离开聊天室的 channel 收到消息时，要将用户的 MessageChannel 关闭的原因。MessageChannel 关闭了，`for msg := range u.MessageChannel` 就会退出循环，goroutine 结束，避免了内存泄露。

3）库开启的 goroutine

在本章第 1 节，我们用 TCP 实现简单聊天室时，每一个用户到来，都会新开启一个 goroutine 服务该用户。在我们的 WebSocket 聊天室中，这个新开启 goroutine 的动作，由库给我们做了（具体是 net/http 库）。也许你不明白为什么是 http 库开启的，这里教大家一个思考思路。

一个程序能够长时间运行而不停止，肯定是程序里有死循环。在本章第 1 节中，我们自己写了一个死循环：

```go
func main() {
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}
```

那 WebSocket 版本的聊天室的死循环在哪里呢？回到 cmd/chatroom/main.go ：

```go
func main() {
	fmt.Printf(banner, addr)

	server.RegisterHandle()

	log.Fatal(http.ListenAndServe(addr, nil))
}
```

很显然在不出错时，http.ListenAndServe(addr, nil) 函数调用不会返回。因为 HTTP 协议基于 TCP 协议，因此 http 库中肯定存在类似我们上面实现 tcp 聊天室时的死循环代码。

通过跟踪 http.ListenAndServe -> Server.ListenAndServe，我们找到了如下代码：

```go
func (srv *Server) ListenAndServe() error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}
```

这一句 ln, err := net.Listen(“tcp”, addr) 和我们自己实现时一样。接着看 Server.Serve 方法，看到 for 死循环了（只保留关键代码）：

```go
func (srv *Server) Serve(l net.Listener) error {
	...
	origListener := l
	l = &onceCloseListener{Listener: l}
	defer l.Close()
  ...
	for {
		rw, e := l.Accept()
		if e != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
    ...
		tempDelay = 0
		c := srv.newConn(rw)
		c.setState(c.rwc, StateNew) // before Serve can return
		go c.serve(ctx)
	}
}
```

在这个死循环的最后一句：go c.serve(ctx) ，即当有新客户端连接时，开启一个新 goroutine 为其服务。最终，根据我们定义的路由，进入相应的函数进行处理。

那由 net/http 开启的 goroutine 什么时候结束呢？根据上面的分析，该 goroutine 最终会执行到我们定义的路由处理器中。所以，当我们的处理函数返回后，该 goroutine 也就结束了。因此，我们要确保 WebSocketHandleFunc 函数是有可能返回的。通过上一节的分析知道，当用户退出聊天室或其他原因导致连接断开时，User.ReceiveMessage 中的循环都会结束，函数退出。

**总结一下容易导致 goroutine 或内存泄露的场景**

1）time.After

这是很多人实际遇到过的内存泄露场景。如下代码：

```go
func ProcessMessage(ctx context.Context, in <-chan string) {
	for {
		select {
		case s, ok := <-in:
			if !ok {
				return
			}
			// handle `s`
		case <-time.After(5 * time.Minute):
			// do something
		case <-ctx.Done():
			return
		}
	}
}
```

在标准库 time.After 的文档中有一段说明：

> 等待持续时间过去，然后在返回的 channel 上发送当前时间。它等效于 NewTimer().C。在计时器触发之前，计时器不会被垃圾收集器回收。

所以，如果还没有到 5 分钟，该函数返回了，计时器就不会被 GC 回收，因此出现了内存泄露。因此大家使用 time.After 时一定要仔细，一般建议不用它，而是使用 time.NewTimer：

```go
func ProcessMessage(ctx context.Context, in <-chan string) {
	idleDuration := 5 * time.Minute
	idleDelay := time.NewTimer(idleDuration)
  // 这句必须的
	defer idleDelay.Stop()
	for {
		idleDelay.Reset(idleDuration)
		select {
		case s, ok := <-in:
			if !ok {
				return
			}
			// handle `s`
		case <-idleDelay.C:
			// do something
		case <-ctx.Done():
			return
		}
	}
}
```

2）发送到 channel 阻塞导致 goroutine 泄露

假如存在如下的程序：

```go
func process(term string) error {
     // 创建一个在 100 ms 内取消的 context
     ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
     defer cancel()

     // 为 goroutine 创建一个传递结果的 channel
     ch := make(chan string)

     // 启动一个 goroutine 来寻找记录，然后得到结果
     // 并将返回值从 channel 中传回
     go func() {
         ch <- search(term)
     }()

     select {
     case <-ctx.Done():
         return errors.New("search canceled")
     case result := <-ch:
         fmt.Println("Received:", result)
         return nil
    }
 }

// search 模拟成一个查找记录的函数
// 在查找记录时。执行此工作需要 200 ms。
func search(term string) string {
     time.Sleep(200 * time.Millisecond)
     return "some value"
}
```

这是一个挺常见的场景：要进行一些耗时操作，因此开启一个 goroutine 进行处理，它的处理结果，通过 channel 回传给原来的 goroutine；同时，这个耗时操作不能太长，因此有了 WithTimeout Context。最后通过 select-case 来监控 ctx.Done 和传递数据的 channel 是否就绪。

如果超时没处理完，ctx.Done 会执行，函数返回，新开启的 goroutine 会因为 channel 中的另一端没有就绪的接收 goroutine 而一直阻塞，导致 goroutine 泄露。

解决这种因为发送到 channel 阻塞导致 goroutine 泄露的简单办法是将 channel 改为有缓冲的 channel，并保证容量充足。比如上面例子，将 ch 改为：ch := make(chan string, 1) 即可。

3）从 channel 接收阻塞导致 goroutine 泄露

我们聊天室可能导致 goroutine 泄露就属于这种情况。

```go
func (u *User) SendMessage(ctx context.Context) {
	for msg := range u.MessageChannel {
		wsjson.Write(ctx, u.conn, msg)
	}
}
```

for-range 循环直到 MessageChannel 这个 channel 关闭才会结束，因此需要有地方调用 close(u.MessageChannel)。

这种情况的另一种情形是：虽然没有 for-range，但给 channel 发送数据的一方已经不再发送数据了，接收的一方还在等待，这个等待会无限持续下去。唯一能取消它等待的就是 close 这个 channel。

### 4.6.2.3 广播器和外界的通信

从广播器的结构定义知道，它和其他 goroutine 的通信通过 channel 进行。判断用户是否存在的方式前面讲解了，这里看用户进入、离开和消息的通信。

```go
func(b *broadcaster) UserEntering(u *User) {
	b.enteringChannel <- u
}

func(b *broadcaster) UserLeaving(u *User) {
	b.leavingChannel <- u
}

func(b *broadcaster) Broadcast(msg *Message) {
	b.messageChannel <- msg
}
```

通过 channel 和其他 goroutine 通信，可以有几种方式，以用户进入聊天室为例。

**方式一：**

在 broadcast.go 中定义导出的 channel：var EnteringChannel = make(chan *User) 或者还是作为 broadcaster 的字段，但是导出的，各个 goroutine 都可以直接对 EnteringChannel 进行读写。这种方式显然不好，用面向对象说法，封装性不好，容易被乱用。

**方式二：**

broadcaster 结构和现在不变，通过方法将 enteringChannel 暴露出去：

```go
func (b *broadcaster) EnteringChannel() chan<- *User {
	return b.enteringChannel
}
```

前面讲过单向 channel，该方法的返回值类型：`chan<- *User` 就是一个单向 channel，它是只写的（only send channel），这限制了外部 goroutine 使用它的方式：只能往 channel 写数据，读取由我自己负责。

使用方式：logic.Broadcaster.EnteringChannel() <- user 。

整体上这种方式没有大问题，只是使用方式有点别扭。

**方式三：**

这种方式就是我们目前采用的方式，对外完全隐藏 channel，调用方不需要知道有 channel 的存在，只是感觉在做普通的方法调用。channel 的处理由内部自己处理，保证了安全性。这种方式比较优雅。

> User 中的 MessageChannel 我们没有采用这种方式，而是使用了方式一，让大家感受一下两种方式的不同。读者可以试着改为方式三的形式。

回到 Start 的循环中，这是在 broadcaster goroutine 中运行的，负责循环接收各个 channel 发送的数据，根据不同的 channel 处理不同的业务逻辑。

## 4.6.3 小结

本节讲解了单例模式，以及在 Go 中如何实现单例模式。

在讲解广播器的具体实现时引出了一个很重要的知识点：goroutine 泄漏，详细讲解了各种可能泄露的场景，读者在实际项目中一定要注意。

至此，一个 WebSocket 聊天室就实现了，但功能相对比较简单。下节我们会实现聊天室的一些非核心功能。

# 4.7 非核心功能

在日常的互联网项目开发中，一般先快速开发出一个最小可运行版本（MVP），投入市场验证。之后快速迭代，并进行其他非核心功能的开发。本文介绍聊天室的一些非核心功能如何实现。

> 说明：这里涉及到的功能，对一个聊天室来说，并不一定就是非核心功能。只是针对本书来说，它是非核心功能，因为没有它们，聊天室也可以正常运作。当然，核心还是非核心，并没有严格的界定。

## 4.7.1 @ 提醒功能

现在各种聊天工具或社区类网站，基本会支持 @ 提醒的功能。我们的聊天室如何实现它呢？

可以有两种做法：

1. @ 当做私聊，这条消息只会发给被 @ 的人，这么做的比较少，不过我们可以看如何实现；
2. 所有人都能收到，但被 @ 的人有不一样的显示提醒；

### 私信

先看第一种，只关注服务端的实现，但要告知对方这是一条私信。

在广播器中给所有用户广播消息时，做了一个过滤：消息不发给自己。

```go
for _, user := range b.users {
  if user.UID == msg.User.UID {
    continue
  }
  user.MessageChannel <- msg
}
```

私信因为是发给一个人，因此没必要遍历所有人。根据我们的设计，可以直接取出目标用户，进行消息发送。

为了方便服务端和客户端知晓这是一条私信消息，同时服务端发送前知道这是发给谁，在 Message 结构中增加一个字段 To：

```go
type Message struct {
	// 哪个用户发送的消息
	User    *User     `json:"user"`
	Type    int       `json:"type"`
	Content string    `json:"content"`
	MsgTime time.Time `json:"msg_time"`

	// 消息发送给谁，表明这是一条私信
	To string `json:"to"`

	Users map[string]*User `json:"users"`
}
```

接着在接收用户发送消息的地方，对接收到的用户消息进行解析，为 Message.To 字段赋值。

```go
// logic/user.go 中的 ReceiveMessage 方法
// 内容发送到聊天室
sendMsg := NewMessage(u, receiveMsg["content"])

// 解析 content，看是否是一条私信消息
sendMsg.Content = strings.TrimSpace(sendMsg.Content)
if strings.HasPrefix(sendMsg.Content, "@") {
  sendMsg.To = strings.SplitN(sendMsg.Content, " ", 2)[0][1:]
}
```

这句代码别感到奇怪：`strings.SplitN(sendMsg.Content, " ", 2)[0][1:]` ，Go 中，函数/方法返回的 slice 可以直接取值、reslice。

> 注意：这个实现要求必须是 @ 开始，消息中间的 @ 没有进行处理。

在广播器中需要对接收到的消息进行处理，由原来的代码改为（else 部分）：

```go
if msg.To == "" {
  // 给所有在线用户发送消息
  for _, user := range b.users {
    if user.UID == msg.User.UID {
      continue
    }
    user.MessageChannel <- msg
  }
} else {
  if user, ok := b.users[msg.To]; ok {
    user.MessageChannel <- msg
  } else {
    // 对方不在线或用户不存在，直接忽略消息
    log.Println("user:", msg.To, "not exists!")
  }
}
```

这里如果用户不存在或不在线，选择了直接忽略。当然可以有其他处理方法，比如当做普通广播消息发给所有人或提示发送者，对方目前的状态。

### 被 @ 的人收到提醒

这种方式是普遍采用的方式，聊天室中所有人都能收到消息，但被 @ 的人有提醒。

首先，我们依然需要在 Message 结构中增加一个 Ats 字段，表示能够一次 @ 多个人。

```go
type Message struct {
	// 哪个用户发送的消息
	User    *User     `json:"user"`
	Type    int       `json:"type"`
	Content string    `json:"content"`
	MsgTime time.Time `json:"msg_time"`

	// 消息 @ 了谁
	Ats []string `json:"ats"`

	Users map[string]*User `json:"users"`
}
```

其次，在 User 接收消息时（ReceiveMessage），同样需要解析出 @ 谁了。这次我们解析出所有被 @ 的人，而且不区分是不是以 @ 开始。

```go
// logic/user.go 中的 ReceiveMessage 方法
// 内容发送到聊天室
sendMsg := NewMessage(u, receiveMsg["content"])

// 解析 content，看看 @ 谁了
reg := regexp.MustCompile(`@[^\s@]{2,20}`)
sendMsg.Ats = reg.FindAllString(sendMsg.Content, -1)
```

这里要求昵称必须 2-20 个字符，跟前面的昵称校验保持一致。（昵称没有做特殊字符处理）

以上就是服务端要做的事情。

下面看看前端。因为前端不是重点，我们只会简单的提示有人 @ 你，在将消息 push 到 msgList 之前做提示，5 秒后消失。

```js
if (data.ats != null) {
		data.ats.forEach(function(nickname) {
        if (nickname == '@'+that.nickname) {
            that.usertip = '有人 @ 你了';
        }
    })
}
```

效果图如下：

![image](https://golang2.eddycjy.com/images/ch4/at-tip.png)

注意，以上做法，方法 1 代码在仓库中没有保留，方法 2 保留了。

## 4.7.2 敏感词处理

任何由用户产生内容的公开软件，都必须做好敏感词的处理。作为一个聊天室，当然要处理敏感词。

其实敏感词（包括广告）检测一直以来都是让人头疼的话题，很多大厂，比如微信、微博、头条等，每天产生大量内容，它们在处理敏感词这块，会投入很多资源。所以，这不是一个简单的问题，本书不可能深入探讨，但尽可能多涉及一些相关内容。

一般来说，目前敏感词处理有如下方法：

- 简单替换或正则替换
- DFA（Deterministic Finite Automaton，确定性有穷自动机算法）
- 基于朴素贝叶斯分类算法

### 1）简单替换或正则替换

```go
// 1. strings.Replace
keywords := []string{"坏蛋", "坏人", "发票", "傻子", "傻大个", "傻人"}
content := "不要发票，你就是一个傻子，只会发呆"
for _, keyword := range keywords {
  content = strings.ReplaceAll(content, keyword, "**")
}
fmt.Println(content)

// 2. strings.Replacer
replacer := strings.NewReplacer("坏蛋", "**", "坏人", "**", "发票", "**", "傻子", "**", "傻大个", "**", "傻人", "**")
fmt.Println(replacer.Replace("不要发票，你就是一个傻子，只会发呆"))

// Output: 不要**，你就是一个**，只会发呆
```

类似于上面的代码（两种代码类似），我们会使用一个敏感词列表（坏蛋、发票、傻子、傻大个、傻人），来对目标字符串进行检测与替换。比较适合于敏感词列表和待检测目标字符串都比较小的场景，否则性能会有较大影响。（正则替换和这个是类似的）

### 2）DFA

DFA 基本思想是基于状态转移来检索敏感词，只需要扫描一次待检测文本，就能对所有敏感词进行检测，所以效率比方案 1 高不少。

假设我们有以下 6 个敏感词需要检测：坏蛋、发票、傻子、傻大个、傻人。那么我们可以先把敏感词中有相同前缀的词组合成一个树形结构，不同前缀的词分属不同树形分支，以上述 6 个敏感词为例，可以初始化成如下 3 棵树：

![image](https://golang2.eddycjy.com/images/ch4/sensitive-tree.png)

把敏感词组成树形结构有什么好处呢？最大的好处就是可以减少检索次数，我们只需要遍历一次待检测文本，然后在敏感词库中检索出有没有该字符对应的子树就行了，如果没有相应的子树，说明当前检测的字符不在敏感词库中，则直接跳过继续检测下一个字符；如果有相应的子树，则接着检查下一个字符是不是前一个字符对应的子树的子节点，这样迭代下去，就能找出待检测文本中是否包含敏感词了。

我们以文本“不要发票，你就是一个傻子，只会发呆”为例，我们依次检测每个字符，因为前 2 个字符都不在敏感词库里，找不到相应的子树，所以直接跳过。当检测到“发”字时，发现敏感词库中有相应的子树，我们把它记为 tree-1，接着再搜索下一个字符“票”是不是子树 tree-1 的子节点，发现恰好是，接下来再判断“票”这个字符是不是叶子节点，如果是，则说明匹配到了一个敏感词了，在这里“票”这个字符刚好是 tree-1 的叶子节点，所以成功检索到了敏感词：“发票”。接着检测，“你就是一个”这几个字符都没有找到相应的子树，跳过。检测到“傻”字时，处理过程和前面的“发”是一样的，“傻子”的检测过程略过。

接着往后检测，“只会”也跳过。当检测到“发”字时，发现敏感词库中有相应的子树，我们把它记为 tree-3，接着再搜索下一个字符“呆”是不是子树 tree-3 的子节点，发现不是，因此这不是一个敏感词。

大家发现了没有，在我们的搜索过程中，我们只需要扫描一次被检测文本就行了，而且对于被检测文本中不存在的敏感词，如这个例子中的“坏蛋”、“傻大个”和“傻人”，我们完全不会扫描到，因此相比方案一效率大大提升了。

Go 中有一个库实现了该算法：github.com/antlinker/go-dirtyfilter。

### 3）基于朴素贝叶斯分类算法

贝叶斯分类是一类分类算法的总称，这类算法均以贝叶斯定理为基础，故统称为贝叶斯分类。而朴素朴素贝叶斯分类是贝叶斯分类中最简单，也是常见的一种分类方法。这是一种“半学习”形式的方法，它的准确性依赖于先验概率的准确性。

Go 中有一个库实现了该算法：github.com/jbrukh/bayesian。

### 小结

对于聊天室来说，每次的内容比较少，简单替换就可以满足大部分需求。实际中会涉及比较多的变种，比如敏感词中间加一些其他字符，有一个简单的方法是初始化一个无效字符库，比如：空格、*、#、@等字符，然后在检测文本前，先将待检测文本中的无效字符去除，这样的话被检测字符中就不存在这些无效字符了。

### 聊天室加上敏感词处理

聊天室一般发送的内容比较短，因此可以采用简单替换的方法。为了方便随时对敏感词列表进行修改，将敏感词存入配置文件中，通过 viper 库来处理配置文件。

由于不确定哪些地方可能需要用到配置文件中的内容，因此要求配置文件解析尽可能早的进行，同时方便其他地方进行引用或读取。因此进行代码重构，新创建一个包：global，用来存放配置文件和项目根目录等一些全局用的代码。

```go
// global/init.go

func init() {
	Init()
}

var RootDir string

var once = new(sync.Once)

func Init() {
	once.Do(func() {
		inferRootDir()
		initConfig()
	})
}

// inferRootDir 推断出项目根目录
func inferRootDir() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var infer func(d string) string
	infer = func(d string) string {
		// 这里要确保项目根目录下存在 template 目录
		if exists(d + "/template") {
			return d
		}

		return infer(filepath.Dir(d))
	}

	RootDir = infer(cwd)
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
```

以上代码核心要讲解的是 sync.Once。该类型的 Do 方法中的代码保证只会执行一次。这正好符合根目录推断和配置文件读取和解析。根据 Go 语言包的执行顺序，我们将相关初始化方法放在了单独的 Init 函数中，然后在 main.go 的 init 方法中调用它：

```go
func init() {
	global.Init()
}
```

为了支持敏感词的动态修改，及时生效，在 global 包中的 config.go 文件做相关处理：

```go
// global/config.go
var (
	SensitiveWords []string
)

func initConfig() {
	viper.SetConfigName("chatroom")
	viper.AddConfigPath(RootDir + "/config")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	SensitiveWords = viper.GetStringSlice("sensitive")

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		viper.ReadInConfig()

		SensitiveWords = viper.GetStringSlice("sensitive")
	})
}
```

其他配置项，如果不希望每次都通过 viper 调用获取，可以定义为 global 的包级变量，供其他地方使用。

配置文件放在项目根目录的 config/chatroom.yaml 中：

```yaml
sensitive:
  - 坏蛋
  - 坏人
  - 发票
  - 傻子
  - 傻大个
  - 傻人
```

在接收到用户发送的消息后，对敏感词进行处理。在 logic/user.go 的 ReceiveMessage 方法中增加对以下函数的调用：sendMsg.Content = FilterSensitive(sendMsg.Content)

```Go
// logic/sensitive.go
func FilterSensitive(content string) string {
	for _, word := range global.SensitiveWords {
		content = strings.ReplaceAll(content, word, "**")
	}

	return content
}
```

当用户发送：不要发票，你就是一个傻子，只会发呆。最终效果：

![image](https://golang2.eddycjy.com/images/ch4/sensitive.png)

## 4.7.3 离线消息处理（更确切说是最近的消息）

当用户不在线时，这期间发送的消息，是否需要存储，等下次上线时发送给 TA，这就是离线消息处理。

一般来说，聊天室不需要处理离线消息，而且我们的聊天室没有实现注册功能，同一个昵称不同时间可能被不同人使用，因此离线消息存储的意义不大。但有两种情况可以保存离线消息。

- 对某个用户的 @ 消息
- 最近发送的 10 条消息

我们聊天室要做到离线消息存储，需要解决一个问题：用户退出再登录，确保是同一个人，而不是另外一个人用了相同的昵称。但因为我们没有实现注册功能，于是这里需要对用户登录后进行一些处理。

### 1、正确识别同一个用户

目前聊天室虽然通过前端的 localStorage 存储了用户信息，方便记住和让同一个用户自动进入聊天室，但只要用户退出再登录，用户的 UID 就会变。为了正确识别同一个用户，我们需要保证同一个用户的 UID 和昵称都不变。

因为我们的聊天室不要求登录，为了更好的识别同一用户，同时避免恶意用户直接修改 localStorage 的数据，在用户进入聊天室时，为其生成一个 token，用来标识该用户，token 和用户昵称一起，存入 localStorage 中。

因为之前 localStorage 只是存储了用户昵称，所以需要进行修改。

- 之前的 nickname 改为 curUser，包含 nickname、uid 和 token 等用户信息；
- localStorage 中存入 curUser，通过 json 进行系列化后存入：localStorage.setItem(‘user’, JSON.stringify(data.user))
- 建立 WebSocket 连接时，除了之前的 nickname，额外传递 token：new WebSocket(“ws://“+host+”/ws?nickname="+this.curUser.nickname+”&token="+this.curUser.token);

为此，服务端要需要进行相关的修改。首先 User 结构增加两个字段：isNew bool 和 token string ，isNew 用来判断进来的用户是不是第一次加入聊天室。相应的，NewUser 方法修改为：

```go
func NewUser(conn *websocket.Conn, token, nickname, addr string) *User {
	user := &User{
		NickName:       nickname,
		Addr:           addr,
		EnterAt:        time.Now(),
		MessageChannel: make(chan *Message, 8),
		Token:          token,

		conn: conn,
	}

	if user.Token != "" {
		uid, err := parseTokenAndValidate(token, nickname)
		if err == nil {
			user.UID = uid
		}
	}

	if user.UID == 0 {
		user.UID = int(atomic.AddUint32(&globalUID, 1))
		user.Token = genToken(user.UID, user.NickName)
		user.isNew = true
	}

	return user
}
```

当没有传递 token 时，当做新用户处理，为用户生成一个 token：

```go
// logic/user.go
func genToken(uid int, nickname string) string {
	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)

	messageMAC := macSha256([]byte(message), []byte(secret))

	return fmt.Sprintf("%suid%d", base64.StdEncoding.EncodeToString(messageMAC), uid)
}

func macSha256(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}
```

token 的生成算法：

- 基于 HMAC-SHA256；
- nickname+secret+uid 构成待 hash 的字符串，记为：message
- 将 message 使用 HMAC-SHA256 计算 hash，记为：messageMAC
- 将 messageMAC 使用 base64 进行处理，记为：messageMACStr
- messageMACStr+“uid”+uid 就是 token

接着看看 token 的解析和校验，解析是为了得到 uid：

```go
// logic/user.go
func parseTokenAndValidate(token, nickname string) (int, error) {
	pos := strings.LastIndex(token, "uid")
	messageMAC, err := base64.StdEncoding.DecodeString(token[:pos])
	if err != nil {
		return 0, err
	}
	uid := cast.ToInt(token[pos+3:])

	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)

	ok := validateMAC([]byte(message), messageMAC, []byte(secret))
	if ok {
		return uid, nil
	}

	return 0, errors.New("token is illegal")
}

func validateMAC(message, messageMAC, secret []byte) bool {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
```

总体的思路就是按照生成 token 的方式，再得到一次 token，然后跟用户传递的 token 进行比较。因为 HMAC-SHA256 得到的结果是二进制的，因此相等比较使用了 hmac 包的 Equal 函数。这里大家可以借鉴下 uid 放入 token 中的技巧。

### 2、离线消息的实现

能够正确识别用户后，就可以来实现离线消息了。

在 logic 包中创建一个 offline.go 文件，创建 offlineProcessor 结构体对外提供一个单实例：OfflineProcessor。

```go
type offlineProcessor struct {
	n int

	// 保存所有用户最近的 n 条消息
	recentRing *ring.Ring

	// 保存某个用户离线消息（一样 n 条）
	userRing map[string]*ring.Ring
}

var OfflineProcessor = newOfflineProcessor()

func newOfflineProcessor() *offlineProcessor {
	n := viper.GetInt("offline-num")

	return &offlineProcessor{
		n:          n,
		recentRing: ring.New(n),
		userRing:   make(map[string]*ring.Ring),
	}
}
```

由于资源的限制，而且我们是直接将离线消息存在进程的内存中，因此不可能保留所有消息，而是保存最近的 n 条消息，其中 n 可以通过配置文件进行配置。这样的需求，标准库 container/ring 刚好满足。

**container/ring 详解**

这个包代码量很少，有效代码行数：87，包含注释和空格也就 141 行。因此，我们可以详细学习下它的实现。

从名字知晓，ring 实现了一个环形的链表，因此它没有起点或终点，指向环中任何元素的指针都可用作整个环的引用。空环表示为 nil 环指针。环的零值是一个包含一个元素，元素值是 nil 的环，如：

```go
var r ring.Ring 
fmt.Println(r.Len())	// Output: 1
fmt.Println(r.Value)  // Output: nil
```

但实际使用时，应该通过 New 函数来获得一个 Ring 的实例指针。

看看 Ring 结构体：

```go
type Ring struct {
	next, prev *Ring
	Value      interface{} // for use by client; untouched by this library
}
```

该结构体同时包含了 next 和 prev 字段，方便进行正反两个方向进行移动。我们可以通过 ring.New(n int) 函数得到一个 Ring 的实例指针，n 表示环的元素个数。

```go
func New(n int) *Ring {
	if n <= 0 {
		return nil
	}
	r := new(Ring)
	p := r
	for i := 1; i < n; i++ {
		p.next = &Ring{prev: p}
		p = p.next
	}
	p.next = r
	r.prev = p
	return r
}
```

New 函数一共创建了 n 个 Ring 实例指针，在 for 循环中，将这 n 个 Ring 实例指针链接起来。

为了更好的理解包中其他方法，我们使用一个图来表示。先构造一个 5 个元素的环，同时将每个元素的值分别设置为 1-5:

```go
r := ring.New(5)
n := r.Len()
for i := 1; i <= n; i++ {
  r.Value = i
  r = r.Next()
}
```

其中，Len 获得当前环的元素个数，时间复杂度是 O(n)。如图：

![image](https://golang2.eddycjy.com/images/ch4/ring-init.png)

当前 r 的值是 1（图中黑色箭头所指，这是为了表示方便，虚拟的）。分别看看 Ring 结构的方法。注意，移动相关的方法，都应该用返回值赋值给原 r，比如：r = r.Next()。

1）r.Next() 和 r.Prev()

这两个方法很简单。当前 r 代表值是 1 的元素，r.Next() 返回的 r 就代表值是 2 的元素；而 r.Prev() 返回的 r 则代表值是 5 的元素。

2）r.Move()

Next 和 Prev 一次只能移动一步（注意，可以理解为移动的是上图中黑色的箭头），而 Move 可以通过指定 n 来告知移动多少步，负数表示向后移动，正数表示向前移动。实际上，内部还是依赖于 Next 或 Prev 进行移动的。

这里要特别提醒一下，因为是环，所以参数 n 应该在 n % r.Len() 这个范围，否则做的是无用功。因为环的长度需要额外 O(n) 的时间计算，因此对 n 并没有做 n % r.Len() 的处理，传递的是多少就进行多少步移动，虽然最后结果跟 n % r.Len() 是一样的。

```go
func (r *Ring) Move(n int) *Ring {
	if r.next == nil {
		return r.init()
	}
	switch {
	case n < 0:
		for ; n < 0; n++ {
			r = r.prev
		}
	case n > 0:
		for ; n > 0; n-- {
			r = r.next
		}
	}
	return r
}
```

比如 r.Move(-2) 则把上图中的箭头移到了元素 4 处。

3）r.Do()

这是一个方便的遍历环的方法。该方法接收一个回调函数，函数的参数是当前环元素的 Value。该遍历是按照向前的方向进行的。因此，我们可以这样输出我们初始化的环：

```go
r.Do(func(value interface{}){
  fmt.Print(value.(int), " ")
})
```

输出：

```
1 2 3 4 5
```

4）r.Link() 和 r.Unlink()

这两个函数的作用相反，但接收参数不同。我们先看 r.Link()，向环中增加一个元素 6：

```go
nr := &ring.Ring{Value: 6}
or := r.Link(nr)
```

加上以上代码后，结果如图：

![image](https://golang2.eddycjy.com/images/ch4/ring-link.png)

类似的，r.Unlink 则是删除元素，参数 n 表示从下个元素起删除 n%r.Len() 个元素。

```go
dr := r.Unlink(3)
```

![image](https://golang2.eddycjy.com/images/ch4/ring-unlink.png)

从图中可以看出，环形链表被分成了两个，原来那个即 r， 从 1 开始，依次是 4、5，而被 unlink 掉的，即 dr，从 6 开始，依次是 2、3。

讲完 container/ring，我们回到离线消息上来。

**离线消息实现的两个核心方法：存和取**

先看离线消息如何存。

```go
func (o *offlineProcessor) Save(msg *Message) {
	if msg.Type != MsgTypeNormal {
		return
	}
	o.recentRing.Value = msg
	o.recentRing = o.recentRing.Next()

	for _, nickname := range msg.Ats {
		nickname = nickname[1:]
		var (
			r  *ring.Ring
			ok bool
		)
		if r, ok = o.userRing[nickname]; !ok {
			r = ring.New(o.n)
		}
		r.Value = msg
		o.userRing[nickname] = r.Next()
	}
}
```

- 根据 Ring 的使用方式，将用户消息直接存入 recentRing 中，并后移一个位置；
- 判断消息中是否有 @ 谁，需要单独为它保存一个消息列表；

这个方法在广播完消息后调用。

```go
case msg := <-b.messageChannel:
  // 给所有在线用户发送消息
  for _, user := range b.users {
    if user.UID == msg.User.UID {
      continue
    }
    user.MessageChannel <- msg
  }
  OfflineProcessor.Save(msg)
```

接着看用户离线后，再次进入聊天室取消息的实现。

```go
func (o *offlineProcessor) Send(user *User) {
	o.recentRing.Do(func(value interface{}) {
		if value != nil {
			user.MessageChannel <- value.(*Message)
		}
	})

	if user.isNew {
		return
	}

	if r, ok := o.userRing[user.NickName]; ok {
		r.Do(func(value interface{}) {
			if value != nil {
				user.MessageChannel <- value.(*Message)
			}
		})

		delete(o.userRing, user.NickName)
	}
}
```

首先遍历最近消息，发送给该用户。之后，如果不是新用户，查询是否有 @ 该用户的消息，有则发送给它，之后将这些消息删除。因为最近的消息是所有用户共享的，不能删除；@ 用户的消息是用户独有的，可以删除。

很显然，这个方法在用户进入聊天室后调用：

```go
case user := <-b.enteringChannel:
  // 新用户进入
  b.users[user.NickName] = user

  b.sendUserList()

  OfflineProcessor.Send(user)
```

细心的读者会发现以上处理方式，用户可能会收到重复的消息。的确如此。关于消息排重我们不做讲解了，大体思路是会为消息生成 ID，消息按时间排序，去重。实际业务中，去重更多会由客户端来做。

## 4.7.4 小结

一个产品，非核心功能是很多的，需要不断迭代。对于聊天室，肯定还有其他更多的功能可以开发，这就留给有兴趣的读者自己去探索、实现了。

在实现功能的过程中，把需要用到的库能够系统的学习一遍，你会掌握的很牢固，比如本节中的 container/ring，希望在以后的学习工作中，你能够做到。

