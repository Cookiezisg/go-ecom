package main

import (
	"flag"
	"fmt"
	"net"

	userpb "ecommerce-system/api/user/v1"
	usersvc "ecommerce-system/internal/service/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var listenAddr = flag.String("listen", "0.0.0.0:8000", "gRPC listen address")

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	userpb.RegisterUserServiceServer(s, usersvc.NewUserService())

	// 开发调试：开启反射，方便 grpcurl 调试
	reflection.Register(s)

	fmt.Printf("user-service listening on %s\n", *listenAddr)
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}
