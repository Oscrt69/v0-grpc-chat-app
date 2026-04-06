# gRPC AI Chat Application

Aplikasi chat real-time dengan AI menggunakan gRPC dan Go, dilengkapi dengan integrasi Grok API.

## 📋 Fitur

### Services
1. **User Service** - Autentikasi dan manajemen user
   - `Login` (Unary RPC) - Login dengan username dan password
   - `Register` (Unary RPC) - Registrasi user baru

2. **Room Service** - Manajemen chat rooms
   - `CreateRoom` (Unary RPC) - Buat room baru
   - `ListRooms` (Unary RPC) - Daftar semua rooms
   - `JoinRoom` (Unary RPC) - Bergabung ke room
   - `LeaveRoom` (Unary RPC) - Keluar dari room

3. **Chat Service** - Real-time chat
   - `Chat` (Bidirectional Streaming) - Chat real-time dengan broadcast ke semua users

### Features
- ✅ Multi-user support
- ✅ Bidirectional streaming untuk real-time chat
- ✅ AI-powered responses menggunakan Grok API
- ✅ In-memory state management
- ✅ JWT authentication dengan token-based access
- ✅ Proper gRPC error handling dengan status codes
- ✅ Interactive CLI client untuk multi-user
- ✅ System messages untuk user join/leave events

## 🏗️ Arsitektur Sistem

```
┌─────────────────────────────────────────────┐
│           Client CLI (Multi-user)           │
└────────────────────┬────────────────────────┘
                     │ gRPC
        ┌────────────┴────────────┐
        ▼                         ▼
    ┌──────────────┐      ┌──────────────┐
    │ User Service │      │Room Service  │
    └──────────────┘      └──────────────┘
        │                         │
        └────────────┬────────────┘
                     ▼
            ┌──────────────────┐
            │  Chat Service    │
            │ (Bidirectional   │
            │  Streaming)      │
            └────────┬─────────┘
                     ▼
        ┌────────────────────────┐
        │  In-Memory State Mgr   │
        │ - Users, Sessions      │
        │ - Rooms, Members       │
        └────────┬───────────────┘
                 ▼
        ┌────────────────────┐
        │   Grok AI API      │
        └────────────────────┘
```

## 📁 Struktur Direktori

```
grpc-ai-chat/
├── cmd/
│   ├── server/
│   │   └── main.go              # Server entry point
│   └── client/
│       └── main.go              # CLI client entry point
├── internal/
│   ├── models/
│   │   └── models.go            # Data models dan state manager
│   ├── services/
│   │   ├── user_service.go      # User service implementation
│   │   ├── room_service.go      # Room service implementation
│   │   └── chat_service.go      # Chat service implementation
│   ├── auth/
│   │   └── auth.go              # JWT token handling
│   └── ai/
│       └── grok.go              # Grok API client
├── proto/
│   └── chat.proto               # Protocol buffers definition
├── pkg/
│   └── api/
│       └── chat/
│           └── v1/              # Generated proto files
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Installation & Setup

### Prerequisites
- Go 1.21+
- Protocol Buffers compiler (`protoc`)
- grpc-go: `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`

### 1. Setup Project

```bash
# Clone atau download project
cd grpc-ai-chat

# Initialize go modules
go mod download

# Generate proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/chat.proto
```

### 2. Set Environment Variables

```bash
# Create .env file
cat > .env << EOF
GROK_API_KEY=your-grok-api-key-here
EOF

# Load environment
export $(cat .env | xargs)
```

### 3. Run Server

```bash
cd cmd/server
go run main.go
```

Output:
```
🚀 Starting gRPC AI Chat Server...
👤 Created user: alice
👤 Created user: bob
👤 Created user: charlie
✅ Server listening at [::]:50051
📝 Default users:
   - Username: alice, Password: password123
   - Username: bob, Password: password123
   - Username: charlie, Password: password123
```

### 4. Run Client (di terminal lain)

```bash
cd cmd/client
go run main.go
```

## 💬 Usage Guide

### Terminal 1: Server
```bash
go run cmd/server/main.go
```

### Terminal 2: Client Alice
```bash
go run cmd/client/main.go
# login alice password123
# create_room "Dev Chat" "Technical discussion" true
# list_rooms
# join_room <room_id>
# chat
```

### Terminal 3: Client Bob
```bash
go run cmd/client/main.go
# login bob password123
# list_rooms
# join_room <room_id>
# chat
```

### Interactive Commands

```
🎉 Welcome to gRPC AI Chat!
📝 Commands:
   register <username> <password> <fullname>  - Register user
   login <username> <password>                - Login
   create_room <room_name> [desc] [ai]      - Create room
   list_rooms                                  - List rooms
   join_room <room_id>                        - Join room
   chat                                        - Start chatting
   leave_room                                  - Leave room
   logout                                      - Logout
   exit                                        - Exit
