# Implementation Details

Dokumentasi teknis mendalam tentang implementasi setiap komponen.

## 1. State Management Implementation

### In-Memory Storage Architecture

```go
type StateManager struct {
    Users    map[string]*User        // userID -> User
    Sessions map[string]*Session     // token -> Session
    Rooms    map[string]*Room        // roomID -> Room
    mu       sync.RWMutex
}
```

### Thread Safety Pattern

**Pattern: Reader-Writer Mutex**

```go
// Multiple readers dapat access concurrently
sm.mu.RLock()
defer sm.mu.RUnlock()
user := sm.Users[userID]

// Single writer, exclusive access
sm.mu.Lock()
defer sm.mu.Unlock()
sm.Users[userID] = newUser
```

**Why RWMutex?**
- 90% operations adalah read (lookups)
- Multiple goroutines dapat read simultaneously
- Writes block everything (rare operation)

### Garbage Collection

**Auto-cleanup mechanisms:**

1. **Empty Rooms**
```go
func (rs *RoomService) LeaveRoom(...) {
    room.RemoveMember(userID)
    if room.GetMemberCount() == 0 {
        rs.stateManager.DeleteRoom(req.RoomId)
    }
}
```

2. **Expired Sessions** (Manual cleanup needed for production)
```go
// Should be scheduled periodically
func (sm *StateManager) CleanupExpiredSessions() {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    now := time.Now()
    for token, session := range sm.Sessions {
        if now.After(session.ExpiresAt) {
            delete(sm.Sessions, token)
        }
    }
}
```

### Memory Usage Calculation

```
Per User:
  - ID (UUID): 36 bytes
  - Username: ~20 bytes average
  - Password (hashed): ~60 bytes
  - FullName: ~30 bytes
  - Created: 8 bytes
  - Total: ~154 bytes per user
  
Per Session:
  - UserID: 36 bytes
  - Token: ~200 bytes
  - ExpiresAt: 8 bytes
  - Created: 8 bytes
  - Total: ~252 bytes per session

Per Room:
  - ID: 36 bytes
  - Name: ~30 bytes
  - Description: ~100 bytes
  - CreatedBy: 36 bytes
  - CreatedAt: 8 bytes
  - Members map: ~100 bytes per member
  - Total: ~210 bytes + (100 * memberCount)

Example for 100 users, 10 rooms with 50 members each:
  - Users: 100 * 154 = 15.4 KB
  - Sessions: 100 * 252 = 25.2 KB
  - Rooms: 10 * (210 + 100*50) = 52.1 KB
  - Total: ~92.7 KB
```

---

## 2. Authentication & JWT Implementation

### Token Generation Flow

```go
func GenerateToken(userID, username string) (string, error) {
    claims := jwt.MapClaims{
        "user_id":  userID,
        "username": username,
        "exp":      time.Now().Add(TokenDuration).Unix(),  // 24 hours
        "iat":      time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(JWTSecret))
    
    return tokenString, nil
}
```

**Token Structure Example:**

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.
eyJ1c2VyX2lkIjoiMTIzNDU2Nzg5MCIsInVzZXJuYW1lIjoiYWxpY2UiLCJleHAiOjE2OTMyNDI0NjcsImlhdCI6MTY5MzE1NjA2N30.
X2WDC3Ty4EjxoKiDRm4g8-QT5yQMJLYdJvYKKxKrXL0

Header:  {"alg":"HS256","typ":"JWT"}
Payload: {"user_id":"alice-123","username":"alice","exp":1693242467,"iat":1693156067}
Signature: (HMAC-SHA256)
```

### Token Validation Flow

```go
func ValidateToken(tokenString string) (map[string]interface{}, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        jwt.MapClaims{},
        func(token *jwt.Token) (interface{}, error) {
            return []byte(JWTSecret), nil
        },
    )
    
    if err != nil {
        return nil, err  // Malformed token
    }
    
    if !token.Valid {
        return nil, fmt.Errorf("invalid token")  // Bad signature
    }
    
    claims := token.Claims.(jwt.MapClaims)
    exp := claims["exp"].(float64)
    
    if int64(exp) < time.Now().Unix() {
        return nil, fmt.Errorf("token expired")  // Expired
    }
    
    return claims, nil
}
```

### Session Management

```go
// Create session on login
session := &models.Session{
    UserID:    user.ID,
    Token:     token,
    ExpiresAt: time.Now().Add(24 * time.Hour),
    Created:   time.Now(),
}
sm.AddSession(session)

