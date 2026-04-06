# Quick Start Guide - gRPC AI Chat

Panduan cepat untuk menjalankan aplikasi dalam 5 menit!

## ⚡ 5-Minute Setup

### Step 1: Install Dependencies (1 min)

```bash
# Clone project
cd grpc-ai-chat

# Download Go modules
go mod download

# Install protoc tools (if not installed)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Step 2: Generate Proto Files (1 min)

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/chat.proto
```

**Atau gunakan Makefile (jika installed):**
```bash
make proto
```

### Step 3: Setup Environment (30 sec)

```bash
# Copy example env file
cp .env.example .env

# Edit .env dan add Grok API key (optional)
# GROK_API_KEY=your-key-here
```

### Step 4: Run Server (1 min)

```bash
# Terminal 1
cd cmd/server
go run main.go
```

Expected output:
```
🚀 Starting gRPC AI Chat Server...
👤 Created user: alice
👤 Created user: bob
👤 Created user: charlie
✅ Server listening at [::]:50051
```

### Step 5: Run Client (1 min)

```bash
# Terminal 2
cd cmd/client
go run main.go

# In client, login:
login alice password123

# Create room:
create_room "Hello World" "Test room" true

# List rooms (note the room ID):
list_rooms

# Join room (use the ID from list):
join_room <room-id>

# Start chat:
chat

# Type messages, press Enter, type "exit" to quit chat
```

## 🎯 Common Workflows

### Workflow 1: Single User Testing

```
Terminal 1: Server
$ go run cmd/server/main.go
✅ Server listening at [::]:50051

Terminal 2: Client
$ go run cmd/client/main.go
chat> login alice password123
✅ Login successful! Welcome alice

chat> create_room "My Room" "Testing" true
✅ Room created! ID: abc123def456

chat> join_room abc123def456
✅ joined successfully

chat> chat
📢 Starting chat...
(type messages here)
```

### Workflow 2: Two-User Chat

```
Terminal 1: Server
$ go run cmd/server/main.go

Terminal 2: Alice
$ go run cmd/client/main.go
chat> login alice password123
✅ Login successful!
chat> create_room "Dev Team" "Let's code" true
✅ Room created! ID: xyz789
chat> join_room xyz789
✅ joined successfully
chat> chat
(waiting for messages)

Terminal 3: Bob
$ go run cmd/client/main.go
chat> login bob password123
✅ Login successful!
chat> list_rooms
📋 Available Rooms (1):
   [xyz789...] Dev Team (1 members) 🤖
chat> join_room xyz789
✅ joined successfully
(ℹ️  [SYSTEM] bob telah bergabung dengan chat - appears in Alice's chat)
chat> chat
📝 You> Hi Alice!
💬 [bob] Hi Alice!  (appears in Alice's chat)
🤖 [AI] Hello! I'm happy to help...  (AI response appears for both)
```

### Workflow 3: Test All Commands

```bash
# Register new user
chat> register john password123 "John Developer"
✅ Registration successful! User ID: ...

# Login
chat> login john password123
✅ Login successful!

# Create room with AI disabled
chat> create_room "General Chat" "No AI here" false
✅ Room created! ID: ...

# Create another room with AI
chat> create_room "AI Room" "With AI" true
✅ Room created! ID: ...

# List rooms
chat> list_rooms
📋 Available Rooms (2):
   [id1...] General Chat (1 members)
   [id2...] AI Room (1 members) 🤖

# Join first room
chat> join_room <id1>
✅ joined successfully

# Chat in non-AI room (no AI responses)
chat> chat
📝 You> Hello everyone
(only user messages, no AI responses)
exit

# Leave room
chat> leave_room
✅ berhasil keluar dari room

# Join second room with AI
chat> join_room <id2>
✅ joined successfully

# Chat with AI enabled (will get AI responses)
chat> chat
📝 You> What is gRPC?
💬 [john] What is gRPC?
🤖 [AI] gRPC is a modern open source high performance RPC framework...
```

## 🔧 Makefile Commands

Jika Makefile installed, bisa gunakan:

```bash
# Setup development environment
make setup

# Generate proto files
make proto

# Download dependencies
make deps

# Build binaries
make build-server
make build-client
make build

# Run server
make run-server

# Run client
make run-client

# Clean build artifacts
make clean

# Format code
make fmt

# Run tests
make test
```

## 🐛 Troubleshooting

### Problem: Proto compilation fails
```bash
# Solution: Verify protoc installed
protoc --version

# If not installed, install protoc from:
# https://github.com/protocolbuffers/protobuf/releases

# Then reinstall Go plugins:
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Problem: "connection refused"
```bash
# Solution: Ensure server is running first
# Terminal 1: cd cmd/server && go run main.go
# Terminal 2: cd cmd/client && go run main.go
```

### Problem: "token tidak valid" when joining room
```bash
# Solution: Make sure you logged in first
login alice password123
# Then try again
join_room <room_id>
```

### Problem: AI responses not working
```bash
# Solution: 
# 1. Set GROK_API_KEY environment variable
export GROK_API_KEY=your-key-here

