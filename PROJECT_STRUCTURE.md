# Project Structure Complete Overview

## Directory Tree

```
grpc-ai-chat/
│
├── cmd/                          # Command line applications
│   ├── server/
│   │   └── main.go              # gRPC server entry point
│   │       • Initializes services
│   │       • Creates gRPC server
│   │       • Registers service handlers
│   │       • Listens on port 50051
│   │       • Loads default users
│   │
│   └── client/
│       └── main.go              # Interactive CLI client
│           • Connects to gRPC server
│           • Handles user commands
│           • Manages interactive loop
│           • Bidirectional streaming chat
│
├── internal/                     # Internal application code
│   │
│   ├── models/
│   │   └── models.go            # Data models & state manager
│   │       • User struct
│   │       • Session struct
│   │       • Room & RoomMember structs
│   │       • StateManager with thread-safe maps
│   │       • RWMutex for concurrency control
│   │
│   ├── services/
│   │   ├── user_service.go      # User service implementation
│   │   │   • Login(username, password)
│   │   │   • Register(username, password, fullname)
│   │   │   • Token generation
│   │   │   • Session management
│   │   │
│   │   ├── room_service.go      # Room service implementation
│   │   │   • CreateRoom(token, name, description, ai_enabled)
│   │   │   • ListRooms(token)
│   │   │   • JoinRoom(token, room_id)
│   │   │   • LeaveRoom(token, room_id)
│   │   │   • Member management
│   │   │
│   │   └── chat_service.go      # Chat service implementation
│   │       • Chat(stream ChatMessage) - Bidirectional streaming
│   │       • Message broadcasting
│   │       • AI response generation
│   │       • Stream management
│   │       • System notifications
│   │
│   ├── auth/
│   │   └── auth.go              # Authentication utilities
│   │       • JWT token generation
│   │       • Token validation
│   │       • User ID extraction from token
│   │       • UUID generation
│   │       • Password hashing (simplified)
│   │       • Password verification
│   │
│   └── ai/
│       └── grok.go              # Grok AI API client
│           • GrokClient struct
│           • GenerateResponse(message, context)
│           • API request/response handling
│           • Error handling
│           • Health check
│
├── proto/                        # Protocol buffer definitions
│   └── chat.proto               # gRPC service & message definitions
│       • UserService
│       •   ├─ Login (Unary RPC)
│       •   └─ Register (Unary RPC)
│       • RoomService
│       •   ├─ CreateRoom (Unary RPC)
│       •   ├─ ListRooms (Unary RPC)
│       •   ├─ JoinRoom (Unary RPC)
│       •   └─ LeaveRoom (Unary RPC)
│       • ChatService
│       •   └─ Chat (Bidirectional Streaming)
│       • Message definitions (LoginRequest, ChatMessage, etc.)
│
├── pkg/
│   └── api/
│       └── chat/
│           └── v1/              # Generated proto files (auto-generated)
│               ├── chat.pb.go   # Protocol buffer message definitions
│               └── chat_grpc.pb.go # gRPC service client/server stubs
│
├── .env.example                 # Environment variables template
│   • GROK_API_KEY
│   • SERVER_HOST
│   • SERVER_PORT
│   • JWT_SECRET
│   • TOKEN_DURATION
│
├── go.mod                       # Go module definition
│   • Module path: grpc-ai-chat
│   • Go version: 1.21
│   • Dependencies: gRPC, protobuf, JWT, UUID
│
├── go.sum                       # Go dependencies checksum (auto-generated)
│
├── Makefile                     # Build automation
│   • make proto - Generate proto files
│   • make build-server - Build server binary
│   • make build-client - Build client binary
│   • make run-server - Run server
│   • make run-client - Run client
│   • make clean - Clean build artifacts
│   • make test - Run tests
│
├── README.md                    # Main documentation
│   • Features overview
│   • Installation guide
│   • Usage examples
│   • Environment setup
│   • Troubleshooting
│   • API reference
│
├── QUICKSTART.md               # Quick start guide
│   • 5-minute setup
│   • Common workflows
│   • Example outputs
│   • Troubleshooting tips
│
├── ARCHITECTURE.md             # Detailed architecture documentation
│   • System design
│   • Component details
│   • Data flow diagrams
│   • Concurrency patterns
│   • Error handling strategy
│   • Performance considerations
│   • Security considerations
│   • Testing strategy
│   • Deployment architecture
│
├── PROJECT_STRUCTURE.md        # This file
│   • Directory tree with descriptions
│   • File-by-file overview
│   • Code organization
│   • Dependencies mapping
│
└── bin/                        # Build output (created by make build)
    ├── server                  # Compiled server binary
    └── client                  # Compiled client binary
```

## File Descriptions

### Core Application Files

