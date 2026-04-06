package services

import (
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "grpc-ai-chat/pkg/api/chat/v1"
	"grpc-ai-chat/internal/ai"
	"grpc-ai-chat/internal/auth"
	"grpc-ai-chat/internal/models"
)

// ChatService mengimplementasikan ChatService gRPC
type ChatService struct {
	pb.UnimplementedChatServiceServer
	stateManager *models.StateManager
	grokClient   *ai.GrokClient
	// Map untuk menyimpan active streams per room
	roomStreams map[string][]pb.ChatService_ChatServer
	streamsMu   sync.RWMutex
}

// NewChatService membuat instance ChatService baru
func NewChatService(stateManager *models.StateManager) *ChatService {
	return &ChatService{
		stateManager: stateManager,
		grokClient:   ai.NewGrokClient(),
		roomStreams:  make(map[string][]pb.ChatService_ChatServer),
	}
}

// Chat menghandle bidirectional streaming RPC untuk chat realtime
func (cs *ChatService) Chat(stream pb.ChatService_ChatServer) error {
	// Channel untuk menerima messages
	msgChan := make(chan *pb.ChatMessage, 100)
	errChan := make(chan error, 1)

	// Goroutine untuk menerima messages dari client
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				close(msgChan)
				return
			}
			if err != nil {
				errChan <- status.Error(codes.Internal, "gagal menerima message")
				return
			}
			msgChan <- msg
		}
	}()

	// Variables untuk tracking session
	var userID, username, roomID string
	var isAuthenticated bool = false
	var roomMembers = make(map[string]bool)

	// Main message processing loop
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				// Client disconnect
				if isAuthenticated && roomID != "" {
					cs.broadcastSystemMessage(roomID, fmt.Sprintf("%s telah keluar dari chat", username))
					room := cs.stateManager.GetRoom(roomID)
					if room != nil {
						room.RemoveMember(userID)
					}
				}
				return nil
			}

			// Proses message
			err := cs.processMessage(msg, stream, &userID, &username, &roomID, &isAuthenticated, roomMembers)
			if err != nil {
				return err
			}

		case <-stream.Context().Done():
			return stream.Context().Err()

		case err := <-errChan:
			return err
		}
	}
}

// processMessage memproses incoming message
func (cs *ChatService) processMessage(
	msg *pb.ChatMessage,
	stream pb.ChatService_ChatServer,
	userID, username, roomID *string,
	isAuthenticated *bool,
	roomMembers map[string]bool,
) error {
	// Jika belum authenticated, validate token
	if !*isAuthenticated {
		var err error
		*userID, err = auth.GetUserIDFromToken(msg.Content)
		if err != nil {
			return status.Error(codes.Unauthenticated, "token tidak valid")
		}

		// Validate session
		if !cs.stateManager.IsSessionValid(msg.Content) {
			return status.Error(codes.Unauthenticated, "session telah kadaluarsa")
		}

		// Get username
		*username, err = auth.GetUsernameFromToken(msg.Content)
		if err != nil {
			return status.Error(codes.Internal, "gagal mendapat username")
		}

		*isAuthenticated = true

		// Send auth success message
		return stream.Send(&pb.ChatMessage{
			MessageId:   auth.GenerateID(),
			MessageType: "system",
			Content:     "authenticated",
			Timestamp:   time.Now().Unix(),
		})
	}

	// Jika belum join room, proses join
	if *roomID == "" {
		room := cs.stateManager.GetRoom(msg.RoomId)
		if room == nil {
			return status.Error(codes.NotFound, "room tidak ditemukan")
		}

		*roomID = msg.RoomId
		room.AddMember(*userID, *username)

		// Broadcast join message
		cs.broadcastSystemMessage(*roomID, fmt.Sprintf("%s telah bergabung dengan chat", *username))

		// Send room members info
		members := room.GetMembers()
		for memberID := range members {
			roomMembers[memberID] = true
		}

		return nil
	}

	// Process regular chat message
	if msg.Content == "" {
		return status.Error(codes.InvalidArgument, "content tidak boleh kosong")
	}

	// Create message response
	responseMsg := &pb.ChatMessage{
		MessageId:   auth.GenerateID(),
		UserId:      *userID,
		Username:    *username,
		RoomId:      *roomID,
		Content:     msg.Content,
		Timestamp:   time.Now().Unix(),
		MessageType: "user",
	}

	// Broadcast user message ke semua clients di room
	cs.broadcastToRoom(*roomID, responseMsg)

	// Jika room AI-enabled, generate AI response
	room := cs.stateManager.GetRoom(*roomID)
	if room != nil && room.IsAIEnabled {
		go cs.generateAndBroadcastAIResponse(*roomID, msg.Content, *username)
	}

	return nil
}

// broadcastToRoom mengirim message ke semua clients di room
func (cs *ChatService) broadcastToRoom(roomID string, msg *pb.ChatMessage) {
	cs.streamsMu.RLock()
	streams, exists := cs.roomStreams[roomID]
	cs.streamsMu.RUnlock()

	if !exists || len(streams) == 0 {
		return
	}

	// Broadcast ke semua connected clients
	for _, s := range streams {
		s.Send(msg)
	}
}

// broadcastSystemMessage mengirim system message ke semua clients di room
func (cs *ChatService) broadcastSystemMessage(roomID, content string) {
	msg := &pb.ChatMessage{
		MessageId:   auth.GenerateID(),
		RoomId:      roomID,
		Content:     content,
		Timestamp:   time.Now().Unix(),
		MessageType: "system",
	}
	cs.broadcastToRoom(roomID, msg)
}

// generateAndBroadcastAIResponse menghasilkan dan broadcast AI response
func (cs *ChatService) generateAndBroadcastAIResponse(roomID, userMessage, username string) {
	// Generate response dari Grok
	aiResponse, err := cs.grokClient.GenerateResponse(userMessage, "")
	if err != nil {
		// Broadcast error message
		errorMsg := &pb.ChatMessage{
			MessageId:    auth.GenerateID(),
			RoomId:       roomID,
			Content:      fmt.Sprintf("AI Error: %v", err),
			Timestamp:    time.Now().Unix(),
			MessageType:  "ai",
			IsAiResponse: true,
			ErrorMessage: err.Error(),
		}
		cs.broadcastToRoom(roomID, errorMsg)
		return
	}

	// Broadcast AI response
	aiMsg := &pb.ChatMessage{
		MessageId:    auth.GenerateID(),
		Username:     "AI Assistant",
		RoomId:       roomID,
		Content:      aiResponse,
		Timestamp:    time.Now().Unix(),
		MessageType:  "ai",
		IsAiResponse: true,
	}
	cs.broadcastToRoom(roomID, aiMsg)
}

// AddStream menambah stream untuk room
func (cs *ChatService) AddStream(roomID string, stream pb.ChatService_ChatServer) {
	cs.streamsMu.Lock()
	defer cs.streamsMu.Unlock()
	cs.roomStreams[roomID] = append(cs.roomStreams[roomID], stream)
}

// RemoveStream menghapus stream dari room
func (cs *ChatService) RemoveStream(roomID string, stream pb.ChatService_ChatServer) {
	cs.streamsMu.Lock()
	defer cs.streamsMu.Unlock()

	streams := cs.roomStreams[roomID]
	for i, s := range streams {
		if s == stream {
			cs.roomStreams[roomID] = append(streams[:i], streams[i+1:]...)
			break
		}
	}
}
