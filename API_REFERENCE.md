# API Reference - gRPC AI Chat

Dokumentasi lengkap semua service dan RPC methods yang tersedia.

## Table of Contents
1. [User Service](#user-service)
2. [Room Service](#room-service)
3. [Chat Service](#chat-service)
4. [Error Codes](#error-codes)
5. [Examples](#examples)

---

## User Service

Service untuk autentikasi dan manajemen user.

### Login (Unary RPC)

**Request:**
```protobuf
message LoginRequest {
  string username = 1;  // Username user (required)
  string password = 2;  // Password user (required)
}
```

**Response:**
```protobuf
message LoginResponse {
  string token = 1;         // JWT token untuk requests berikutnya
  string user_id = 2;       // Unique user identifier
  string username = 3;      // Username yang login
  int32 expires_in = 4;     // Token expiry time dalam detik
}
```

**Error Codes:**
- `INVALID_ARGUMENT` - Username atau password kosong
- `UNAUTHENTICATED` - Username atau password salah
- `INTERNAL` - Gagal membuat token

**Example:**
```bash
# CLI
login alice password123

# Response
✅ Login successful! Welcome alice
   Token expires in 86400 seconds
```

---

### Register (Unary RPC)

**Request:**
```protobuf
message RegisterRequest {
  string username = 1;   // Username baru (required, harus unik)
  string password = 2;   // Password (required, minimum 6 karakter)
  string full_name = 3;  // Nama lengkap (required)
}
```

**Response:**
```protobuf
message RegisterResponse {
  bool success = 1;      // Status registrasi
  string message = 2;    // Pesan status
  string user_id = 3;    // User ID yang baru dibuat
}
```

**Error Codes:**
- `INVALID_ARGUMENT` - Input field kosong atau password < 6 karakter
- `ALREADY_EXISTS` - Username sudah terdaftar
- `INTERNAL` - Gagal membuat user

**Example:**
```bash
# CLI
register john password123 "John Doe"

# Response
✅ Registration successful! User ID: 550e8400-e29b-41d4-a716-446655440000
```

**Validation Rules:**
- Username: Min 1 karakter, Max 50 karakter, alphanumeric
- Password: Min 6 karakter
- Full Name: Min 1 karakter, Max 100 karakter

---

## Room Service

Service untuk manajemen chat rooms.

### CreateRoom (Unary RPC)

**Request:**
```protobuf
message CreateRoomRequest {
  string token = 1;           // JWT token dari login (required)
  string room_name = 2;       // Nama room (required)
  string description = 3;     // Deskripsi room (optional)
  bool is_ai_enabled = 4;     // Enable AI responses (optional, default: false)
}
```

**Response:**
```protobuf
message CreateRoomResponse {
  bool success = 1;           // Status create room
  string room_id = 2;         // UUID room yang dibuat
  string room_name = 3;       // Nama room
  string message = 4;         // Pesan status
}
```

**Error Codes:**
- `INVALID_ARGUMENT` - Room name kosong
- `UNAUTHENTICATED` - Token tidak valid atau session expired
- `INTERNAL` - Gagal membuat room

**Example:**
```bash
# CLI
create_room "Dev Team" "For developers" true

# Response
✅ Room created! ID: 550e8400-e29b-41d4-a716-446655440000
   🤖 AI is enabled in this room
```

**Notes:**
- User yang create room otomatis menjadi member
- Room akan dihapus otomatis saat tidak ada members
- AI enabled berarti AI akan respond di chat room ini

---

### ListRooms (Unary RPC)

**Request:**
```protobuf
message ListRoomsRequest {
  string token = 1;  // JWT token dari login (required)
}
```

**Response:**
```protobuf
message ListRoomsResponse {
  repeated RoomInfo rooms = 1;  // List of rooms
  int32 total = 2;              // Total rooms count
}

message RoomInfo {
  string room_id = 1;           // UUID room
  string room_name = 2;         // Nama room
  string description = 3;       // Deskripsi room
  int32 member_count = 4;       // Jumlah members aktif
  bool is_ai_enabled = 5;       // AI enabled flag
  int64 created_at = 6;         // Timestamp created (Unix)
}
```

**Error Codes:**
- `UNAUTHENTICATED` - Token tidak valid atau session expired
- `INTERNAL` - Gagal fetch rooms

**Example:**
```bash
# CLI
list_rooms

# Response
📋 Available Rooms (2):
   [550e8400...] Dev Team (5 members) 🤖
       Description: For developers
   [650e8400...] General Chat (3 members)
       Description: General discussion
```

**Notes:**
- Menampilkan semua public rooms
- Member count update real-time
- Sorted by created_at (newest first)

---

### JoinRoom (Unary RPC)

**Request:**
```protobuf
message JoinRoomRequest {
  string token = 1;    // JWT token dari login (required)
  string room_id = 2;  // ID room yang ingin diikuti (required)
}
```

**Response:**
```protobuf
message JoinRoomResponse {
  bool success = 1;             // Status join
  string message = 2;           // Pesan status
  RoomInfo room_info = 3;       // Info room yang diikuti
}
```

**Error Codes:**
- `UNAUTHENTICATED` - Token tidak valid atau session expired
- `NOT_FOUND` - Room tidak ditemukan
- `INTERNAL` - Gagal join room

**Example:**
```bash
# CLI
join_room 550e8400-e29b-41d4-a716-446655440000

# Response
✅ joined successfully
   Room: Dev Team (6 members)
   🤖 This room has AI enabled!
```

**Notes:**
- User dapat join multiple rooms
- Membership tracking per user-room
- System message: "{username} telah bergabung dengan chat"

---

### LeaveRoom (Unary RPC)

**Request:**
```protobuf
message LeaveRoomRequest {
  string token = 1;    // JWT token dari login (required)
  string room_id = 2;  // ID room yang ingin ditinggalkan (required)
}
```

**Response:**
```protobuf
message LeaveRoomResponse {
  bool success = 1;      // Status leave
  string message = 2;    // Pesan status
}
```

**Error Codes:**
- `UNAUTHENTICATED` - Token tidak valid atau session expired
- `NOT_FOUND` - Room tidak ditemukan
- `INTERNAL` - Gagal leave room

**Example:**
```bash
# CLI
leave_room

# Response
✅ berhasil keluar dari room
```

**Notes:**
- Room dihapus otomatis jika tidak ada members
- System message: "{username} telah keluar dari chat"
- Disconnect dari chat stream otomatis

---

## Chat Service

Service untuk real-time bidirectional streaming chat.

### Chat (Bidirectional Streaming RPC)

**Request Stream:**
```protobuf
message ChatMessage {
  string message_id = 1;    // Unique message ID (server-generated)
  string user_id = 2;       // User ID (optional di request)
  string username = 3;      // Username (optional di request)
  string room_id = 4;       // Room ID untuk join (required di awal)
  string content = 5;       // Message content
  int64 timestamp = 6;      // Message timestamp (server-generated)
  string message_type = 7;  // "user", "ai", "system"
  bool is_ai_response = 8;  // Flag untuk AI response
  string error_message = 9; // Error message (jika ada)
}
```

**Response Stream:**
```protobuf
// Same structure as ChatMessage
message ChatMessage {
  // Server akan return populated message_id, timestamp, dll
}
```

**Flow:**

1. **Connect & Authenticate**
   ```
   Client: Send ChatMessage with content = JWT_TOKEN
   Server: Validate token, respond with authenticated message
   ```

2. **Join Room**
   ```
   Client: Send ChatMessage with room_id = ROOM_ID
   Server: Add user to room members
   Server: Broadcast system message to all users in room
   ```

3. **Send Messages**
   ```
   Client: Send ChatMessage with content = "Hello everyone"
   Server: Broadcast to all users in same room
   Server: If room AI-enabled, generate AI response async
   ```

4. **Receive AI Response**
   ```
   Server: Call Grok API with user message
   Server: Receive AI response
   Server: Broadcast AI message to all users
   ```

5. **Disconnect**
   ```
   Client: Close stream (Ctrl+C or "exit")
   Server: Remove user from room members
   Server: Broadcast leave system message
   ```

**Error Codes:**
- `UNAUTHENTICATED` - Invalid token
- `NOT_FOUND` - Room tidak ditemukan
- `INTERNAL` - Server error

**Message Types:**
- `"user"` - User-generated message
- `"ai"` - AI-generated response
- `"system"` - System notification (join/leave)

**Example Chat Session:**

```
Client A (alice):
  │
  ├─ Send: {content: "jwt_token_123"}
  ├─ Receive: {message_type: "system", content: "authenticated"}
  │
  ├─ Send: {room_id: "room_123"}
  ├─ Receive: {message_type: "system", content: "alice telah bergabung dengan chat"}
  │
  ├─ Send: {content: "What is gRPC?"}
  ├─ Receive: {message_type: "user", username: "alice", content: "What is gRPC?"}
  │ (broadcast ke semua clients di room_123)
  │
  └─ Receive: {message_type: "ai", username: "AI Assistant", 
               content: "gRPC is a high-performance RPC framework..."}

Client B (bob):
  │
  ├─ Send: {content: "jwt_token_456"}
  ├─ Receive: {message_type: "system", content: "authenticated"}
  │
  ├─ Send: {room_id: "room_123"}
  ├─ Receive: {message_type: "system", content: "bob telah bergabung dengan chat"}
  ├─ Receive: {message_type: "system", content: "alice telah bergabung dengan chat"}
  │ (alice's join message)
  │
  ├─ Receive: {message_type: "user", username: "alice", content: "What is gRPC?"}
  │ (message dari alice)
  │
  └─ Receive: {message_type: "ai", username: "AI Assistant", 
               content: "gRPC is a high-performance RPC framework..."}
```

**Notes:**
- Messages broadcast instantly ke semua users di room
- AI response generate async (non-blocking)
- Keep-alive: Stream stays open sampai client disconnect
- Each stream = separate goroutine di server

---

## Error Codes

Complete list of gRPC error codes yang digunakan:

| Code | Name | Usage | HTTP |
|------|------|-------|------|
| 0 | OK | Success response | 200 |
| 3 | INVALID_ARGUMENT | Input validation error | 400 |
| 5 | NOT_FOUND | Resource tidak ditemukan | 404 |
| 6 | ALREADY_EXISTS | Resource sudah ada | 409 |
| 16 | UNAUTHENTICATED | Invalid/expired token | 401 |
| 13 | INTERNAL | Server error | 500 |

**Examples:**

```go
// Input validation
return nil, status.Error(codes.InvalidArgument, "username harus diisi")

// Authentication
return nil, status.Error(codes.Unauthenticated, "token tidak valid")

// Not found
return nil, status.Error(codes.NotFound, "room tidak ditemukan")

// Already exists
return nil, status.Error(codes.AlreadyExists, "username sudah terdaftar")

// Server error
return nil, status.Error(codes.Internal, "gagal membuat token")
```

---

## Examples

### Example 1: Full Authentication & Chat Flow

```bash
# Terminal 1: Server
$ go run cmd/server/main.go
✅ Server listening at [::]:50051

# Terminal 2: User Alice
$ go run cmd/client/main.go

# Register
> register alice password123 "Alice Johnson"
✅ Registration successful! User ID: user-001

# Login
> login alice password123
✅ Login successful! Welcome alice

# Create room with AI
> create_room "AI Chat" "Chat with AI" true
✅ Room created! ID: room-001
   🤖 AI is enabled in this room

# Join room
> join_room room-001
✅ joined successfully
   Room: AI Chat (1 members)
   🤖 This room has AI enabled!

# Start chatting
> chat
📢 Starting chat...
📝 You> What is machine learning?
💬 [alice] What is machine learning?
🤖 [AI] Machine learning is a subset of artificial intelligence...
📝 You> exit
👋 Left chat
```

### Example 2: Multi-User Chat

```bash
# Terminal 1: Server (same as above)

# Terminal 2: Alice
> login alice password123
> create_room "Dev Discussion" "Talk about development" false
  ✅ Room created! ID: dev-room-001
> join_room dev-room-001
> chat
(waiting for messages from Bob)

# Terminal 3: Bob
> login bob password123
> list_rooms
📋 Available Rooms (1):
   [dev-room-001] Dev Discussion (1 members)

> join_room dev-room-001
✅ joined successfully
(ℹ️  [SYSTEM] bob telah bergabung dengan chat - appears in Alice's terminal)

> chat
📝 You> Hi Alice!
💬 [bob] Hi Alice! (appears in both terminals)

# Back in Terminal 2 (Alice):
💬 [bob] Hi Alice!
📝 You> Hi Bob! How are you?
💬 [alice] Hi Bob! How are you? (appears in Bob's terminal)
```

### Example 3: Error Handling

```bash
# Wrong password
> login alice wrongpassword
❌ Login failed: rpc error: code = Unauthenticated desc = username atau password salah

# Duplicate username
> register alice password456 "Alice New"
❌ Registration failed: rpc error: code = AlreadyExists desc = username sudah terdaftar

# Invalid room
> join_room nonexistent-room
❌ Join room failed: rpc error: code = NotFound desc = room tidak ditemukan

# Missing token
(If send request without valid token)
❌ Error: rpc error: code = Unauthenticated desc = token tidak valid
```

### Example 4: Token Management

```
Token Lifecycle:

1. User login
   > login alice password123
   ✅ Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ✅ Expires in 86400 seconds (24 hours)

2. Token stored in client
   - Used for all subsequent requests
   - Attached to room operations
   - Sent as first message in chat stream

3. Token validation
   - Server checks signature
   - Server checks expiry time
   - Returns error if invalid/expired

4. Token refresh
   - If expired, must login again
   - New token generated
   - Old session invalidated
```

### Example 5: AI-Powered Room

```bash
# Create room with AI
> create_room "AI Assistant" "Get help from AI" true
✅ Room created! ID: ai-room-001

> join_room ai-room-001
> chat

# User sends question
📝 You> How to sort array in Go?

# Server processes
1. Broadcast user message
2. Detect room AI-enabled = true
3. Spawn goroutine:
   - Call Grok API
   - Pass message: "How to sort array in Go?"
   - Wait for response
4. Broadcast AI response when ready

# User receives response
💬 [alice] How to sort array in Go?
🤖 [AI] In Go, you can sort arrays using the sort package:

    package main
    import "sort"
    
    arr := []int{3, 1, 2}
    sort.Ints(arr)
    // arr is now [1, 2, 3]
    
    For custom types, use sort.Slice()
```

---

## Status Codes Reference

| Status | Code | When Used | Retry |
|--------|------|-----------|-------|
| OK | 0 | Success | N/A |
| CANCELLED | 1 | Request cancelled | No |
| UNKNOWN | 2 | Unknown error | No |
| INVALID_ARGUMENT | 3 | Bad input | No |
| DEADLINE_EXCEEDED | 4 | Timeout | Yes (may help) |
| NOT_FOUND | 5 | Resource missing | No |
| ALREADY_EXISTS | 6 | Duplicate resource | No |
| PERMISSION_DENIED | 7 | Access denied | No |
| RESOURCE_EXHAUSTED | 8 | Quota exceeded | Yes (after delay) |
| FAILED_PRECONDITION | 9 | Bad state | Depends |
| ABORTED | 10 | Transaction aborted | Yes |
| OUT_OF_RANGE | 11 | Index out of range | No |
| UNIMPLEMENTED | 12 | Not implemented | No |
| INTERNAL | 13 | Server error | Yes |
| UNAVAILABLE | 14 | Service unavailable | Yes |
| DATA_LOSS | 15 | Data loss | Depends |
| UNAUTHENTICATED | 16 | No auth | Yes (re-login) |

---

## Timeout Recommendations

```
Service Call | Recommended Timeout | Notes
-------------|--------------------|-----------
Login        | 5 seconds          | Fast operation
Register     | 5 seconds          | Fast operation
CreateRoom   | 5 seconds          | Fast operation
ListRooms    | 5 seconds          | Can be large
JoinRoom     | 5 seconds          | Fast operation
Chat Stream  | 0 (indefinite)     | Long-lived connection
```

---

## Rate Limiting (Future)

Currently no rate limiting. For production, implement:

```
Per User:
  - 1000 messages/hour
  - 100 room creates/day
  - 10 login attempts/minute

Per IP:
  - 10000 requests/hour
  - 100 registrations/day
```

---

## Conclusion

This API reference covers all available gRPC methods and their usage. For implementation details, see `ARCHITECTURE.md`. For quick start, see `QUICKSTART.md`.