```

## 📚 Example Workflow

```
1. Register user baru (atau gunakan default users)
   > register john password123 "John Doe"

2. Login
   > login john password123
   ✅ Login successful! Welcome john

3. Buat room dengan AI enabled
   > create_room "Tech Support" "AI-powered help" true
   ✅ Room created! ID: abc123...
   🤖 AI is enabled in this room

4. List available rooms
   > list_rooms
   📋 Available Rooms (1):
      [abc123...] Tech Support (1 members) 🤖

5. Join room (dari client lain)
   > join_room abc123...
   ✅ joined successfully

6. Start chatting
   > chat
   📢 Starting chat... (type 'exit' to leave)
   ℹ️  [SYSTEM] bob telah bergabung dengan chat
   📝 You> Hello, can you help me with this bug?
   💬 [bob] Hello!
   🤖 [AI] I'd be happy to help! Can you describe the issue you're experiencing?
```

## 🔐 Authentication & Security

### Token-based Auth
- User login menghasilkan JWT token
- Token valid selama 24 jam
- Setiap service call memerlukan token valid
- Session disimpan di in-memory state

### Error Handling
Menggunakan gRPC status codes:
- `InvalidArgument` - Input validation error
- `Unauthenticated` - Invalid token atau session expired
- `NotFound` - Resource tidak ditemukan
- `AlreadyExists` - Resource sudah ada (e.g., duplicate username)
- `Internal` - Server error

## 🤖 Grok AI Integration

### Fitur
- Generate response otomatis untuk messages di AI-enabled rooms
- Context-aware responses berdasarkan user message
- Non-blocking async processing

### Setup
1. Dapatkan Grok API key dari https://x.ai/
2. Set environment variable:
   ```bash
   export GROK_API_KEY=your-key-here
   ```

### Example
```
User: "How to sort array in Go?"
AI: "In Go, you can use the sort package. Here's an example:
    
    package main
    import "sort"
    
    arr := []int{3,1,2}
    sort.Ints(arr)
    // arr is now [1,2,3]"
```

## 🔄 Bidirectional Streaming Details

### Chat Flow
1. Client connect dan send auth token
2. Server validate token
3. Client send room ID untuk join
4. Semua clients di room menerima system message
5. Client send chat messages
6. Server broadcast ke semua clients di room
7. Jika AI enabled, generate AI response
8. Broadcast AI response ke semua clients

### Message Types
- `user` - User message
- `ai` - AI response
- `system` - System message (join/leave events)

## 📊 State Management

### In-Memory Storage
```go
// Users map
Users: map[userID]*User

// Sessions map
Sessions: map[token]*Session

// Rooms map
Rooms: map[roomID]*Room
  └─ Members: map[userID]*RoomMember
```

### Thread Safety
- Semua maps dilindungi dengan RWMutex
- Concurrent read/write operations aman
- Lock hanya saat operasi state change

## 🧪 Testing

### Test Login
```bash
# Terminal 1
go run cmd/server/main.go

# Terminal 2
go run cmd/client/main.go
> login alice password123
✅ Login successful! Welcome alice
```

### Test Multi-user Chat
```bash
# Terminal 1: Server
go run cmd/server/main.go

# Terminal 2: Alice
go run cmd/client/main.go
# login alice password123
# create_room "Test" "" true
# (note the room ID)
# join_room <room_id>
# chat

# Terminal 3: Bob
go run cmd/client/main.go
# login bob password123
# join_room <same_room_id>
# chat
# (type messages)

# Terminal 2: Alice dapat melihat messages dari Bob
# Terminal 3: Bob dapat melihat messages dari Alice
# Keduanya dapat melihat AI responses
```

## 🚨 Troubleshooting

### Proto compilation error
```bash
protoc --version  # Pastikan installed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Connection refused
- Pastikan server sudah running di port 50051
- Check firewall settings
- Ubah `localhost:50051` ke IP address jika connecting dari machine lain

### Grok API errors
- Verify GROK_API_KEY environment variable set
- Check API key valid dan tidak expired
- Verify internet connection untuk Grok API call

### Token expired
- Login lagi untuk mendapatkan token baru
- Token valid selama 24 jam

## 📝 Implementation Notes

### Password Hashing
Untuk production, implement bcrypt:
```go
import "golang.org/x/crypto/bcrypt"

hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// store hash di database
```

### Database Integration
Untuk production, replace in-memory state dengan database:
- PostgreSQL dengan GORM
- MongoDB
- Firebase Realtime Database

### Monitoring
Tambahkan:
- Structured logging (slog, zap)
- Metrics collection (Prometheus)
- Tracing (Jaeger)

## 📖 References

- [gRPC Documentation](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://protobuf.dev/)
- [JWT Authentication](https://jwt.io/)
- [Grok API Documentation](https://x.ai/)

## 📄 License

MIT License

## 👨‍💻 Author

Created for gRPC AI Chat Tutorial
