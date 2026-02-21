package main

import (
	"fmt"
	"log"
	"net"

	paymentservice "e-comm/payment-service"
	paymentpb "e-comm/proto/gen/go/payment/v1"

	"google.golang.org/grpc"
)

type PaymentService struct {
	paymentservice.PaymentService
}

func main() {
	fmt.Println("Setting up Payment Service")
	fmt.Println("------------------------------------------------>")

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	paymentServer := &PaymentService{}

	grpcServer := grpc.NewServer()

	paymentpb.RegisterPaymentServiceServer(grpcServer, paymentServer)

	fmt.Println("Payment Service is running on port 8082")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
