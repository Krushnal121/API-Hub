package main

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "github.com/Krushnal121/API-Hub/gRPC/Go/api/gen/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the gRPC server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client from the generated gRPC code
	client := v1.NewGreeterClient(conn)

	// Prepare the request
	req := &v1.HelloRequest{Name: "Krushnal"}

	// Call the RPC
	resp, err := client.SayHello(ctx, req)
	if err != nil {
		log.Fatalf("Error calling SayHello: %v", err)
	}

	// Print the result
	fmt.Printf("Response from server: %s\n", resp.GetMessage())
}