// Validate session on service calls
func (sm *StateManager) IsSessionValid(token string) bool {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    session, exists := sm.Sessions[token]
    if !exists {
        return false  // Session not found
    }
    
    return time.Now().Before(session.ExpiresAt)  // Not expired
}
```

**Security Notes:**
- Tokens valid 24 jam
- Token stored di client saja
- Server validate setiap request
- Session bisa invalidated kapan saja
- (Production: Add token refresh mechanism)

---

## 3. Bidirectional Streaming Implementation

### Stream Lifecycle

```
1. CONNECT
   ├─ Client establishes stream
   ├─ Server creates goroutine untuk stream
   └─ Keep connection open

2. SEND (Client → Server)
   ├─ Client sends ChatMessage
   ├─ Server Recv() blocks waiting
   ├─ Once received, process message
   └─ Continue listening

3. BROADCAST (Server → All Clients)
   ├─ Server iterates all room streams
   ├─ Send message to each stream
   ├─ Use buffered channel for non-blocking
   └─ Continue processing

4. CLOSE (Client disconnect)
   ├─ Client closes stream
   ├─ Server Recv() returns io.EOF
   ├─ Cleanup room membership
   └─ Broadcast leave message
```

### Message Processing Loop

```go
func (cs *ChatService) Chat(stream pb.ChatService_ChatServer) error {
    msgChan := make(chan *pb.ChatMessage, 100)  // Buffered
    errChan := make(chan error, 1)
    
    // Goroutine 1: Receive messages
    go func() {
        for {
            msg, err := stream.Recv()
            if err == io.EOF {
                close(msgChan)
                return
            }
            msgChan <- msg
        }
    }()
    
    // Goroutine 2: Main loop
    for {
        select {
        case msg, ok := <-msgChan:
            if !ok {
                return nil  // Client disconnect
            }
            cs.processMessage(msg, stream, ...)
            
        case <-stream.Context().Done():
            return stream.Context().Err()  // Context cancelled
            
        case err := <-errChan:
            return err
        }
    }
}
```

### Broadcasting Pattern

```go
// Store streams per room
cs.roomStreams map[string][]pb.ChatService_ChatServer

// Broadcast message
func (cs *ChatService) broadcastToRoom(roomID string, msg *pb.ChatMessage) {
    cs.streamsMu.RLock()
    streams := cs.roomStreams[roomID]
    cs.streamsMu.RUnlock()
    
    for _, s := range streams {
        // Non-blocking send (use goroutine if needed)
        s.Send(msg)
    }
}
```

### Non-Blocking AI Response

```go
// Generate AI response in background
go func() {
    aiResponse, err := cs.grokClient.GenerateResponse(
        userMessage,
        "",
    )
    
    if err != nil {
        // Send error message
        cs.broadcastToRoom(roomID, &pb.ChatMessage{
            MessageType:  "ai",
            ErrorMessage: err.Error(),
        })
        return
    }
    
    // Send AI response
    cs.broadcastToRoom(roomID, &pb.ChatMessage{
        MessageType:  "ai",
        Content:      aiResponse,
    })
}()

// Don't wait - return immediately
```

---

## 4. gRPC Service Implementation Patterns

### Service Registration

```go
// In main.go
pb.RegisterUserServiceServer(grpcServer, userService)
pb.RegisterRoomServiceServer(grpcServer, roomService)
pb.RegisterChatServiceServer(grpcServer, chatService)

