package main

import (
	"fmt"
	"log"
	"net"

	userpb "e-comm/proto/gen/go/user/v1"
	userservice "e-comm/user-service"

	"google.golang.org/grpc"
)

type UserService struct {
	userservice.UserService
}

func main() {
	fmt.Println("Setting up User Service")
	fmt.Println("------------------------------------------------>")

	lis, err := net.Listen("tcp", ":8081")

	if err != nil {
		log.Fatal("TCP Server Error: ", err)
	}

	userServer := &UserService{}

	grpcServer := grpc.NewServer()

	userpb.RegisterUserServiceServer(grpcServer, userServer)

	fmt.Println("User Service is running on port 8081")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
