# API Testing Guide

## Overview
This guide covers testing the gRPC chat application including unit tests, integration tests, and manual testing scenarios.

## Test Scenarios

### 1. User Service Testing

#### 1.1 Register User
```
POST /UserService/Register
Body: {"username": "alice", "password": "password123"}

✓ Success Response:
{
  "user_id": "user-uuid-1",
  "username": "alice",
  "created_at": timestamp
}

✗ Error - Duplicate User:
{
  "code": 6 (ALREADY_EXISTS),
  "message": "user with username 'alice' already exists"
}

✗ Error - Invalid Input:
{
  "code": 3 (INVALID_ARGUMENT),
  "message": "username cannot be empty"
}
```

#### 1.2 Login User
```
POST /UserService/Login
Body: {"username": "alice", "password": "password123"}

✓ Success Response:
{
  "user_id": "user-uuid-1",
  "username": "alice",
  "token": "jwt-token-here",
  "expires_at": timestamp
}

✗ Error - User Not Found:
{
  "code": 5 (NOT_FOUND),
  "message": "user not found"
}

✗ Error - Invalid Password:
{
  "code": 16 (UNAUTHENTICATED),
  "message": "invalid password"
}
```

### 2. Room Service Testing

#### 2.1 Create Room
```
POST /RoomService/CreateRoom
Headers: {"Authorization": "Bearer jwt-token"}
Body: {
  "name": "General Chat",
  "description": "Main discussion room",
  "enable_ai": true
}

✓ Success Response:
{
  "room_id": "room-uuid-1",
  "name": "General Chat",
  "description": "Main discussion room",
  "enable_ai": true,
  "created_by": "user-uuid-1",
  "created_at": timestamp
}

✗ Error - Unauthorized:
{
  "code": 16 (UNAUTHENTICATED),
  "message": "invalid token"
}

✗ Error - Empty Room Name:
{
  "code": 3 (INVALID_ARGUMENT),
  "message": "room name cannot be empty"
}
```

#### 2.2 List Rooms
```
GET /RoomService/ListRooms
Headers: {"Authorization": "Bearer jwt-token"}

✓ Success Response:
{
  "rooms": [
    {
      "room_id": "room-uuid-1",
      "name": "General Chat",
      "description": "Main discussion room",
      "enable_ai": true,
      "member_count": 3,
      "created_at": timestamp
    },
    {
      "room_id": "room-uuid-2",
      "name": "Tech Discussion",
      "description": "For tech topics",
      "enable_ai": false,
      "member_count": 5,
      "created_at": timestamp
    }
  ]
}
```

#### 2.3 Join Room
```
POST /RoomService/JoinRoom
Headers: {"Authorization": "Bearer jwt-token"}
Body: {"room_id": "room-uuid-1"}

✓ Success Response:
{
  "room_id": "room-uuid-1",
  "name": "General Chat",
  "member_count": 4,
  "joined_at": timestamp
}

✗ Error - Room Not Found:
{
  "code": 5 (NOT_FOUND),
  "message": "room not found"
}

✗ Error - Already in Room:
{
  "code": 6 (ALREADY_EXISTS),
  "message": "user is already in this room"
}
```

#### 2.4 Leave Room
```
POST /RoomService/LeaveRoom
Headers: {"Authorization": "Bearer jwt-token"}
Body: {"room_id": "room-uuid-1"}

✓ Success Response:
{
  "room_id": "room-uuid-1",
  "left_at": timestamp
}

✗ Error - Not in Room:
{
  "code": 5 (NOT_FOUND),
  "message": "user is not in this room"
}
```

### 3. Chat Service Testing (Bidirectional Streaming)

#### 3.1 Chat Stream Connection

```
STREAM /ChatService/Chat
Headers: {"Authorization": "Bearer jwt-token"}

Client sends message:
{
  "type": "message",
  "room_id": "room-uuid-1",
  "content": "Hello everyone!"
}

Server broadcasts to all clients in room:
{
  "type": "user_message",
  "room_id": "room-uuid-1",
  "user_id": "user-uuid-1",
  "username": "alice",
  "content": "Hello everyone!",
  "timestamp": timestamp
}

Other clients in room see:
{
  "type": "user_message",
  "room_id": "room-uuid-1",
  "user_id": "user-uuid-1",
  "username": "alice",
  "content": "Hello everyone!",
  "timestamp": timestamp
}
```

#### 3.2 System Notifications

When user joins:
```
{
  "type": "user_joined",
  "room_id": "room-uuid-1",
  "username": "bob",
  "timestamp": timestamp
}
```

When user leaves:
```
{
  "type": "user_left",
  "room_id": "room-uuid-1",
  "username": "bob",
  "timestamp": timestamp
}
```

#### 3.3 AI Response

User sends:
```
{
  "type": "message",
  "room_id": "room-uuid-1",
  "content": "ai: What is gRPC?"
}
```

If room has AI enabled, server sends:
```
{
  "type": "ai_response",
  "room_id": "room-uuid-1",
  "content": "gRPC is a high-performance RPC framework...",
  "timestamp": timestamp
}
```

#### 3.4 Stream Error - Unauthorized

```
STREAM /ChatService/Chat
Headers: {"Authorization": "Bearer invalid-token"}

Server response:
{
  "code": 16 (UNAUTHENTICATED),
  "message": "invalid token",
  "details": "token validation failed"
}
```

## Manual Testing Steps

