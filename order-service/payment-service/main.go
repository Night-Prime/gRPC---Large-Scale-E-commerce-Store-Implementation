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
	fmt.Println("Setting up Payment Service")
	fmt.Println("------------------------------------------------>")

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	orderServer := &OrderService{}

	grpcServer := grpc.NewServer()

	orderpb.RegisterOrderServiceServer(grpcServer, orderServer)

	fmt.Println("Payment Service is running on port 8082")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
