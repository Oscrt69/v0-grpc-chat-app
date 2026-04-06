package models

import (
	"sync"
	"time"
)

// User merepresentasikan user dalam sistem
type User struct {
	ID       string
	Username string
	Password string
	FullName string
	Created  time.Time
}

// Session merepresentasikan session aktif
type Session struct {
	UserID    string
	Token     string
	ExpiresAt time.Time
	Created   time.Time
}

// Room merepresentasikan chat room
type Room struct {
	ID          string
	Name        string
	Description string
	CreatedBy   string
	CreatedAt   time.Time
	Members     map[string]*RoomMember
	IsAIEnabled bool
	mu          sync.RWMutex
}

// RoomMember merepresentasikan member dalam room
type RoomMember struct {
	UserID    string
	Username  string
	JoinedAt  time.Time
	Stream    interface{} // Chat stream
}

// StateManager mengelola seluruh state in-memory
type StateManager struct {
	Users    map[string]*User
	Sessions map[string]*Session
	Rooms    map[string]*Room
	mu       sync.RWMutex
}

// NewStateManager membuat instance StateManager baru
func NewStateManager() *StateManager {
	return &StateManager{
		Users:    make(map[string]*User),
		Sessions: make(map[string]*Session),
		Rooms:    make(map[string]*Room),
	}
}

// AddUser menambah user baru
func (sm *StateManager) AddUser(user *User) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.Users[user.ID] = user
}

// GetUser mendapatkan user by ID
func (sm *StateManager) GetUser(userID string) *User {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.Users[userID]
}

// GetUserByUsername mendapatkan user by username
func (sm *StateManager) GetUserByUsername(username string) *User {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, user := range sm.Users {
		if user.Username == username {
			return user
		}
	}
	return nil
}

// AddSession menambah session baru
func (sm *StateManager) AddSession(session *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.Sessions[session.Token] = session
}

// GetSession mendapatkan session by token
func (sm *StateManager) GetSession(token string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.Sessions[token]
}

// DeleteSession menghapus session
func (sm *StateManager) DeleteSession(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.Sessions, token)
}

// IsSessionValid mengecek apakah session masih valid
func (sm *StateManager) IsSessionValid(token string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.Sessions[token]
	if !exists {
		return false
	}
	return time.Now().Before(session.ExpiresAt)
}

// AddRoom menambah room baru
func (sm *StateManager) AddRoom(room *Room) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.Rooms[room.ID] = room
}

// GetRoom mendapatkan room by ID
func (sm *StateManager) GetRoom(roomID string) *Room {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.Rooms[roomID]
}

// ListRooms mendapatkan semua rooms
func (sm *StateManager) ListRooms() []*Room {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	rooms := make([]*Room, 0, len(sm.Rooms))
	for _, room := range sm.Rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// DeleteRoom menghapus room
func (sm *StateManager) DeleteRoom(roomID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.Rooms, roomID)
}

// Room Methods

// AddMember menambah member ke room
func (r *Room) AddMember(userID, username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Members[userID] = &RoomMember{
		UserID:   userID,
		Username: username,
		JoinedAt: time.Now(),
	}
}

// RemoveMember menghapus member dari room
func (r *Room) RemoveMember(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Members, userID)
}

// GetMembers mendapatkan semua members
func (r *Room) GetMembers() map[string]*RoomMember {
	r.mu.RLock()
	defer r.mu.RUnlock()
	members := make(map[string]*RoomMember)
	for k, v := range r.Members {
		members[k] = v
	}
	return members
}

// GetMemberCount mendapatkan jumlah members
func (r *Room) GetMemberCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Members)
}
