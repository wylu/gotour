package main

import (
	"context"
	"flag"
	"fmt"
	pb "grpc-server-side/proto"
	"io"
	"log"

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

func main() {
	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Client Connect: %v:%v\n", host, port)
	client := pb.NewGreeterClient(conn)
	stream, err := client.SayList(context.Background(), &pb.HelloRequest{Name: "wylu"})
	if err != nil {
		log.Fatalln(err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("Received: %s\n", resp)
	}
}
