package userservice

import (
	"context"
	"fmt"
	"os"
	"time"

	userpb "e-comm/proto/gen/go/user/v1"
	"e-comm/user-service/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type UserService struct {
	userpb.UnimplementedUserServiceServer
}

func generateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString(jwtSecretKey)
}

// Verify JWT token
func verifyToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["user_id"].(string), nil
	}
	return "", fmt.Errorf("invalid token")
}

func (s *UserService) Signup(ctx context.Context, req *userpb.SignupRequest) (*userpb.SignupResponse, error) {
	fmt.Printf("Attempting signup for email: %v\n", req.Email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to hash password")
	}

	userID := uuid.New().String()

	user := config.User{
		ID:       userID,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	result := config.DB.Create(&user)
	if result.Error != nil {
		return nil, status.Errorf(codes.AlreadyExists, "Email may already exist: %v", result.Error)
	}

	token, err := generateToken(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate token")
	}

	fmt.Printf("User %v signed up successfully\n", req.Email)

	return &userpb.SignupResponse{
		UserId: userID,
		Token:  token,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	fmt.Printf("Attempting login for email: %v\n", req.Email)

	var user config.User
	result := config.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid password")
	}

	token, err := generateToken(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate token")
	}

	fmt.Printf("User %v logged in successfully\n", req.Email)

	return &userpb.LoginResponse{
		Token: token,
	}, nil
}

func (s *UserService) VerifyUser(ctx context.Context, req *userpb.VerifyUserRequest) (*userpb.VerifyUserResponse, error) {
	fmt.Printf("Verifying User via Token...\n")

	userID, err := verifyToken(req.Token)
	if err != nil {
		fmt.Printf("Token Verification Failed: %v\n", err)
		return &userpb.VerifyUserResponse{
			IsVerified: false,
			UserId:     "",
		}, nil
	}

	fmt.Printf("Token verified successfully for user: %v\n", userID)
	return &userpb.VerifyUserResponse{
		IsVerified: true,
		UserId:     userID,
	}, nil
}
