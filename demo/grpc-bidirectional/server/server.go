package main

import (
	"flag"
	"fmt"
	pb "grpc-bidirectional/proto"
	"io"
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

func (g *GreeterServer) SayRoute(stream pb.Greeter_SayRouteServer) error {
	i := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Printf("Receive: %v\n", req.Name)
		err = stream.Send(&pb.HelloReply{Message: "Server Say: " + strconv.Itoa(i)})
		if err != nil {
			return err
		}
		i++
	}
}

func main() {
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &GreeterServer{})
	listen_socket, _ := net.Listen("tcp", host+":"+port)
	err := server.Serve(listen_socket)
	if err != nil {
		log.Fatalln(err)
	}
}
