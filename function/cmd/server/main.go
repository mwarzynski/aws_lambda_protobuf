package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"mwarzynski/aws-grpc-lambda/api/hello"
)

type gRPCServer struct{}

func (s gRPCServer) SayHello(context.Context, *hello.HelloRequest) (*hello.HelloReply, error) {
	return &hello.HelloReply{Message: "Hello from Golang and gRPC!!1"}, nil
}
func (gRPCServer) mustEmbedUnimplementedGreeterServer() {}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)

	x := gRPCServer{}

	hello.RegisterGreeterServer(s, x)
	s.Serve(lis)
}
