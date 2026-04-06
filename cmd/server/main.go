package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "grpc-ai-chat/pkg/api/chat/v1"
	"grpc-ai-chat/internal/models"
	"grpc-ai-chat/internal/services"
	"grpc-ai-chat/internal/auth"
)

const (
	port = ":50051"
	host = "0.0.0.0"
)

func main() {
	fmt.Println("🚀 Starting gRPC AI Chat Server...")

	// Initialize state manager
	stateManager := models.NewStateManager()

	// Initialize default users untuk testing
	initializeDefaultUsers(stateManager)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	userService := services.NewUserService(stateManager)
	roomService := services.NewRoomService(stateManager)
	chatService := services.NewChatService(stateManager)

	pb.RegisterUserServiceServer(grpcServer, userService)
	pb.RegisterRoomServiceServer(grpcServer, roomService)
	pb.RegisterChatServiceServer(grpcServer, chatService)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf("%s%s", host, port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Printf("✅ Server listening at %s\n", listener.Addr())
	fmt.Println("📝 Default users:")
	fmt.Println("   - Username: alice, Password: password123")
	fmt.Println("   - Username: bob, Password: password123")
	fmt.Println("   - Username: charlie, Password: password123")
	fmt.Println("\n🔌 gRPC Services:")
	fmt.Println("   - UserService (Login, Register)")
	fmt.Println("   - RoomService (Create, List, Join, Leave)")
	fmt.Println("   - ChatService (Bidirectional Streaming Chat)")
	fmt.Println("\n⚙️  AI Integration: Grok API")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// initializeDefaultUsers membuat default users untuk testing
func initializeDefaultUsers(sm *models.StateManager) {
	users := []struct {
		username string
		password string
		fullName string
	}{
		{"alice", "password123", "Alice Johnson"},
		{"bob", "password123", "Bob Smith"},
		{"charlie", "password123", "Charlie Brown"},
	}

	for _, u := range users {
		user := &models.User{
			ID:       auth.GenerateID(),
			Username: u.username,
			Password: auth.HashPassword(u.password),
			FullName: u.fullName,
		}
		sm.AddUser(user)
		fmt.Printf("👤 Created user: %s\n", u.username)
	}
}
