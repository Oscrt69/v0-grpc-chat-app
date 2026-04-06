package services

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "grpc-ai-chat/pkg/api/chat/v1"
	"grpc-ai-chat/internal/auth"
	"grpc-ai-chat/internal/models"
)

// RoomService mengimplementasikan RoomService gRPC
type RoomService struct {
	pb.UnimplementedRoomServiceServer
	stateManager *models.StateManager
}

// NewRoomService membuat instance RoomService baru
func NewRoomService(stateManager *models.StateManager) *RoomService {
	return &RoomService{
		stateManager: stateManager,
	}
}

// CreateRoom menghandle create room request (Unary RPC)
func (rs *RoomService) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	// Validasi token
	userID, err := auth.GetUserIDFromToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "token tidak valid")
	}

	// Cek apakah session masih valid
	if !rs.stateManager.IsSessionValid(req.Token) {
		return nil, status.Error(codes.Unauthenticated, "session telah kadaluarsa")
	}

	// Validasi input
	if req.RoomName == "" {
		return nil, status.Error(codes.InvalidArgument, "nama room harus diisi")
	}

	// Create room baru
	roomID := auth.GenerateID()
	room := &models.Room{
		ID:          roomID,
		Name:        req.RoomName,
		Description: req.Description,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		Members:     make(map[string]*models.RoomMember),
		IsAIEnabled: req.IsAiEnabled,
	}

	// Add creator sebagai member
	user := rs.stateManager.GetUser(userID)
	if user == nil {
		return nil, status.Error(codes.Internal, "user tidak ditemukan")
	}

	room.AddMember(userID, user.Username)
	rs.stateManager.AddRoom(room)

	return &pb.CreateRoomResponse{
		Success:  true,
		RoomId:   roomID,
		RoomName: req.RoomName,
		Message:  "room berhasil dibuat",
	}, nil
}

// ListRooms menghandle list rooms request (Unary RPC)
func (rs *RoomService) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	// Validasi token
	_, err := auth.GetUserIDFromToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "token tidak valid")
	}

	// Cek apakah session masih valid
	if !rs.stateManager.IsSessionValid(req.Token) {
		return nil, status.Error(codes.Unauthenticated, "session telah kadaluarsa")
	}

	// Get semua rooms
	rooms := rs.stateManager.ListRooms()

	// Convert to proto message
	protoRooms := make([]*pb.RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		protoRooms = append(protoRooms, &pb.RoomInfo{
			RoomId:      room.ID,
			RoomName:    room.Name,
			Description: room.Description,
			MemberCount: int32(room.GetMemberCount()),
			IsAiEnabled: room.IsAIEnabled,
			CreatedAt:   room.CreatedAt.Unix(),
		})
	}

	return &pb.ListRoomsResponse{
		Rooms: protoRooms,
		Total: int32(len(protoRooms)),
	}, nil
}

// JoinRoom menghandle join room request (Unary RPC)
func (rs *RoomService) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.JoinRoomResponse, error) {
	// Validasi token
	userID, err := auth.GetUserIDFromToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "token tidak valid")
	}

	// Cek apakah session masih valid
	if !rs.stateManager.IsSessionValid(req.Token) {
		return nil, status.Error(codes.Unauthenticated, "session telah kadaluarsa")
	}

	// Cek apakah room exists
	room := rs.stateManager.GetRoom(req.RoomId)
	if room == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("room %s tidak ditemukan", req.RoomId))
	}

	// Get user info
	user := rs.stateManager.GetUser(userID)
	if user == nil {
		return nil, status.Error(codes.Internal, "user tidak ditemukan")
	}

	// Add member ke room
	room.AddMember(userID, user.Username)

	return &pb.JoinRoomResponse{
		Success: true,
		Message: fmt.Sprintf("berhasil bergabung dengan room %s", room.Name),
		RoomInfo: &pb.RoomInfo{
			RoomId:      room.ID,
			RoomName:    room.Name,
			Description: room.Description,
			MemberCount: int32(room.GetMemberCount()),
			IsAiEnabled: room.IsAIEnabled,
			CreatedAt:   room.CreatedAt.Unix(),
		},
	}, nil
}

// LeaveRoom menghandle leave room request (Unary RPC)
func (rs *RoomService) LeaveRoom(ctx context.Context, req *pb.LeaveRoomRequest) (*pb.LeaveRoomResponse, error) {
	// Validasi token
	userID, err := auth.GetUserIDFromToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "token tidak valid")
	}

	// Cek apakah session masih valid
	if !rs.stateManager.IsSessionValid(req.Token) {
		return nil, status.Error(codes.Unauthenticated, "session telah kadaluarsa")
	}

	// Cek apakah room exists
	room := rs.stateManager.GetRoom(req.RoomId)
	if room == nil {
		return nil, status.Error(codes.NotFound, "room tidak ditemukan")
	}

	// Remove member dari room
	room.RemoveMember(userID)

	// Hapus room jika tidak ada members
	if room.GetMemberCount() == 0 {
		rs.stateManager.DeleteRoom(req.RoomId)
	}

	return &pb.LeaveRoomResponse{
		Success: true,
		Message: "berhasil keluar dari room",
	}, nil
}
