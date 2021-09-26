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

