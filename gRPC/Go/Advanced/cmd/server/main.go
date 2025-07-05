package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/Krushnal121/API-Hub/gRPC/Go/api/gen/v1"
	"github.com/Krushnal121/API-Hub/gRPC/Go/internal/greeter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Logging middleware
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)

	slog.Info("gRPC call completed",
		"method", info.FullMethod,
		"duration", time.Since(start),
		"error", err != nil,
	)

	return resp, err
}

// Authentication middleware (simple example)
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip auth for health checks or specific methods
	if info.FullMethod == "/greeter.v1.Greeter/SayHello" {
		// Simple token validation (in production, use proper JWT validation)
		// For demo purposes, we'll skip actual validation
		return handler(ctx, req)
	}
	return handler(ctx, req)
}

func main() {
	// Configure logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Listen on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	// Configure server options
	var opts []grpc.ServerOption

	// Add TLS if certificates exist
	if _, err := os.Stat("certs/server.crt"); err == nil {
		cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
		if err != nil {
			slog.Warn("Failed to load TLS certificates", "error", err)
		} else {
			creds := credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{cert},
			})
			opts = append(opts, grpc.Creds(creds))
			slog.Info("TLS enabled")
		}
	}

	// Add middleware
	opts = append(opts,
		grpc.UnaryInterceptor(chainUnaryInterceptors(
			loggingInterceptor,
			authInterceptor,
		)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 15 * time.Second,
			MaxConnectionAge:  30 * time.Second,
			Time:              5 * time.Second,
			Timeout:           1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	// Create gRPC server
	grpcServer := grpc.NewServer(opts...)

	// Register service
	v1.RegisterGreeterServer(grpcServer, &service.GreeterService{})

	slog.Info("Starting gRPC server", "port", 50051)

	// Graceful shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	grpcServer.GracefulStop()
	slog.Info("Server stopped")
}

// Helper function to chain multiple interceptors
func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptor(currentCtx, currentReq, info, next)
			}
		}
		return chain(ctx, req)
	}
}