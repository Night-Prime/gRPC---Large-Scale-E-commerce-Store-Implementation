package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	orderpb "e-comm/order-service/gen/go/order/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("Setting up the Order Service")
	fmt.Println("---------------------------------------->")

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial("localhost:8081", opts...)
	if err != nil {
		log.Fatalf("Fail to Dial: %v", err)
	}

	defer conn.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to E-comm Order Service\n")
	})

	http.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		// verify user through rpc call to user service:
		client := orderpb.NewOrderServiceClient(conn)
		stream, err := client.CreateOrder(context.Background())
		if err != nil {
			log.Printf("Fail to Create Order: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		response, err := client.VerifyUser(context.Background(),
			&orderpb.VerifyUserRequest{
				UserId:   "a23467585x-58686689-t5676d",
				Password: "Esther1012?",
			})
		if err != nil {
			log.Printf("Fail to Verify User: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if response.IsVerified {
			fmt.Fprintf(w, "User Verified Successfully\n")
			fmt.Println("---------------------------------------->")

			if err := stream.Send(&orderpb.CreateOrderRequest{
				OrderId:      "167585ax566867bc",
				OrderProduct: "New Girlfriend",
				IsPaid:       true,
				UserId:       "a23467585x-58686689-t5676d",
				OrderStatus:  orderpb.OrderStatus_ORDER_STATUS_UNSPECIFIED,
				OrderAmount:  "60000",
			}); err != nil {
				log.Printf("Fail to Send Order: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			streamResp, err := stream.Recv()
			if err != nil {
				log.Printf("Fail to Receive Order: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "Streaming response: %v\n", streamResp)
		}

		fmt.Fprintf(w, "Response From User Service: %v\n", response)
	})

	fmt.Println("Order Service is listening on :8080")
	httpErr := http.ListenAndServe(":8080", nil)
	if httpErr != nil {
		log.Fatal("HTTP Server Error: ", httpErr)
	}
}
