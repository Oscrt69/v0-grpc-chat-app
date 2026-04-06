# Architecture Documentation

## System Overview

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │          Interactive CLI Client (Multi)             │ │
│  │  - Register/Login                                   │ │
│  │  - Room Management (Create/List/Join/Leave)        │ │
│  │  - Real-time Chat (Bidirectional Streaming)        │ │
│  └────────────────────────────────────────────────────┘ │
└──────────────────┬───────────────────────────────────────┘
                   │ gRPC over HTTP/2
        ┌──────────┴──────────┐
        │                     │
┌───────▼──────────┐   ┌──────▼───────────┐
│  SERVICE LAYER   │   │  SERVICE LAYER   │
├──────────────────┤   ├──────────────────┤
│ UserService      │   │ RoomService      │
│ - Login (RPC)    │   │ - CreateRoom     │
│ - Register (RPC) │   │ - ListRooms      │
└──────────────────┘   │ - JoinRoom       │
                       │ - LeaveRoom      │
                       └──────┬───────────┘
                              │
                    ┌─────────▼──────────┐
                    │  ChatService       │
                    │  - Chat (BiStream) │
                    └─────────┬──────────┘
                              │
        ┌─────────────────────┴─────────────────────┐
        │                                           │
┌───────▼──────────────────┐      ┌────────────────▼───┐
│  STATE MANAGER LAYER     │      │   AI LAYER         │
│ (In-Memory)              │      │  (Grok API)        │
├──────────────────────────┤      ├────────────────────┤
│ Users Map                │      │ GenerateResponse   │
│ Sessions Map             │      │ HealthCheck        │
│ Rooms Map                │      │                    │
│ Broadcast Manager        │      │ HTTP Calls to      │
│ Thread-safe (RWMutex)    │      │ Grok API Endpoint  │
└──────────────────────────┘      └────────────────────┘
```

## Component Details

### 1. User Service (`internal/services/user_service.go`)

**Responsibilities:**
- User authentication (login)
- User registration
- Session management
- Token generation and validation

**RPC Methods:**
- `Login(username, password) -> (token, user_id, username, expires_in)`
- `Register(username, password, full_name) -> (success, message, user_id)`

**Error Codes:**
- `InvalidArgument` - Missing or invalid input
- `Unauthenticated` - Wrong credentials
- `AlreadyExists` - Username taken
- `Internal` - Server error

### 2. Room Service (`internal/services/room_service.go`)

**Responsibilities:**
- Create chat rooms
- List available rooms
- Join rooms
- Leave rooms
- Room member management

**RPC Methods:**
- `CreateRoom(token, room_name, description, is_ai_enabled) -> (success, room_id)`
- `ListRooms(token) -> (rooms[])`
- `JoinRoom(token, room_id) -> (success, room_info)`
- `LeaveRoom(token, room_id) -> (success)`

**Features:**
- Auto-delete empty rooms
- Room creator tracking
- Member count tracking
- AI enablement per room

### 3. Chat Service (`internal/services/chat_service.go`)

**Responsibilities:**
- Bidirectional streaming chat
- Message broadcasting
- User authentication in stream
- Room joining in stream
- AI response generation
- User join/leave notifications

**RPC Method:**
- `Chat(stream ChatMessage) -> stream ChatMessage` (Bidirectional)

**Stream Protocol:**
1. Client sends token for auth
2. Client sends room_id to join
3. Server validates and broadcasts join message
4. Client sends chat messages
5. Server broadcasts to all connected clients
6. If room AI-enabled, generate AI response
7. Broadcast AI response

**Message Types:**
- `user` - User message
- `ai` - AI-generated response
- `system` - System notifications

### 4. State Manager (`internal/models/models.go`)

**Data Structures:**

```go
// User
type User struct {
    ID       string
    Username string
    Password string
    FullName string
    Created  time.Time
}

// Session
type Session struct {
    UserID    string
    Token     string
    ExpiresAt time.Time
    Created   time.Time
}

// Room
type Room struct {
    ID          string
    Name        string
    Description string
    CreatedBy   string
    CreatedAt   time.Time
    Members     map[string]*RoomMember
    IsAIEnabled bool
    mu          sync.RWMutex  // Thread safety
}

