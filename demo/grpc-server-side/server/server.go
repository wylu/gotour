package main

import (
	"flag"
	"fmt"
	pb "grpc-server-side/proto"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
)

var (
	host string
	port string
)

func init() {
	flag.StringVar(&host, "host", "localhost", "Hostname to connect to")
	flag.StringVar(&port, "port", "8000", "Port to connect to")
	flag.Parse()
}

type GreeterServer struct{}

// 在 Server 端，主要留意 stream.Send 方法，通过阅读源码，可得知是 protoc 在生成时，
// 根据定义生成了各式各样符合标准的接口方法。最终再统一调度内部的 SendMsg 方法，
// 该方法涉及以下过程:
//
// 1.消息体（对象）序列化。
// 2.压缩序列化后的消息体。
// 3.对正在传输的消息体增加 5 个字节的 header（标志位）。
// 4.判断压缩 + 序列化后的消息体总字节长度是否大于预设的 maxSendMessageSize
//（预设值为 math.MaxInt32），若超出则提示错误。
// 5.写入给流的数据集。
func (g *GreeterServer) SayList(r *pb.HelloRequest, stream pb.Greeter_SayListServer) error {
	fmt.Printf("Receive: %v\n", r.Name)
	for i := 0; i < 6; i++ {
		_ = stream.Send(&pb.HelloReply{Message: "Hello.List" + strconv.Itoa(i)})
	}
	return nil
}

// 服务器端流式 RPC，也就是是单向流，并代指 Server 为 Stream，Client 为普通的
// 一元 RPC 请求。
//
// 简单来讲就是客户端发起一次普通的 RPC 请求，服务端通过流式响应多次发送数据集，
// 客户端 Recv 接收数据集。

func main() {
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &GreeterServer{})
	listen_socket, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Server Listen: %v:%v\n", host, port)
	server.Serve(listen_socket)
}
