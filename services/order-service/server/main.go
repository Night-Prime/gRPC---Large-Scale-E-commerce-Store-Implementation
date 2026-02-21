package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	orderpb "e-comm/proto/gen/go/order/v1"
	paymentpb "e-comm/proto/gen/go/payment/v1"
	userpb "e-comm/proto/gen/go/user/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("Setting up the Order Service")
	fmt.Println("---------------------------------------->")

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	// connecting to the user service
	conn, err := grpc.NewClient("localhost:8081", opts...)
	if err != nil {
		log.Fatalf("Fail to Dial: %v", err)
	}

	defer conn.Close()

	// connecting to payment service
	paymentConn, err := grpc.NewClient("localhost:8082", opts...)
	if err != nil {
		log.Fatalf("Fail to Dial: %v", err)
	}

	defer paymentConn.Close()

	// ==================================================================>

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to E-comm Order Service\n")
	})

	http.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// verify user through rpc call to user service:
		client := userpb.NewUserServiceClient(conn)

		response, err := client.VerifyUser(ctx,
			&userpb.VerifyUserRequest{
				UserId:   "a23467585x-58686689-t5676d",
				Password: "Esther1012?",
			})
		if err != nil {
			log.Printf("Fail to Verify User: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		//then we create the order
		createdOrder := &orderpb.CreateOrderRequest{
			OrderId:      "167585ax566867bc",
			OrderProduct: "Escalade",
			IsPaid:       false,
			UserId:       "a23467585x-58686689-t5676d",
			OrderStatus:  orderpb.OrderStatus_ORDER_STATUS_UNSPECIFIED,
			OrderAmount:  "600000000",
		}

		var serverResponse *orderpb.CreateOrderResponse

		if response.IsVerified {
			fmt.Printf("User Verified Successfully\n")
			fmt.Println("---------------------------------------->")

			// Native execution of Create Order logic instead of gRPC
			fmt.Println("Creating an Order Natively")
			fmt.Printf("Processing Order Request: %v\n", createdOrder)

			serverResponse = &orderpb.CreateOrderResponse{
				OrderStatus: orderpb.OrderStatus_ORDER_STATUS_PENDING,
			}
			fmt.Printf("Order natively created with status: %v\n", serverResponse.OrderStatus)
		}

		if serverResponse.OrderStatus == orderpb.OrderStatus_ORDER_STATUS_PENDING {
			// charge the user through rpc call to payment service:
			paymentClient := paymentpb.NewPaymentServiceClient(paymentConn)

			response, err := paymentClient.Charge(ctx, &paymentpb.ChargeRequest{
				UserId:      "a23467585x-58686689-t5676d",
				OrderId:     "167585ax566867bc",
				OrderAmount: "600000000",
			})
			if err != nil {
				log.Printf("Fail to Charge User: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if response.IsPaid {
				fmt.Printf("User Charged Successfully \n")
				fmt.Println("---------------------------------------->")

				createdOrder.OrderStatus = orderpb.OrderStatus_ORDER_STATUS_SUCCESS
			}
			fmt.Printf("Response From Payment Service: %v\n", response)
		} else {
			createdOrder.OrderStatus = orderpb.OrderStatus_ORDER_STATUS_FAILED
		}

		fmt.Printf("Response From User Service: %v\n", response)

		orderBytes, err := json.Marshal(createdOrder)
		if err != nil {
			log.Printf("Fail to Marshal Order: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s", orderBytes)
	})

	fmt.Println("Order Service is listening on :8080")
	httpErr := http.ListenAndServe(":8080", nil)
	if httpErr != nil {
		log.Fatal("HTTP Server Error: ", httpErr)
	}
}
