package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	pb "grpc-ai-chat/pkg/api/chat/v1"
)

const (
	serverAddress = "localhost:50051"
	timeout       = 5 * time.Second
)

// ClientApp merepresentasikan aplikasi client
type ClientApp struct {
	conn      *grpc.ClientConn
	userConn  pb.UserServiceClient
	roomConn  pb.RoomServiceClient
	chatConn  pb.ChatServiceClient
	token     string
	userID    string
	username  string
	roomID    string
	reader    *bufio.Reader
}

func main() {
	app := NewClientApp()
	defer app.Close()

	if err := app.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	fmt.Println("🎉 Welcome to gRPC AI Chat!")
	fmt.Println("📝 Commands:")
	fmt.Println("   register <username> <password> <fullname> - Register user")
	fmt.Println("   login <username> <password> - Login")
	fmt.Println("   create_room <room_name> [description] [enable_ai] - Create room")
	fmt.Println("   list_rooms - List available rooms")
	fmt.Println("   join_room <room_id> - Join a room")
	fmt.Println("   chat - Start chatting (type 'exit' to leave chat)")
	fmt.Println("   leave_room - Leave current room")
	fmt.Println("   logout - Logout")
	fmt.Println("   exit - Exit application")
	fmt.Println("")

	app.InteractiveLoop()
}

// NewClientApp membuat instance ClientApp baru
func NewClientApp() *ClientApp {
	return &ClientApp{
		reader: bufio.NewReader(os.Stdin),
	}
}

// Connect menghubungkan ke gRPC server
func (app *ClientApp) Connect() error {
	conn, err := grpc.Dial(
		serverAddress,
		grpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	app.conn = conn
	app.userConn = pb.NewUserServiceClient(conn)
	app.roomConn = pb.NewRoomServiceClient(conn)
	app.chatConn = pb.NewChatServiceClient(conn)

	return nil
}

// Close menutup koneksi
func (app *ClientApp) Close() {
	if app.conn != nil {
		app.conn.Close()
	}
}

// InteractiveLoop menjalankan interactive command loop
func (app *ClientApp) InteractiveLoop() {
	for {
		// Tampilkan prompt
		prompt := "chat"
		if app.username != "" {
			prompt = fmt.Sprintf("%s@%s", app.username, prompt)
		}
		if app.roomID != "" {
			prompt = fmt.Sprintf("%s(room)", prompt)
		}
		fmt.Printf("%s> ", prompt)

		// Baca input
		input, err := app.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("❌ Error reading input: %v\n", err)
			continue
		}

		// Parse command
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := parts[0]

		app.HandleCommand(command, parts[1:])
	}
}

// HandleCommand menangani command dari user
func (app *ClientApp) HandleCommand(command string, args []string) {
	switch command {
	case "register":
		if len(args) < 3 {
			fmt.Println("❌ Usage: register <username> <password> <fullname>")
			return
		}
		app.Register(args[0], args[1], strings.Join(args[2:], " "))

	case "login":
		if len(args) < 2 {
			fmt.Println("❌ Usage: login <username> <password>")
			return
		}
		app.Login(args[0], args[1])

	case "create_room":
		if len(args) < 1 {
			fmt.Println("❌ Usage: create_room <room_name> [description] [enable_ai]")
			return
		}
		description := ""
		enableAI := false
		if len(args) > 1 {
			description = args[1]
		}
		if len(args) > 2 && (args[2] == "true" || args[2] == "1") {
			enableAI = true
		}
		app.CreateRoom(args[0], description, enableAI)

	case "list_rooms":
		app.ListRooms()

	case "join_room":
		if len(args) < 1 {
			fmt.Println("❌ Usage: join_room <room_id>")
			return
		}
		app.JoinRoom(args[0])

	case "chat":
		if app.roomID == "" {
			fmt.Println("❌ Please join a room first")
			return
		}
		app.StartChat()

	case "leave_room":
		app.LeaveRoom()

	case "logout":
		app.Logout()

	case "exit":
		fmt.Println("👋 Goodbye!")
		os.Exit(0)

	default:
		fmt.Println("❌ Unknown command. Type 'help' for commands.")
	}
}

// Register melakukan registrasi user
func (app *ClientApp) Register(username, password, fullName string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := app.userConn.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
		FullName: fullName,
	})
	if err != nil {
		fmt.Printf("❌ Registration failed: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("✅ Registration successful! User ID: %s\n", resp.UserId)
	} else {
		fmt.Printf("❌ %s\n", resp.Message)
	}
}

// Login melakukan login user
func (app *ClientApp) Login(username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := app.userConn.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		return
	}

	app.token = resp.Token
	app.userID = resp.UserId
	app.username = resp.Username

	fmt.Printf("✅ Login successful! Welcome %s\n", resp.Username)
	fmt.Printf("   Token expires in %d seconds\n", resp.ExpiresIn)
}

