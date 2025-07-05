package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	v1 "github.com/Krushnal121/API-Hub/gRPC/Go/api/gen/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var (
	conn *grpc.ClientConn
	once sync.Once
)

// Connection pool/reuse
func getConnection() *grpc.ClientConn {  
	once.Do(func() {
		var err error
		var creds credentials.TransportCredentials

		// Check if certificates exist
		if _, err := os.Stat("certs/server.crt"); err == nil {
			// Load the server's certificate
			certificate, err := os.ReadFile("certs/server.crt")
			if err != nil {
				log.Printf("could not read server certificate: %v", err)
				creds = insecure.NewCredentials()
			} else {
				// Create a certificate pool and add the server's certificate
				certPool := x509.NewCertPool()
				if !certPool.AppendCertsFromPEM(certificate) {
					log.Printf("failed to add server's certificate to the certificate pool")
					creds = insecure.NewCredentials()
				} else {
					// Create the TLS credentials
					config := &tls.Config{
						RootCAs: certPool,
					}
					creds = credentials.NewTLS(config)
				}
			}
		} else {
			creds = insecure.NewCredentials()
		}

		conn, err = grpc.NewClient("localhost:50051",
			grpc.WithTransportCredentials(creds),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			}),
		)
		if err != nil {
			log.Fatal(err)
		}
	})
	return conn
}

func testUnaryRPC(client v1.GreeterClient) {
	fmt.Println("\n=== Testing Unary RPC ===")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
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
	fmt.Printf("Timestamp: %d\n", resp.GetTimestamp())
	fmt.Printf("Server Info: %s\n", resp.GetServerInfo())
}

func testServerStreaming(client v1.GreeterClient) {
	fmt.Println("\n=== Testing Server Streaming ===")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req := &v1.HelloRequest{
		Name: "Krushnal",
		Age:  21,
	}

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

		fmt.Printf("Stream message: %s (Info: %s)\n", resp.GetMessage(), resp.GetServerInfo())
	}
}

func testClientStreaming(client v1.GreeterClient) {
	fmt.Println("\n=== Testing Client Streaming ===")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stream, err := client.SayHelloBulk(ctx)
	if err != nil {
		log.Printf("Error calling SayHelloBulk: %v", err)
		return
	}

	// Send multiple requests
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
		req := &v1.HelloRequest{
			Name: person.name,
			Age:  person.age,
		}

		if err := stream.Send(req); err != nil {
			log.Printf("Error sending to stream: %v", err)
			return
		}
		fmt.Printf("Sent: %s (age %d)\n", person.name, person.age)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	stream, err := client.SayHelloChat(ctx)
	if err != nil {
		log.Printf("Error calling SayHelloChat: %v", err)
		return
	}

	// Send messages in a goroutine
	go func() {
		messages := []string{"Emma", "Frank", "Grace"}
		for _, name := range messages {
			req := &v1.HelloRequest{
				Name: name,
				Age:  20 + int32(len(name)), // Simple age calculation
			}

			if err := stream.Send(req); err != nil {
				log.Printf("Error sending chat message: %v", err)
				return
			}
			fmt.Printf("Sent chat: %s\n", name)
			time.Sleep(1 * time.Second)
		}
		stream.CloseSend()
	}()

	// Receive messages
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error receiving chat: %v", err)
			break
		}

		fmt.Printf("Chat response: %s\n", resp.GetMessage())
	}
}

func main() {
	// Connect to the gRPC server
	conn := getConnection()
	defer conn.Close()

	// Create a client
	client := v1.NewGreeterClient(conn)

	// Test all RPC types
	testUnaryRPC(client)
	testServerStreaming(client)
	testClientStreaming(client)
	testBidirectionalStreaming(client)

	fmt.Println("\n=== All tests completed ===")
}