#### `cmd/server/main.go` (87 lines)
```
Purpose: Server entry point
Responsibilities:
  - Initialize StateManager for in-memory storage
  - Create default test users (alice, bob, charlie)
  - Instantiate all service handlers
  - Create gRPC server
  - Register services with gRPC
  - Start TCP listener on port 50051
  - Block and serve requests
```

#### `cmd/client/main.go` (444 lines)
```
Purpose: Interactive CLI client
Responsibilities:
  - Connect to gRPC server
  - Provide interactive command prompt
  - Handle user input parsing
  - Implement 7+ commands
  - Manage bidirectional chat stream
  - Display formatted output
  - Handle multiple concurrent operations
```

### Internal Packages

#### `internal/models/models.go` (189 lines)
```
Data Structures:
  - User: id, username, password, fullname, created_at
  - Session: user_id, token, expires_at, created_at
  - Room: id, name, description, created_by, members, is_ai_enabled
  - RoomMember: user_id, username, joined_at, stream
  - StateManager: users map, sessions map, rooms map

Thread Safety:
  - RWMutex on StateManager
  - RWMutex on Room for member operations
  - All methods handle locks properly
```

#### `internal/services/user_service.go` (106 lines)
```
RPC Methods:
  1. Login(username, password)
     - Validate input
     - Lookup user
     - Verify password
     - Generate JWT token
     - Create session
     - Return token + metadata
  
  2. Register(username, password, fullname)
     - Validate input length
     - Check duplicate username
     - Hash password
     - Create user
     - Store in state
     - Return success/error

Error Codes:
  - InvalidArgument: Missing/invalid input
  - Unauthenticated: Wrong password
  - AlreadyExists: Username taken
  - Internal: Server error
```

#### `internal/services/room_service.go` (185 lines)
```
RPC Methods:
  1. CreateRoom(token, name, description, ai_enabled)
     - Authenticate with token
     - Validate session
     - Generate room ID
     - Create room with metadata
     - Add creator as first member
     - Store in state
  
  2. ListRooms(token)
     - Authenticate
     - Get all rooms
     - Convert to proto messages
     - Return list with member counts
  
  3. JoinRoom(token, room_id)
     - Authenticate
     - Find room
     - Add user to members
     - Broadcast join event
  
  4. LeaveRoom(token, room_id)
     - Authenticate
     - Find room
     - Remove user from members
     - Delete empty rooms
     - Broadcast leave event
```

#### `internal/services/chat_service.go` (263 lines)
```
Bidirectional Streaming:
  - Chat(stream ChatMessage) -> stream ChatMessage
  
Flow:
  1. Client connects with token
  2. Server validates token
  3. Client sends room_id to join
  4. Server broadcasts join message
  5. Loop: Receive -> Process -> Broadcast
  
Features:
  - Message broadcasting to all room members
  - Concurrent stream handling
  - AI response generation (async)
  - System messages
  - Error propagation
```

#### `internal/auth/auth.go` (109 lines)
```
Functions:
  - GenerateToken(userID, username) -> string
  - ValidateToken(token) -> claims, error
  - GetUserIDFromToken(token) -> string
  - GetUsernameFromToken(token) -> string
  - GenerateID() -> UUID string
  - HashPassword(password) -> string
  - VerifyPassword(hash, password) -> bool
  
Token Format:
  {
    "user_id": "uuid",
    "username": "alice",
    "exp": timestamp,
    "iat": timestamp
  }
```

#### `internal/ai/grok.go` (167 lines)
```
GrokClient:
  - Constructor: NewGrokClient()
  - GenerateResponse(message, context) -> string
  - HealthCheck() -> error
  
API Integration:
  - HTTP POST to Grok API
  - Bearer token authentication
  - JSON request/response
  - Error handling
```

### Configuration & Documentation

#### `proto/chat.proto` (124 lines)
```
Services:
  1. UserService
     - Login(LoginRequest) -> LoginResponse [Unary]
     - Register(RegisterRequest) -> RegisterResponse [Unary]
  
  2. RoomService
     - CreateRoom(CreateRoomRequest) -> CreateRoomResponse [Unary]
     - ListRooms(ListRoomsRequest) -> ListRoomsResponse [Unary]
     - JoinRoom(JoinRoomRequest) -> JoinRoomResponse [Unary]
     - LeaveRoom(LeaveRoomRequest) -> LeaveRoomResponse [Unary]
  
  3. ChatService
     - Chat(ChatMessage) -> ChatMessage [Bidirectional Streaming]

Messages:
  - LoginRequest/Response
  - RegisterRequest/Response
  - CreateRoomRequest/Response
  - ListRoomsRequest/Response
  - RoomInfo
  - JoinRoomRequest/Response
  - LeaveRoomRequest/Response
  - ChatMessage (for streaming)
```

#### `go.mod` (20 lines)
```
Module: grpc-ai-chat
Go: 1.21

Direct Dependencies:
  - google.golang.org/grpc v1.56.2
  - google.golang.org/protobuf v1.31.0
  - github.com/golang-jwt/jwt/v5 v5.0.0
  - github.com/google/uuid v1.3.1
  - github.com/joho/godotenv v1.5.1
```

