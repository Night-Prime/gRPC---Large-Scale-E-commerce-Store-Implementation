package userservice

import (
	"context"
	"fmt"

	userpb "e-comm/proto/gen/go/user/v1"
)

type UserService struct {
	userpb.UnimplementedUserServiceServer
}

func (s *UserService) VerifyUser(ctx context.Context, req *userpb.VerifyUserRequest) (*userpb.VerifyUserResponse, error) {
	fmt.Printf("Verifying User: %v\n", req.UserId)
	return &userpb.VerifyUserResponse{
		IsVerified: true,
	}, nil
}
