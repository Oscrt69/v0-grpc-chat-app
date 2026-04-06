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

// UserService mengimplementasikan UserService gRPC
type UserService struct {
	pb.UnimplementedUserServiceServer
	stateManager *models.StateManager
}

// NewUserService membuat instance UserService baru
func NewUserService(stateManager *models.StateManager) *UserService {
	return &UserService{
		stateManager: stateManager,
	}
}

// Login menghandle login request (Unary RPC)
func (us *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validasi input
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username dan password harus diisi")
	}

	// Cek apakah user exists
	user := us.stateManager.GetUserByUsername(req.Username)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "username atau password salah")
	}

	// Verify password
	if !auth.VerifyPassword(user.Password, req.Password) {
		return nil, status.Error(codes.Unauthenticated, "username atau password salah")
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "gagal membuat token")
	}

	// Create session
	session := &models.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(auth.TokenDuration),
		Created:   time.Now(),
	}
	us.stateManager.AddSession(session)

	return &pb.LoginResponse{
		Token:     token,
		UserId:    user.ID,
		Username:  user.Username,
		ExpiresIn: int32(auth.TokenDuration.Seconds()),
	}, nil
}

// Register menghandle register request (Unary RPC)
func (us *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Validasi input
	if req.Username == "" || req.Password == "" || req.FullName == "" {
		return nil, status.Error(codes.InvalidArgument, "semua field harus diisi")
	}

	// Cek apakah username sudah ada
	existingUser := us.stateManager.GetUserByUsername(req.Username)
	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "username sudah terdaftar")
	}

	// Validasi password strength (minimal 6 karakter)
	if len(req.Password) < 6 {
		return nil, status.Error(codes.InvalidArgument, "password minimal 6 karakter")
	}

	// Create user baru
	userID := auth.GenerateID()
	user := &models.User{
		ID:       userID,
		Username: req.Username,
		Password: auth.HashPassword(req.Password),
		FullName: req.FullName,
		Created:  time.Now(),
	}

	us.stateManager.AddUser(user)

	return &pb.RegisterResponse{
		Success: true,
		Message: "registrasi berhasil",
		UserId:  userID,
	}, nil
}