### Test 1: Basic Multi-Client Chat
```bash
# Terminal 1 - Server
cd cmd/server && go run main.go

# Terminal 2 - Client 1
cd cmd/client && go run main.go
> register alice password123
> login alice password123
> create_room "Test Room" "Testing" false
# Output: Room ID: xyz123

# Terminal 3 - Client 2
cd cmd/client && go run main.go
> register bob password456
> login bob password456
> join_room xyz123
> chat

# Terminal 2 - Continue
> join_room xyz123
> chat
# Now both can chat in real-time
```

### Test 2: AI Integration
```bash
# Terminal 1 - Server (already running)

# Terminal 2 - Client 1
> create_room "AI Room" "For AI testing" true
# Output: Room ID: ai123

# Terminal 3 - Client 2
> join_room ai123
> chat

# Terminal 2 - Continue
> join_room ai123
> chat
> ai: What is Go programming language?
# Should receive AI response
```

### Test 3: Error Handling
```bash
# Test 3.1 - Invalid Login
> login invalid_user password123
# Expected: Error - user not found

# Test 3.2 - Invalid Room
> join_room invalid-room-id
# Expected: Error - room not found

# Test 3.3 - Unauthorized Create Room
# (without login first)
> create_room "Test" "" false
# Expected: Error - unauthenticated
```

### Test 4: Multi-User Broadcast
```bash
# Terminal 2 - Client 1
> login alice password123
> create_room "Broadcast" "" false
# Output: Room ID: bcast1
> join_room bcast1
> chat

# Terminal 3 - Client 2
> login bob password456
> join_room bcast1
> chat

# Terminal 4 - Client 3
> login charlie password789
> join_room bcast1
> chat

# In Terminal 2, send message
> Hello everyone!
# Should appear in Terminal 3 and 4

# Test cross-terminal messaging
```

### Test 5: User Join/Leave Notifications
```bash
# Terminal 2 - Client 1
> login alice password123
> create_room "Notify Test" "" false
# Output: Room ID: notify1
> join_room notify1
> chat
# Terminal 2 shows: "alice joined the room"

# Terminal 3 - Client 2
> login bob password456
> join_room notify1
> chat
# Terminal 2 shows: "bob joined the room"
# Terminal 3 shows: "bob joined the room"

# Terminal 3
> leave_room
# Terminal 2 shows: "bob left the room"
```

### Test 6: Token Expiration
```bash
# Login with token that expires in 24 hours
> login alice password123
# Token: eyJ...

# Wait (simulate after 24 hours)
# Or modify token and try to use it
> chat
# Expected: Error - invalid token after expiration
```

## Using grpcurl for Testing

### Install grpcurl
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Test User Service
```bash
# List services
grpcurl -plaintext localhost:50051 list

# Call Register
grpcurl -plaintext \
  -d '{"username":"test","password":"pass123"}' \
  localhost:50051 UserService/Register

# Call Login
grpcurl -plaintext \
  -d '{"username":"test","password":"pass123"}' \
  localhost:50051 UserService/Login
```

### Test Room Service
```bash
# Call CreateRoom
grpcurl -plaintext \
  -H "authorization: Bearer <token>" \
  -d '{"name":"Test","description":"","enable_ai":false}' \
  localhost:50051 RoomService/CreateRoom

# Call ListRooms
grpcurl -plaintext \
  -H "authorization: Bearer <token>" \
  localhost:50051 RoomService/ListRooms

# Call JoinRoom
grpcurl -plaintext \
  -H "authorization: Bearer <token>" \
  -d '{"room_id":"room-id-here"}' \
  localhost:50051 RoomService/JoinRoom
```

## Load Testing

### Simple Load Test Script
```bash
#!/bin/bash
# test_load.sh

for i in {1..10}; do
  # Register user
  go run cmd/client/main.go <<EOF
register user$i pass$i
EOF
done

# This registers 10 concurrent users
```

### Performance Metrics
- Message latency: < 100ms
- Broadcast time: < 500ms
- Connection establishment: < 1s
- Token validation: < 10ms

## Debugging

### Enable Debug Mode
In service files, add logging:
```go
log.Printf("[DEBUG] Function called with: %v", param)
```

### Check Server Logs
```bash
# Run server with debug
cd cmd/server && go run main.go 2>&1 | grep DEBUG
```

### Monitor Goroutines
Add to server main:
```go
go func() {
  for range time.Tick(10 * time.Second) {
    fmt.Printf("Active goroutines: %d\n", runtime.NumGoroutine())
  }
}()
```

## Known Limitations & Workarounds

### 1. In-Memory Storage
- **Issue**: Data lost on server restart
- **Workaround**: Use database backend for persistence

### 2. No Message History
- **Issue**: Messages not stored
- **Workaround**: Client can maintain local message history

### 3. Single Server Instance
- **Issue**: No horizontal scaling
- **Workaround**: Use Redis/gRPC gateway for multiple instances

## Test Coverage Goals
- [ ] User Service: 100% coverage
- [ ] Room Service: 100% coverage
- [ ] Chat Service: 100% coverage
- [ ] Auth: 100% coverage
- [ ] AI Integration: 90% coverage
- [ ] Integration Tests: 80% coverage

## Continuous Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go test ./...
      - run: go vet ./...
```

## Conclusion
This testing guide ensures the gRPC chat application is robust, reliable, and production-ready. Follow these steps to validate your implementation.
