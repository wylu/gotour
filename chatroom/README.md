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
