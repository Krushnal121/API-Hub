package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	v1 "github.com/Krushnal121/API-Hub/gRPC/Go/api/gen/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Dial server with insecure credentials
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := v1.NewGreeterClient(conn)

	testUnaryRPC(client)
	testServerStreaming(client)
	testClientStreaming(client)
	testBidirectionalStreaming(client)

	fmt.Println("\n=== All tests completed ===")
}

func testUnaryRPC(client v1.GreeterClient) {
	fmt.Println("\n=== Testing Unary RPC ===")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &v1.HelloRequest{
		Name:      "Krushnal Patil",
		Age:       21,
		Interests: []string{"gRPC", "Go", "Microservices"},
	}

	resp, err := client.SayHello(ctx, req)
	if err != nil {
		log.Printf("Error calling SayHello: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.GetMessage())
}

func testServerStreaming(client v1.GreeterClient) {
	fmt.Println("\n=== Testing Server Streaming ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &v1.HelloRequest{Name: "Krushnal", Age: 21}

	stream, err := client.SayHelloStream(ctx, req)
	if err != nil {
		log.Printf("Error calling SayHelloStream: %v", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error receiving stream: %v", err)
			break
		}
		fmt.Printf("Stream message: %s\n", resp.GetMessage())
	}
}

func testClientStreaming(client v1.GreeterClient) {
	fmt.Println("\n=== Testing Client Streaming ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.SayHelloBulk(ctx)
	if err != nil {
		log.Printf("Error calling SayHelloBulk: %v", err)
		return
	}

	names := []struct {
		name string
		age  int32
	}{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 35},
		{"Diana", 28},
	}

	for _, person := range names {
		req := &v1.HelloRequest{Name: person.name, Age: person.age}
		if err := stream.Send(req); err != nil {
			log.Printf("Error sending to stream: %v", err)
			return
		}
		fmt.Printf("Sent: %s\n", person.name)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Error closing stream: %v", err)
		return
	}

	fmt.Printf("Bulk response: %s\n", resp.GetMessage())
}

func testBidirectionalStreaming(client v1.GreeterClient) {
	fmt.Println("\n=== Testing Bidirectional Streaming ===")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	stream, err := client.SayHelloChat(ctx)
	if err != nil {
		log.Printf("Error calling SayHelloChat: %v", err)
		return
	}

	go func() {
		names := []string{"Emma", "Frank", "Grace"}
		for _, name := range names {
			req := &v1.HelloRequest{Name: name, Age: 20 + int32(len(name))}
			if err := stream.Send(req); err != nil {
				log.Printf("Send error: %v", err)
				return
			}
			fmt.Printf("Sent chat: %s\n", name)
			time.Sleep(1 * time.Second)
		}
		stream.CloseSend()
	}()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Receive error: %v", err)
			break
		}
		fmt.Printf("Chat response: %s\n", resp.GetMessage())
	}
}
