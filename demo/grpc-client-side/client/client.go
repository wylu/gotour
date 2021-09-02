package main

import (
	"context"
	"flag"
	"log"
	"strconv"

	pb "grpc-client-side/proto"

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
	client := pb.NewGreeterClient(conn)
	stream, err := client.SayRecord(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < 6; i++ {
		_ = stream.Send(&pb.HelloRequest{Name: "hello " + strconv.Itoa(i)})
	}
	resp, _ := stream.CloseAndRecv()

	log.Printf("Receive: %v\n", resp)
}