#### `Makefile` (104 lines)
```
Targets:
  - help: Show available commands
  - proto: Generate proto files
  - deps: Download dependencies
  - build-server: Build server binary
  - build-client: Build client binary
  - run-server: Run server
  - run-client: Run client
  - run-both: Run server & client (tmux)
  - test: Run tests
  - clean: Remove build artifacts
  - fmt: Format code
  - lint: Run golangci-lint
```

## Code Organization Principles

### Separation of Concerns
- **cmd/** - Application entry points
- **internal/models/** - Data models
- **internal/services/** - Business logic
- **internal/auth/** - Authentication
- **internal/ai/** - External integrations
- **proto/** - Service contracts

### Layering
```
CLI Layer (cmd/client)
    ↓
gRPC Service Layer (internal/services)
    ↓
Business Logic Layer (models + auth + ai)
    ↓
Data Layer (in-memory state manager)
```

### Dependency Flow
```
main.go
  ├─→ services/user_service.go
  ├─→ services/room_service.go
  ├─→ services/chat_service.go
  └─→ models/models.go
        ├─→ auth/auth.go
        ├─→ ai/grok.go
        └─→ sync/mutex
```

## Generated Files (Auto-created)

### `pkg/api/chat/v1/chat.pb.go` (Auto-generated from proto)
```
Generated Message Types:
  - LoginRequest, LoginResponse
  - RegisterRequest, RegisterResponse
  - CreateRoomRequest, CreateRoomResponse
  - ListRoomsRequest, ListRoomsResponse, RoomInfo
  - JoinRoomRequest, JoinRoomResponse
  - LeaveRoomRequest, LeaveRoomResponse
  - ChatMessage
  
Contains:
  - Marshal/Unmarshal methods
  - Message descriptors
  - Proto reflection data
```

### `pkg/api/chat/v1/chat_grpc.pb.go` (Auto-generated from proto)
```
Generated Service Interfaces:
  - UserServiceServer (server interface)
  - UserServiceClient (client interface)
  - RoomServiceServer (server interface)
  - RoomServiceClient (client interface)
  - ChatServiceServer (server interface)
  - ChatServiceClient (client interface)
  
Contains:
  - Service registration
  - Method stubs
  - Client constructors
```

## Total Lines of Code

```
Core Application:
  - cmd/server/main.go: 87
  - cmd/client/main.go: 444
  - internal/models/models.go: 189
  - internal/services/user_service.go: 106
  - internal/services/room_service.go: 185
  - internal/services/chat_service.go: 263
  - internal/auth/auth.go: 109
  - internal/ai/grok.go: 167
  ─────────────────────────
  Total: 1,350 lines

Proto Definition:
  - proto/chat.proto: 124 lines

Configuration:
  - go.mod: 20 lines
  - .env.example: 13 lines
  - Makefile: 104 lines

Documentation:
  - README.md: 408 lines
  - QUICKSTART.md: 436 lines
  - ARCHITECTURE.md: 444 lines
  - PROJECT_STRUCTURE.md: (this file)

Total with documentation: ~3,000 lines
```

## Compilation & Execution Flow

### Compilation
```
protoc
  ↓
Generates: pkg/api/chat/v1/*.pb.go
Generates: pkg/api/chat/v1/*_grpc.pb.go
  ↓
go build cmd/server/main.go
  ├─ Imports: models, services, auth
  └─ Output: bin/server
  
go build cmd/client/main.go
  └─ Output: bin/client
```

### Runtime
```
Server Start:
  1. main() in cmd/server
  2. NewStateManager()
  3. CreateDefaultUsers()
  4. NewUserService(stateManager)
  5. NewRoomService(stateManager)
  6. NewChatService(stateManager)
  7. gRPC.RegisterServices()
  8. listener.Serve()

Client Start:
  1. main() in cmd/client
  2. NewClientApp()
  3. grpc.Dial(localhost:50051)
  4. InteractiveLoop()
  5. HandleCommand()
  6. ServiceClient.Method()
```

## Directory Creation Quick Reference

```bash
# Full structure
grpc-ai-chat/
├── cmd/server/
├── cmd/client/
├── internal/models/
├── internal/services/
├── internal/auth/
├── internal/ai/
├── proto/
├── pkg/api/chat/v1/
├── bin/

# Create with mkdir
mkdir -p cmd/server cmd/client internal/models internal/services internal/auth internal/ai proto pkg/api/chat/v1 bin
```

---

This project demonstrates:
✅ Proper Go project structure
✅ gRPC service implementation
✅ Protocol buffers usage
✅ JWT authentication
✅ Concurrency patterns
✅ Error handling
✅ API integration
✅ Bidirectional streaming
✅ Multi-client support
✅ In-memory state management