# 2. Make sure room was created with AI enabled
create_room "AI Room" "" true
```

### Problem: Can't see other user's messages
```bash
# Solution: Both users must be in same room
# Terminal 2 & 3 must both do:
join_room <same-room-id>
chat
```

## 📊 Example Output

### Server Output:
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

🔌 gRPC Services:
   - UserService (Login, Register)
   - RoomService (Create, List, Join, Leave)
   - ChatService (Bidirectional Streaming Chat)

⚙️  AI Integration: Grok API
```

### Client Terminal 1 (Alice):
```
🎉 Welcome to gRPC AI Chat!

alice@chat> login alice password123
✅ Login successful! Welcome alice
   Token expires in 86400 seconds

alice@chat> create_room "Tech Talk" "Discuss tech stuff" true
✅ Room created! ID: 550e8400-e29b-41d4-a716-446655440000
   🤖 AI is enabled in this room

alice@chat> join_room 550e8400-e29b-41d4-a716-446655440000
✅ joined successfully
   Room: Tech Talk (1 members)
   🤖 This room has AI enabled!

alice@chat(room)> chat
📢 Starting chat... (type 'exit' to leave)
ℹ️  [SYSTEM] alice telah bergabung dengan chat
📝 You> Hello Bob, can you help me understand async programming?
💬 [alice] Hello Bob, can you help me understand async programming?
🤖 [AI] Async programming allows multiple operations to run concurrently...
```

### Client Terminal 2 (Bob):
```
🎉 Welcome to gRPC AI Chat!

bob@chat> login bob password123
✅ Login successful! Welcome bob
   Token expires in 86400 seconds

bob@chat> list_rooms
📋 Available Rooms (1):
   [550e8400...] Tech Talk (1 members) 🤖
       Description: Discuss tech stuff

bob@chat> join_room 550e8400-e29b-41d4-a716-446655440000
✅ joined successfully
   Room: Tech Talk (2 members)
   🤖 This room has AI enabled!

bob@chat(room)> chat
📢 Starting chat... (type 'exit' to leave)
ℹ️  [SYSTEM] bob telah bergabung dengan chat
💬 [alice] Hello Bob, can you help me understand async programming?
🤖 [AI] Async programming allows multiple operations to run concurrently...
📝 You> Sure! Async is super useful. The AI explanation is spot on.
💬 [bob] Sure! Async is super useful. The AI explanation is spot on.
```

## 📚 Next Steps

1. **Explore the code:**
   - Read `ARCHITECTURE.md` for system design
   - Check `README.md` for detailed documentation
   - Look at `proto/chat.proto` for service definitions

2. **Customize:**
   - Add more AI models beyond Grok
   - Implement persistent storage (PostgreSQL)
   - Add user profiles and avatars
   - Implement direct messaging

3. **Deploy:**
   - Build Docker image
   - Deploy to cloud (GCP, AWS, Azure)
   - Setup load balancer
   - Use managed database

4. **Test:**
   - Run multiple concurrent clients
   - Test with high message volume
   - Measure latency and throughput

## 🎓 Learning Resources

- [gRPC Basics](https://grpc.io/docs/languages/go/basics/)
- [Protocol Buffers Guide](https://protobuf.dev/getting-started/gotutorial/)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [JWT.io - Learn JWT](https://jwt.io/introduction)

## 💡 Tips & Tricks

1. **Quick Testing with Multiple Users:**
   ```bash
   # Use multiple terminals with different users
   Terminal 1: server
   Terminal 2: login alice
   Terminal 3: login bob
   Terminal 4: login charlie
   ```

2. **Debug Mode:**
   ```bash
   # Add debug output by modifying code
   fmt.Printf("[DEBUG] ...")
   ```

3. **Monitor Network:**
   ```bash
   # View gRPC traffic (if using grpcurl)
   grpcurl -plaintext list localhost:50051
   ```

4. **Performance Testing:**
   ```bash
   # Send many messages
   # Measure response time
   # Monitor memory usage
   ```

## ✅ Checklist

Before moving to production:

- [ ] Replace in-memory with database
- [ ] Implement bcrypt for passwords
- [ ] Add TLS/SSL encryption
- [ ] Setup rate limiting
- [ ] Add comprehensive logging
- [ ] Implement monitoring (Prometheus)
- [ ] Add distributed tracing (Jaeger)
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Document API with examples
- [ ] Setup CI/CD pipeline
- [ ] Load test the system
- [ ] Security audit
- [ ] Performance optimization

---

**Ready to build something amazing?** 🚀

For more help, check `README.md` and `ARCHITECTURE.md`!
