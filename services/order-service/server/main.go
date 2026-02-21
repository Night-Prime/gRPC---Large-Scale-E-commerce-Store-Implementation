package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	orderpb "e-comm/proto/gen/go/order/v1"
	paymentpb "e-comm/proto/gen/go/payment/v1"
	userpb "e-comm/proto/gen/go/user/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
	userClient    userpb.UserServiceClient
	paymentClient paymentpb.PaymentServiceClient
}

func (s *OrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	fmt.Println("Creating an Order Natively")
	fmt.Printf("Processing Order Request: %v\n", req)

	// Verify User
	verifyCtx, verifyCancel := context.WithTimeout(ctx, 30*time.Second)
	defer verifyCancel()

	userRes, err := s.userClient.VerifyUser(verifyCtx, &userpb.VerifyUserRequest{
		Token: req.UserId, // Client will pass JWT via user_id field
	})
	if err != nil {
		log.Printf("Fail to Verify User: %v", err)
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	serverResponse := &orderpb.CreateOrderResponse{
		OrderStatus: orderpb.OrderStatus_ORDER_STATUS_FAILED,
	}

	if userRes.IsVerified {
		fmt.Printf("User Verified Successfully\n")
		fmt.Println("---------------------------------------->")

		serverResponse.OrderStatus = orderpb.OrderStatus_ORDER_STATUS_PENDING
		fmt.Printf("Order natively created with status: %v\n", serverResponse.OrderStatus)

		// Charge User
		chargeCtx, chargeCancel := context.WithTimeout(ctx, 30*time.Second)
		defer chargeCancel()

		paymentRes, err := s.paymentClient.Charge(chargeCtx, &paymentpb.ChargeRequest{
			UserId:      userRes.UserId,
			OrderId:     req.OrderId,
			OrderAmount: req.OrderAmount,
		})
		if err != nil {
			log.Printf("Fail to Charge User: %v", err)
			return nil, fmt.Errorf("failed to charge user: %w", err)
		}

		if paymentRes.IsPaid {
			fmt.Printf("User Charged Successfully \n")
			fmt.Println("---------------------------------------->")
			serverResponse.OrderStatus = orderpb.OrderStatus_ORDER_STATUS_SUCCESS
		}
	} else {
		log.Println("User not verified.")
	}

	fmt.Printf("Final Order Status: %v\n", serverResponse.OrderStatus)
	return serverResponse, nil
}

func main() {
	fmt.Println("Setting up the Order Service (gRPC)")
	fmt.Println("---------------------------------------->")

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// connecting to the user service
	userConn, err := grpc.NewClient("localhost:8081", opts...)
	if err != nil {
		log.Fatalf("Fail to Dial User Service: %v", err)
	}
	defer userConn.Close()

	// connecting to payment service
	paymentConn, err := grpc.NewClient("localhost:8082", opts...)
	if err != nil {
		log.Fatalf("Fail to Dial Payment Service: %v", err)
	}
	defer paymentConn.Close()

	// Setup Server
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	orderServer := &OrderService{
		userClient:    userpb.NewUserServiceClient(userConn),
		paymentClient: paymentpb.NewPaymentServiceClient(paymentConn),
	}

	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, orderServer)

	fmt.Println("Order Service is running via gRPC on :8080")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
