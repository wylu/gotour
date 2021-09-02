package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	pb "grpc-unary/proto"

	"google.golang.org/grpc"
)

type GreeterServer struct{}

func (g *GreeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("Receive Message: %v\n", (*req).Name)
	return &pb.HelloReply{Message: "Hello World!"}, nil
}

var host string
var port string

func init() {
	flag.StringVar(&host, "host", "localhost", "监听地址")
	flag.StringVar(&port, "port", "8000", "监听端口")
	flag.Parse()
}

func main() {
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &GreeterServer{})
	listen_socket, _ := net.Listen("tcp", ":"+port)
	fmt.Printf("Listen: %v:%v\n", host, port)
	server.Serve(listen_socket)
}
