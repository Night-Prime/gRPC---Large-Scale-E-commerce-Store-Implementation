package paymentservice

import (
	"context"
	"fmt"

	paymentpb "e-comm/proto/gen/go/payment/v1"
)

type PaymentService struct {
	paymentpb.UnimplementedPaymentServiceServer
}

func (s *PaymentService) Charge(ctx context.Context, req *paymentpb.ChargeRequest) (*paymentpb.ChargeResponse, error) {
	fmt.Printf("Charging User: %v\n", req.UserId)
	return &paymentpb.ChargeResponse{
		IsPaid: true,
	}, nil
}
