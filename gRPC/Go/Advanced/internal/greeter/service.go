package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	v1 "github.com/Krushnal121/API-Hub/gRPC/Go/api/gen/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GreeterService struct {
	v1.UnimplementedGreeterServer
}

// Unary RPC implementation
func (s *GreeterService) SayHello(ctx context.Context, req *v1.HelloRequest) (*v1.HelloReply, error) {
	// Input validation
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	if req.Age < 0 {
		return nil, status.Error(codes.InvalidArgument, "age must be non-negative")
	}

	message := fmt.Sprintf("Hello, %s! You are %d years old.", req.Name, req.Age)
	
	if len(req.Interests) > 0 {
		message += fmt.Sprintf(" Your interests: %s", strings.Join(req.Interests, ", "))
	}

	slog.Info("Received SayHello request", 
		"name", req.Name, 
		"age", req.Age,
		"interests", req.Interests)

	return &v1.HelloReply{
		Message:    message,
		Timestamp:  time.Now().Unix(),
		ServerInfo: "gRPC-Server-v1.0",
	}, nil
}

// Server streaming RPC implementation
func (s *GreeterService) SayHelloStream(req *v1.HelloRequest, stream v1.Greeter_SayHelloStreamServer) error {
	if req.Name == "" {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	slog.Info("Starting stream for", "name", req.Name)

	// Send multiple greetings
	greetings := []string{
		fmt.Sprintf("Hello %s! Welcome!", req.Name),
		fmt.Sprintf("Nice to meet you, %s!", req.Name),
		fmt.Sprintf("Hope you're having a great day, %s!", req.Name),
		fmt.Sprintf("Thanks for trying gRPC, %s!", req.Name),
		fmt.Sprintf("Goodbye for now, %s!", req.Name),
	}

	for i, greeting := range greetings {
		reply := &v1.HelloReply{
			Message:    greeting,
			Timestamp:  time.Now().Unix(),
			ServerInfo: fmt.Sprintf("Stream message %d/5", i+1),
		}

		if err := stream.Send(reply); err != nil {
			slog.Error("Failed to send stream message", "error", err)
			return status.Error(codes.Internal, "failed to send message")
		}

		// Simulate processing time
		time.Sleep(500 * time.Millisecond)
	}

	slog.Info("Completed stream for", "name", req.Name)
	return nil
}

// Client streaming RPC implementation
func (s *GreeterService) SayHelloBulk(stream v1.Greeter_SayHelloBulkServer) error {
	var names []string
	var totalAge int32

	slog.Info("Starting bulk greeting processing")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			break
		}
		if err != nil {
			slog.Error("Error receiving from client stream", "error", err)
			return status.Error(codes.Internal, "failed to receive message")
		}

		if req.Name != "" {
			names = append(names, req.Name)
			totalAge += req.Age
		}
	}

	if len(names) == 0 {
		return status.Error(codes.InvalidArgument, "no valid names received")
	}

	message := fmt.Sprintf("Bulk greeting processed for %d people: %s. Average age: %.1f", 
		len(names), 
		strings.Join(names, ", "),
		float64(totalAge)/float64(len(names)))

	reply := &v1.HelloReply{
		Message:    message,
		Timestamp:  time.Now().Unix(),
		ServerInfo: fmt.Sprintf("Processed %d requests", len(names)),
	}

	slog.Info("Completed bulk processing", "count", len(names))
	return stream.SendAndClose(reply)
}

// Bidirectional streaming RPC implementation
func (s *GreeterService) SayHelloChat(stream v1.Greeter_SayHelloChatServer) error {
	slog.Info("Starting chat session")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			slog.Info("Chat session ended")
			return nil
		}
		if err != nil {
			slog.Error("Error in chat stream", "error", err)
			return status.Error(codes.Internal, "chat error")
		}

		// Echo back with a chat-like response
		responses := []string{
			fmt.Sprintf("Hi %s! How can I help you today?", req.Name),
			fmt.Sprintf("That's interesting, %s! Tell me more.", req.Name),
			fmt.Sprintf("I see, %s. Thanks for sharing!", req.Name),
		}

		// Pick a response based on message count (simple logic)
		responseIndex := len(req.Name) % len(responses)
		
		reply := &v1.HelloReply{
			Message:    responses[responseIndex],
			Timestamp:  time.Now().Unix(),
			ServerInfo: "Chat Bot v1.0",
		}

		if err := stream.Send(reply); err != nil {
			slog.Error("Failed to send chat response", "error", err)
			return status.Error(codes.Internal, "failed to send response")
		}

		slog.Info("Chat message processed", "name", req.Name)
	}
}