// CreateRoom membuat room baru
func (app *ClientApp) CreateRoom(name, description string, enableAI bool) {
	if app.token == "" {
		fmt.Println("❌ Please login first")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := app.roomConn.CreateRoom(ctx, &pb.CreateRoomRequest{
		Token:       app.token,
		RoomName:    name,
		Description: description,
		IsAiEnabled: enableAI,
	})
	if err != nil {
		fmt.Printf("❌ Create room failed: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("✅ Room created! ID: %s\n", resp.RoomId)
		if enableAI {
			fmt.Println("   🤖 AI is enabled in this room")
		}
	} else {
		fmt.Printf("❌ %s\n", resp.Message)
	}
}

// ListRooms menampilkan daftar rooms
func (app *ClientApp) ListRooms() {
	if app.token == "" {
		fmt.Println("❌ Please login first")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := app.roomConn.ListRooms(ctx, &pb.ListRoomsRequest{
		Token: app.token,
	})
	if err != nil {
		fmt.Printf("❌ List rooms failed: %v\n", err)
		return
	}

	if len(resp.Rooms) == 0 {
		fmt.Println("ℹ️  No rooms available")
		return
	}

	fmt.Printf("📋 Available Rooms (%d):\n", resp.Total)
	for _, room := range resp.Rooms {
		aiTag := ""
		if room.IsAiEnabled {
			aiTag = " 🤖"
		}
		fmt.Printf("   [%s] %s (%d members)%s\n", room.RoomId[:8], room.RoomName, room.MemberCount, aiTag)
		if room.Description != "" {
			fmt.Printf("       Description: %s\n", room.Description)
		}
	}
}

// JoinRoom bergabung ke room
func (app *ClientApp) JoinRoom(roomID string) {
	if app.token == "" {
		fmt.Println("❌ Please login first")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := app.roomConn.JoinRoom(ctx, &pb.JoinRoomRequest{
		Token:  app.token,
		RoomId: roomID,
	})
	if err != nil {
		fmt.Printf("❌ Join room failed: %v\n", err)
		return
	}

	if resp.Success {
		app.roomID = roomID
		fmt.Printf("✅ %s\n", resp.Message)
		fmt.Printf("   Room: %s (%d members)\n", resp.RoomInfo.RoomName, resp.RoomInfo.MemberCount)
		if resp.RoomInfo.IsAiEnabled {
			fmt.Println("   🤖 This room has AI enabled!")
		}
	} else {
		fmt.Printf("❌ %s\n", resp.Message)
	}
}

// StartChat memulai chat interaktif
func (app *ClientApp) StartChat() {
	fmt.Println("📢 Starting chat... (type 'exit' to leave)")

	// Create bidirectional stream
	stream, err := app.chatConn.Chat(context.Background())
	if err != nil {
		fmt.Printf("❌ Failed to start chat: %v\n", err)
		return
	}
	defer stream.CloseSend()

	// Send token untuk authentication
	if err := stream.Send(&pb.ChatMessage{
		Content: app.token,
	}); err != nil {
		fmt.Printf("❌ Failed to authenticate: %v\n", err)
		return
	}

	// Send room join message
	if err := stream.Send(&pb.ChatMessage{
		RoomId: app.roomID,
	}); err != nil {
		fmt.Printf("❌ Failed to join chat: %v\n", err)
		return
	}

	// Goroutine untuk menerima messages
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("📴 Server closed connection")
				return
			}
			if err != nil {
				fmt.Printf("❌ Error receiving: %v\n", err)
				return
			}

			// Display message
			switch msg.MessageType {
			case "system":
				fmt.Printf("ℹ️  [SYSTEM] %s\n", msg.Content)
			case "ai":
				fmt.Printf("🤖 [AI] %s\n", msg.Content)
			default:
				fmt.Printf("💬 [%s] %s\n", msg.Username, msg.Content)
			}
		}
	}()

	// Read dan send messages
	for {
		fmt.Print("📝 You> ")
		input, _ := app.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("👋 Left chat")
			return
		}

		if input == "" {
			continue
		}

		if err := stream.Send(&pb.ChatMessage{
			UserId:  app.userID,
			RoomId:  app.roomID,
			Content: input,
		}); err != nil {
			fmt.Printf("❌ Failed to send message: %v\n", err)
			return
		}
	}
}

// LeaveRoom keluar dari room
func (app *ClientApp) LeaveRoom() {
	if app.token == "" || app.roomID == "" {
		fmt.Println("❌ Please join a room first")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := app.roomConn.LeaveRoom(ctx, &pb.LeaveRoomRequest{
		Token:  app.token,
		RoomId: app.roomID,
	})
	if err != nil {
		fmt.Printf("❌ Leave room failed: %v\n", err)
		return
	}

	app.roomID = ""
	fmt.Printf("✅ %s\n", resp.Message)
}

// Logout melakukan logout
func (app *ClientApp) Logout() {
	app.token = ""
	app.userID = ""
	app.username = ""
	app.roomID = ""
	fmt.Println("✅ Logged out successfully")
}