// Proto-generated code creates:
// - UserServiceServer interface with all methods
// - Service descriptor for routing
// - Default unimplemented server
```

### Unary RPC Pattern (Login)

```go
func (us *UserService) Login(
    ctx context.Context,
    req *pb.LoginRequest,
) (*pb.LoginResponse, error) {
    // 1. Validate input
    if req.Username == "" {
        return nil, status.Error(codes.InvalidArgument, "...")
    }
    
    // 2. Business logic
    user := us.stateManager.GetUserByUsername(req.Username)
    if user == nil {
        return nil, status.Error(codes.Unauthenticated, "...")
    }
    
    // 3. Generate response
    token, _ := auth.GenerateToken(user.ID, user.Username)
    
    // 4. Update state
    us.stateManager.AddSession(&models.Session{...})
    
    // 5. Return response
    return &pb.LoginResponse{
        Token:     token,
        UserId:    user.ID,
        ExpiresIn: 86400,
    }, nil
}
```

### Streaming RPC Pattern (Chat)

```go
func (cs *ChatService) Chat(
    stream pb.ChatService_ChatServer,
) error {
    // 1. Setup
    msgChan := make(chan *pb.ChatMessage, 100)
    
    // 2. Background receiver
    go func() {
        for {
            msg, err := stream.Recv()
            if err == io.EOF {
                close(msgChan)
                return
            }
            msgChan <- msg
        }
    }()
    
    // 3. Message loop
    for {
        select {
        case msg := <-msgChan:
            // Process
            err := cs.processMessage(msg, stream)
            if err != nil {
                return err
            }
        case <-stream.Context().Done():
            return stream.Context().Err()
        }
    }
}
```

---

## 5. Error Handling Strategy

### Error Propagation

```
Client
    ↓ Request
Server Service Layer
    ├─ Input validation error
    │   └─ Return status.Error(codes.InvalidArgument, "msg")
    │
    ├─ Authentication error
    │   └─ Return status.Error(codes.Unauthenticated, "msg")
    │
    ├─ Resource not found
    │   └─ Return status.Error(codes.NotFound, "msg")
    │
    └─ Internal error
        └─ Return status.Error(codes.Internal, "msg")
    ↓ Response with gRPC Status
Client receives error with code & message
```

### Error Handling Best Practices

```go
// ✅ Good: Meaningful error message
if user == nil {
    return nil, status.Error(codes.NotFound, "user tidak ditemukan")
}

// ❌ Bad: Generic error message
if user == nil {
    return nil, status.Error(codes.NotFound, "error")
}

// ✅ Good: Log for debugging, return to client
log.Printf("Database error: %v", err)
return nil, status.Error(codes.Internal, "gagal mengakses database")

// ❌ Bad: Expose internal details to client
return nil, status.Error(codes.Internal, fmt.Sprintf(
    "Database connection failed: %v", err,
))

// ✅ Good: Check all preconditions
if token == "" {
    return nil, status.Error(codes.InvalidArgument, "token harus diisi")
}
if !sm.IsSessionValid(token) {
    return nil, status.Error(codes.Unauthenticated, "session expired")
}

// ❌ Bad: Only check one condition
if token == "" {
    return nil, status.Error(codes.InvalidArgument, "token harus diisi")
}
// Continue without checking validity...
```

---

## 6. AI Integration Implementation

### Grok API Call Flow

```go
func (gc *GrokClient) GenerateResponse(
    userMessage string,
    conversationContext string,
) (string, error) {
    // 1. Build request
    messages := []GrokMessage{
        {Role: "system", Content: "system prompt"},
        {Role: "user", Content: userMessage},
    }
    
    request := GrokRequest{
        Model:    "grok-2",
        Messages: messages,
        MaxToken: 256,
        Temp:     0.7,
    }
    
    // 2. Serialize to JSON
    jsonBody, _ := json.Marshal(request)
    
    // 3. Create HTTP request
    httpReq, _ := http.NewRequest(
        "POST",
        "https://api.x.ai/v1/chat/completions",
        bytes.NewBuffer(jsonBody),
    )
    
    // 4. Set headers
    httpReq.Header.Set("Authorization", "Bearer API_KEY")
    httpReq.Header.Set("Content-Type", "application/json")
    
    // 5. Execute
    resp, err := gc.client.Do(httpReq)
    
    // 6. Parse response
    var grokResp GrokResponse
    json.Unmarshal(body, &grokResp)
    
    // 7. Return content
    return grokResp.Choices[0].Message.Content, nil
}
```

### API Response Structure

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "grok-2",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "gRPC is a high-performance RPC framework..."
      }
    }
  ]
}
```

