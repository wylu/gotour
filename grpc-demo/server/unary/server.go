package main

import (
	"context"

	pb "github.com/wylu/gotour/tour/grpc-demo/proto"
)

type GreeterServer struct{}

func (g *GreeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello World!"}, nil
}

func main() {

}
