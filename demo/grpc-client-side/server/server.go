package main

import (
	"flag"
	"io"
	"log"
	"net"

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

type GreeterServer struct{}

func (g *GreeterServer) SayRecord(stream pb.Greeter_SayRecordServer) error {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.HelloReply{Message: "Record EOF"})
		}
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Receive: %v\n", resp)
	}
}

func main() {
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &GreeterServer{})
	listen_socket, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalln(err)
	}
	server.Serve(listen_socket)
}