### Error Handling for External API

```go
// Status code check
if resp.StatusCode != http.StatusOK {
    return "", fmt.Errorf("Grok API error: %s (status: %d)",
        string(body), resp.StatusCode)
}

// Response parsing error
if len(grokResp.Choices) == 0 {
    return "", fmt.Errorf("no response choices from Grok")
}

// Network error
if err != nil {
    return "", fmt.Errorf("failed to call Grok API: %w", err)
}
```

---

## 7. Protocol Buffers Compilation

### Proto File Structure

```protobuf
syntax = "proto3";
package chat.v1;
option go_package = "grpc-ai-chat/pkg/api/chat/v1";

service UserService {
  rpc Login(LoginRequest) returns (LoginResponse) {}
}

message LoginRequest {
  string username = 1;
  string password = 2;
}

message LoginResponse {
  string token = 1;
  string user_id = 2;
}
```

### Compilation Command

```bash
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/chat.proto
```

### Generated Code

**Generated files:**
- `pkg/api/chat/v1/chat.pb.go` - Message definitions
- `pkg/api/chat/v1/chat_grpc.pb.go` - Service interfaces

**In chat.pb.go:**
```go
type LoginRequest struct {
    Username string
    Password string
}

func (x *LoginRequest) GetUsername() string
func (x *LoginRequest) ProtoReflect() protoreflect.Message
func (x *LoginRequest) Reset()
func (x *LoginRequest) String() string
func (x *LoginRequest) Validate() error
// ... and more marshal/unmarshal methods
```

**In chat_grpc.pb.go:**
```go
type UserServiceServer interface {
    Login(context.Context, *LoginRequest) (*LoginResponse, error)
    Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
}

type UserServiceClient interface {
    Login(ctx context.Context, in *LoginRequest, ...) (*LoginResponse, error)
    Register(ctx context.Context, in *RegisterRequest, ...) (*RegisterResponse, error)
}
```

---

## 8. Concurrency Patterns Used

### Pattern 1: Goroutine per Stream

```go
// Each Chat() call runs in separate goroutine
go cs.Chat(stream1)  // Alice
go cs.Chat(stream2)  // Bob
go cs.Chat(stream3)  // Charlie
// Each runs independently
```

### Pattern 2: Channel for Message Queue

```go
msgChan := make(chan *pb.ChatMessage, 100)

// Producer: Receiver goroutine
go func() {
    for {
        msg, _ := stream.Recv()
        msgChan <- msg
    }
}()

// Consumer: Main loop
for msg := range msgChan {
    process(msg)
}
```

### Pattern 3: RWMutex for Shared State

```go
// Multiple readers
go func() {
    sm.mu.RLock()
    user := sm.Users[userID]
    sm.mu.RUnlock()
}()

// Single writer
go func() {
    sm.mu.Lock()
    sm.Users[userID] = newUser
    sm.mu.Unlock()
}()
```

### Pattern 4: Non-blocking Operations

```go
// Generate AI response without blocking chat stream
go func() {
    response, _ := grokClient.GenerateResponse(msg)
    broadcastToRoom(roomID, response)
}()
// Return immediately
```

---

## 9. Performance Optimizations

### Current Optimizations

1. **Buffered Channels**
   - `msgChan := make(chan *pb.ChatMessage, 100)`
   - Prevent blocking on send

