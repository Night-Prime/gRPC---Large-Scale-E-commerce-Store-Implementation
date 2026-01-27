package main

import (
	"fmt"
	"log"
	"net"

	orderservice "e-comm/order-service"
	orderpb "e-comm/order-service/gen/go/order/v1"

	"google.golang.org/grpc"
)

type OrderService struct {
	orderservice.OrderService
}

func main() {
	fmt.Println("Setting up User Service")
	fmt.Println("------------------------------------------------>")

	lis, err := net.Listen("tcp", ":8081")

	if err != nil {
		log.Fatal("TCP Server Error: ", err)
	}

	orderServer := &OrderService{}

	grpcServer := grpc.NewServer()

	orderpb.RegisterOrderServiceServer(grpcServer, orderServer)

	fmt.Println("User Service is running on port 8081")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
