package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "grpc-unary/proto"

	"google.golang.org/grpc"
)

var host string
var port string

func init() {
	flag.StringVar(&host, "host", "localhost", "IP")
	flag.StringVar(&port, "port", "8000", "端口")
	flag.Parse()
}

func main() {
	fmt.Printf("connect %v:%v\n", host, port)
	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)
	resp, _ := client.SayHello(context.Background(), &pb.HelloRequest{Name: "wylu"})
	fmt.Printf("client.SayHello resp: %v\n", resp.Message)
}