2. **RWMutex over Full Lock**
   - Multiple concurrent reads
   - Only exclusive lock on writes

3. **Stream Pooling** (Implicit)
   - Keep streams open
   - Avoid connection overhead
   - Reuse for multiple messages

4. **Async AI Processing**
   - Non-blocking goroutine
   - Don't wait for API response
   - Broadcast when ready

### Potential Optimizations

1. **Connection Pooling**
   ```go
   // Current: New connection per API call
   httpReq, _ := http.NewRequest(...)
   
   // Better: Reuse HTTP client
   client := &http.Client{
       Timeout: 30 * time.Second,
       Transport: &http.Transport{
           MaxIdleConns: 100,
           MaxConnsPerHost: 10,
       },
   }
   ```

2. **Message Batching**
   ```go
   // Current: Send each message immediately
   stream.Send(msg)
   
   // Better: Batch multiple messages
   batch := make([]*ChatMessage, 0, 10)
   // Collect messages, send in batch
   ```

3. **Caching**
   ```go
   // Cache user lookups
   userCache := make(map[string]*User)
   lastFetch := time.Now()
   ```

---

## 10. Testing Approaches

### Unit Test Example

```go
func TestLoginSuccess(t *testing.T) {
    // Setup
    sm := models.NewStateManager()
    user := &models.User{
        ID:       "user-1",
        Username: "alice",
        Password: auth.HashPassword("password123"),
    }
    sm.AddUser(user)
    
    service := NewUserService(sm)
    
    // Test
    resp, err := service.Login(context.Background(), &pb.LoginRequest{
        Username: "alice",
        Password: "password123",
    })
    
    // Assert
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if resp.Token == "" {
        t.Error("Expected token, got empty")
    }
}

func TestLoginWrongPassword(t *testing.T) {
    // Setup
    // ... (same as above)
    
    // Test
    _, err := service.Login(context.Background(), &pb.LoginRequest{
        Username: "alice",
        Password: "wrongpassword",
    })
    
    // Assert
    if err == nil {
        t.Error("Expected error, got nil")
    }
    if status.Code(err) != codes.Unauthenticated {
        t.Errorf("Expected Unauthenticated, got %v", status.Code(err))
    }
}
```

### Integration Test

```go
func TestChatFlow(t *testing.T) {
    // Start server
    server := startTestServer()
    defer server.Stop()
    
    // Connect clients
    conn, _ := grpc.Dial(":50051", grpc.WithInsecure())
    userClient := pb.NewUserServiceClient(conn)
    roomClient := pb.NewRoomServiceClient(conn)
    chatClient := pb.NewChatServiceClient(conn)
    
    // Login
    loginResp, _ := userClient.Login(ctx, &pb.LoginRequest{
        Username: "alice",
        Password: "password123",
    })
    
    // Create room
    createResp, _ := roomClient.CreateRoom(ctx, &pb.CreateRoomRequest{
        Token:    loginResp.Token,
        RoomName: "Test Room",
    })
    
    // Chat
    stream, _ := chatClient.Chat(ctx)
    stream.Send(&pb.ChatMessage{Content: loginResp.Token})
    stream.Send(&pb.ChatMessage{RoomId: createResp.RoomId})
    stream.Send(&pb.ChatMessage{Content: "Hello"})
    
    // Verify
    msg, _ := stream.Recv()
    if msg.Content != "Hello" {
        t.Error("Expected message 'Hello'")
    }
}
```

---

## Conclusion

Implementasi ini mendemonstrasikan:
- ✅ Proper gRPC service structure
- ✅ Thread-safe state management
- ✅ JWT authentication
- ✅ Bidirectional streaming
- ✅ Error handling patterns
- ✅ Concurrency best practices
- ✅ External API integration
- ✅ Protocol buffers usage

Untuk production, tambahkan:
- ✅ Persistent database
- ✅ Comprehensive logging
- ✅ Monitoring & metrics
- ✅ Load balancing
- ✅ TLS encryption
- ✅ Rate limiting
- ✅ Unit & integration tests
