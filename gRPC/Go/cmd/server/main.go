package main

import (
	"log/slog"
	"net"
	"os"

	v1 "github.com/Krushnal121/API-Hub/gRPC/Go/api/gen/v1"
	"github.com/Krushnal121/API-Hub/gRPC/Go/internal/greeter"
	"google.golang.org/grpc"
)

func main() {

	// Start listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting gRPC server", "port", 50051)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register service
	v1.RegisterGreeterServer(grpcServer, &service.GreeterService{})

	// Serve
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
	}

}