// StateManager
type StateManager struct {
    Users    map[string]*User
    Sessions map[string]*Session
    Rooms    map[string]*Room
    mu       sync.RWMutex
}
```

**Thread Safety:**
- All maps protected by RWMutex
- Read operations use RLock
- Write operations use Lock
- Safe for concurrent access from multiple goroutines

### 5. Auth Module (`internal/auth/auth.go`)

**Functions:**
- `GenerateToken(userID, username)` - Create JWT token
- `ValidateToken(token)` - Verify JWT signature and expiry
- `GenerateID()` - Create UUID for resources
- `HashPassword(password)` - Hash password (simplified)
- `VerifyPassword(hash, password)` - Verify password match
- `GetUserIDFromToken(token)` - Extract user ID
- `GetUsernameFromToken(token)` - Extract username

**Token Format:**
```json
{
  "user_id": "uuid-here",
  "username": "alice",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**Security Notes:**
- JWT with HS256 signing
- 24-hour token expiry
- Session tracking in StateManager
- (Production: Use bcrypt for password hashing)

### 6. AI Module (`internal/ai/grok.go`)

**Grok Client Features:**
- HTTP POST requests to Grok API
- Message composition with context
- Non-blocking async processing
- Error handling and recovery

**API Integration:**
- Endpoint: `https://api.x.ai/v1/chat/completions`
- Model: `grok-2`
- Max tokens: 256
- Temperature: 0.7

**Flow:**
```
User Message
    ↓
Chat Service detects AI-enabled room
    ↓
Spawn goroutine to call Grok API (non-blocking)
    ↓
Grok generates response
    ↓
Broadcast AI message to all clients
```

## Data Flow Examples

### 1. User Login Flow

```
Client: Login Request (username, password)
  ↓
UserService.Login()
  ├─ GetUserByUsername()
  ├─ VerifyPassword()
  ├─ GenerateToken()
  ├─ CreateSession()
  └─ Return (token, user_id, expires_in)
Client: Store token for future requests
```

### 2. Create Room Flow

```
Client: CreateRoom Request (token, name, description, ai_enabled)
  ↓
RoomService.CreateRoom()
  ├─ ValidateToken()
  ├─ CheckSessionValid()
  ├─ GenerateRoomID()
  ├─ CreateRoom struct
  ├─ AddRoomCreatorAsMember()
  ├─ StateManager.AddRoom()
  └─ Return (room_id, success)
Client: Receive room ID
```

### 3. Chat Streaming Flow

```
Client A: Connect Stream
  ├─ Send token
  ├─ Send room_id
  └─ Start listening for messages

Client B: Send Message
  ├─ Validate token & session
  ├─ Validate room
  ├─ Broadcast message to all clients in room
  ├─ Check if room AI-enabled
  └─ If yes, async call Grok API

Grok API: Generate Response
  ├─ Receive user message
  ├─ Generate contextual AI response
  └─ Return response

Server: Broadcast AI Response
  ├─ Receive Grok response
  ├─ Create AI message
  └─ Send to all connected clients in room

Client A & B: Receive Messages
  ├─ Display user message
  ├─ Display AI response
  └─ Continue listening
```

## Concurrency Patterns

### 1. Goroutine per Stream Connection
```go
// Each Chat stream runs in separate goroutine
// Multiple streams can run concurrently
stream.Recv()  // Blocks until message received
```

### 2. Broadcasting with Goroutines
```go
// Iterate through all connected streams
// Send message to each (non-blocking with buffered channel)
for _, s := range streams {
    s.Send(msg)  // Fire and forget
}
```

### 3. Async AI Processing
```go
// Generate AI response in separate goroutine
// Doesn't block message processing
go cs.generateAndBroadcastAIResponse(roomID, message)
```

### 4. Thread-Safe State Management
```go
// All state access protected by RWMutex
sm.mu.Lock()
defer sm.mu.Unlock()
// Modify state here
```

## Error Handling Strategy

### gRPC Status Codes Used:
```
codes.InvalidArgument    - Input validation failed
codes.Unauthenticated    - Invalid/expired token
codes.NotFound           - Resource doesn't exist
codes.AlreadyExists      - Resource already exists
codes.Internal           - Server-side error
codes.PermissionDenied   - Access denied (can be added)
```

### Example Error Response:
```go
return nil, status.Error(codes.Unauthenticated, "token tidak valid")
```

## Performance Considerations

### Scalability Limits:
- **Current:** In-memory storage - suitable for ~10k users, ~100 concurrent connections
- **Improvements needed for production:**
  - Replace in-memory with database (PostgreSQL, MongoDB)
  - Add connection pooling
  - Implement message queuing (Redis, RabbitMQ)
  - Add rate limiting
  - Implement caching layer

### Memory Usage:
- Each user: ~500 bytes
- Each session: ~200 bytes
- Each room: ~1KB base + 100 bytes per member
- Each message in stream: broadcasted immediately (not stored)

### Throughput:
- Messages/second: Limited by Grok API response time
- Concurrent users: Limited by goroutine overhead
- Typical: 100-1000 concurrent users per server instance

## Security Considerations

### Current Implementation:
- ✅ Token-based authentication
- ✅ Session validation
- ✅ Input validation
- ⚠️ Simplified password hashing (use bcrypt in production)
- ⚠️ JWT secret hardcoded (use environment variable)
- ⚠️ No SSL/TLS (add in production)
- ⚠️ No rate limiting

### Production Improvements:
1. Use bcrypt for password hashing
2. TLS/SSL for gRPC connections
3. Rate limiting per IP/user
4. API key validation for Grok
5. Input sanitization
6. Audit logging
7. Database encryption
8. CORS if exposed via HTTP

## Testing Strategy

### Unit Tests:
- Auth token generation/validation
- State manager operations
- Error handling

### Integration Tests:
- Login flow
- Room creation and joining
- Chat message broadcasting
- AI response generation

### Load Tests:
- Multiple concurrent connections
- Message broadcast latency
- AI response generation performance

## Deployment Architecture

### Development:
```
Single machine
├─ Server (port 50051)
└─ Client(s)
```

### Production:
```
Load Balancer (port 50051)
├─ Server Instance 1
├─ Server Instance 2
└─ Server Instance N
    ↓
Shared Database (PostgreSQL)
    ↓
Message Queue (Redis/RabbitMQ)
    ↓
Grok API (External)
```

## Future Enhancements

1. **Persistence:**
   - Replace in-memory with PostgreSQL
   - Store chat history
   - User profiles with avatar

2. **Advanced Features:**
   - Direct messaging between users
   - File sharing
   - Message reactions/threading
   - User roles (admin, moderator)

3. **Performance:**
   - Message caching
   - Connection pooling
   - Horizontal scaling with database sync

4. **AI Features:**
   - Message summarization
   - Sentiment analysis
   - Content moderation
   - Multi-language support

5. **Monitoring:**
   - Prometheus metrics
   - Jaeger distributed tracing
   - Structured logging (slog)
   - Health check endpoints
