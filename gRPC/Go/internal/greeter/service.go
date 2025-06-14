package service

import (
	"context"
	"fmt"
	"log/slog"

	v1 "github.com/Krushnal121/API-Hub/gRPC/Go/api/gen/v1"
)

type GreeterService struct {
	v1.UnimplementedGreeterServer
}

func (s *GreeterService) SayHello(ctx context.Context, req *v1.HelloRequest) (*v1.HelloReply, error) {
	message := fmt.Sprintf("Hello, %s!", req.Name)
	slog.Info("Received SayHello request", "name", req.Name)
	return &v1.HelloReply{Message: message}, nil
}