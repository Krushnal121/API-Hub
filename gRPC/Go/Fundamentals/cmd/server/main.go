package main

import (
	"net"
	"log"
	"google.golang.org/grpc"

	v1 "github.com/Krushnal121/API-Hub/gRPC/Go/api/gen/v1"
	"github.com/Krushnal121/API-Hub/gRPC/Go/internal/greeter"
)

func main() {
	// Listen on TCP port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create gRPC server with no special options
	grpcServer := grpc.NewServer()

	// Register the Greeter service
	v1.RegisterGreeterServer(grpcServer, &service.GreeterService{})

	log.Println("Starting gRPC server on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
