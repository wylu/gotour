package main

import (
	"context"
	"flag"
	pb "grpc-bidirectional/proto"
	"io"
	"log"
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

func main() {
	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	client := pb.NewGreeterClient(conn)

	stream, err := client.SayRoute(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	for i := 0; i < 6; i++ {
		err := stream.Send(&pb.HelloRequest{Name: "Hello " + strconv.Itoa(i)})
		if err != nil {
			log.Fatalln(err)
		}

		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Receive: %v\n", resp)
	}

	err = stream.CloseSend()
	if err != nil {
		log.Fatalln(err)
	}
}
