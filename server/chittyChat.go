package main

import (
	"context"
	"fmt"
	"log"
	"net"

	chatpackage "github.com/Troelshjarne/mini_project_2/chat"

	"google.golang.org/grpc"
)

type Server struct {
	chatpackage.UnimplementedPublishServer
}

func (s *Server) GetMessage(ctx context.Context, in *chatpackage.ChatMessageRequest) (*chatpackage.ChatMessage, error) {
	fmt.Printf("GetMessage test \n")
	return &chatpackage.ChatMessage{Message: "Test reply"}, nil
}

func main() {

	list, err := net.Listen("tcp", ":9080")
	if err == nil {
		fmt.Print("succes")
	}

	if err != nil {
		log.Fatalf("Failed to listen on port 9080: %v", err)
	}

	grpcServer := grpc.NewServer()

	chatpackage.RegisterPublishServer(grpcServer, &Server{})

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}
