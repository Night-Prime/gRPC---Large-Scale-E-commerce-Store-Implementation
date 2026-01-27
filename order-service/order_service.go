package orderservice

import (
	"context"
	"fmt"
	"io"

	orderpb "e-comm/order-service/gen/go/order/v1"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
}

func (s *OrderService) VerifyUser(ctx context.Context, req *orderpb.VerifyUserRequest) (*orderpb.VerifyUserResponse, error) {
	fmt.Printf("Verifying User: %v\n", req.UserId)
	return &orderpb.VerifyUserResponse{
		IsVerified: true,
	}, nil
}

func (s *OrderService) CreateOrder(stream orderpb.OrderService_CreateOrderServer) error {
	fmt.Printf("Creating an Order")
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	fmt.Printf("Received the Stream Request to process: %v \n", req)
	fmt.Println("---------------------------------------->")

	if streamErr := stream.Send(&orderpb.CreateOrderResponse{
		OrderStatus: orderpb.OrderStatus_ORDER_STATUS_PENDING,
	}); streamErr != nil {
		return streamErr
	}

	return nil
}
