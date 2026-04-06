# Development Guide

## Setup Development Environment

### Prerequisites
- Go 1.21 or later
- Protocol Buffers compiler (protoc)
- Make (optional, for using Makefile)

### Quick Setup

#### On Linux/macOS:
```bash
# Run setup script
bash SETUP.sh

# Or manual setup
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/chat.proto
go mod download
go mod tidy
```

#### On Windows:
```cmd
SETUP.bat
```

## Project Structure

```
.
├── cmd/
│   ├── server/          # Server executable
│   │   └── main.go     # Server entry point
│   └── client/          # Client CLI executable
│       └── main.go     # Client entry point
├── proto/              # Protocol Buffer definitions
│   └── chat.proto      # Service definitions
├── internal/           # Internal packages
│   ├── models/         # Data models and state
│   ├── services/       # gRPC service implementations
│   ├── auth/          # Authentication utilities
│   └── ai/            # AI integration (Grok)
├── go.mod             # Module file
├── go.sum             # Dependency checksums
├── Makefile           # Build commands
└── .env.example       # Environment template
```

## Building

### Using Makefile
```bash
# Build both server and client
make build

# Build only server
make build-server

# Build only client
make build-client

# Clean binaries
make clean

# Run server
make run-server

# Run client
make run-client

# Generate proto files
make generate
```

### Manual Build
```bash
# Generate proto files first
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/chat.proto

# Build server
go build -o bin/server ./cmd/server

# Build client
go build -o bin/client ./cmd/client
```

## Running

### Start Server
```bash
cd cmd/server
go run main.go
# Server starts on localhost:50051
```

### Start Client
```bash
cd cmd/client
go run main.go
# Client connects to localhost:50051
```

### Multiple Clients
Open separate terminals for each client:
```bash
# Terminal 1
cd cmd/client && go run main.go

# Terminal 2
cd cmd/client && go run main.go

# Terminal 3
cd cmd/client && go run main.go
```

## Common Commands in Client

```
help                                    # Show all commands
register <username> <password>          # Register new user
login <username> <password>             # Login user
create_room <name> <description> <ai>   # Create room (ai: true/false)
list_rooms                              # List all rooms
join_room <room-id>                     # Join a room
leave_room                              # Leave current room
chat                                    # Enter chat mode
logout                                  # Logout and exit
```

## Testing

### Manual Testing

1. **Register Users**
```bash
> register alice password123
> register bob pass456
> register charlie pass789
```

2. **Create Room**
```bash
> create_room "General Chat" "Main discussion room" true
# Output: Room created with ID: xyz123
```

3. **Multiple Clients Chat**
```
Client 1: login alice password123
Client 1: join_room xyz123
Client 1: chat

Client 2: login bob pass456
Client 2: join_room xyz123
Client 2: chat

# Now both can chat in real-time
```

4. **Test AI Integration**
- Enable AI when creating room: `ai: true`
- Send message starting with "ai:" to get AI response
- Example: `ai: what is gRPC?`

### Error Testing

1. **Unauthenticated Access**
```bash
# Try to create room without login
> create_room "test" "" false
# Expected: Error - unauthenticated
```

2. **Invalid Token**
- Modify token in memory
- Try to send message
- Expected: Error - invalid token

3. **Room Not Found**
```bash
> join_room invalid-id
# Expected: Error - room not found
```

## Code Style

### Go Conventions
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use `golint` for linting
- Use `go vet` for vetting

### Naming
- Package names: lowercase, short
- Functions/methods: PascalCase for exported, camelCase for unexported
- Variables: camelCase
- Constants: PascalCase

### Comments
- Every public function should have a comment
- Comment explains "why" not "what"
- Use `//` for single line, `/* */` for multi-line

Example:
```go
// CreateRoom creates a new chat room with the given parameters.
// It returns the room ID or an error if creation fails.
func (rs *RoomService) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
    // implementation
}
```

## Proto File Development

### Modifying Services

1. Edit `proto/chat.proto`
2. Regenerate proto files:
```bash
make generate
# or
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/chat.proto
```
3. Update service implementations

### Adding New RPC Method

1. Add to `.proto` file:
```proto
service ChatService {
  rpc NewMethod(Request) returns (Response);
}
```

2. Regenerate:
```bash
make generate
```

3. Implement in service:
```go
func (cs *ChatService) NewMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    // implementation
}
```

## State Management

### In-Memory State
- All data stored in memory
- Protected by RWMutex
- Lost on server restart

### State Structure
```
Users:     map[id]User
Sessions:  map[token]Session
Rooms:     map[id]Room
Streams:   map[room][]*Stream
```

### Thread Safety
- Always lock before accessing shared data
- Use RLock() for reads
- Use Lock() for writes

## Error Handling

### gRPC Status Codes
- `Unauthenticated` - Invalid token or not logged in
- `InvalidArgument` - Bad input parameters
- `NotFound` - Room or user not found
- `AlreadyExists` - Duplicate user or room
- `Internal` - Server error

### Error Pattern
```go
if err != nil {
    return nil, status.Error(codes.InvalidArgument, "message")
}
```

## AI Integration (Grok)

### Configuration
```bash
# In .env
GROK_API_KEY=your-api-key-here
```

### Usage
- Enabled per-room (set in CreateRoom)
- Non-blocking: async call in separate goroutine
- Message prefix: "ai:" to trigger response
- Example: `ai: explain what is gRPC`

### Testing
```bash
# In client
> join_room <ai-enabled-room>
> chat
> ai: hello
# Should get AI response
```

## Debugging

### Enable Debug Logging
Edit service files and add:
```go
log.Printf("[DEBUG] %v", variable)
```

### Common Issues

1. **Connection refused**
   - Server not running
   - Wrong port in client

2. **Proto file not found**
   - Run `make generate` or `bash SETUP.sh`

3. **Token invalid**
   - Token expired (24-hour TTL)
   - Server restarted (in-memory sessions lost)

4. **AI not responding**
   - Check GROK_API_KEY in .env
   - Check internet connection
   - Message must start with "ai:"

## Performance Tips

1. **Room Broadcasting**
   - Message sent to all clients in room
   - Consider limiting room size for large deployments

2. **Goroutine Cleanup**
   - Chat streams properly closed on disconnect
   - Resources released to prevent leaks

3. **Memory Usage**
   - In-memory storage increases with users/messages
   - Consider implementing message history limit

## Production Deployment

### Considerations
- Replace in-memory state with database
- Implement message persistence
- Add authentication token rotation
- Scale with load balancer
- Use TLS for secure communication
- Implement rate limiting
- Add monitoring and logging

### Next Steps
- See `IMPLEMENTATION_DETAILS.md` for architecture details
- See `README.md` for complete reference
- See `TROUBLESHOOTING.md` for common issues
