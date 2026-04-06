# Troubleshooting Guide

Solusi untuk masalah umum yang mungkin dihadapi.

## Installation Issues

### ❌ "protoc: command not found"

**Problem:** Protoc compiler tidak terinstall

**Solution:**
- **macOS:**
  ```bash
  brew install protobuf
  ```

- **Ubuntu/Debian:**
  ```bash
  sudo apt-get install protobuf-compiler
  ```

- **Windows:**
  - Download dari: https://github.com/protocolbuffers/protobuf/releases
  - Extract dan add ke PATH
  - Or gunakan Chocolatey:
    ```bash
    choco install protoc
    ```

**Verify:**
```bash
protoc --version
# Expected: libprotoc 3.x.x
```

---

### ❌ "go: command not found"

**Problem:** Go tidak terinstall

**Solution:**
- Download dari https://golang.org/dl
- Install untuk OS Anda
- Add ke PATH environment variable

**Verify:**
```bash
go version
# Expected: go version go1.21+ ...
```

---

### ❌ "no such file or directory: proto/chat.proto"

**Problem:** Menjalankan command di direktori salah

**Solution:**
```bash
# Pastikan di root project directory
cd grpc-ai-chat

# Verify file exists
ls proto/chat.proto

# Then run proto generation
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/chat.proto
```

---

### ❌ "cannot load module: missing go.sum entry"

**Problem:** Go modules tidak sinkron

**Solution:**
```bash
go mod tidy
go mod download
```

---

## Runtime Issues

### ❌ "listen tcp :50051: bind: address already in use"

**Problem:** Port 50051 sudah digunakan

**Solution - Opsi 1: Kill existing process**
```bash
# Find process using port 50051
lsof -i :50051

# Kill the process
kill -9 <PID>

# Or on Windows
netstat -ano | findstr :50051
taskkill /PID <PID> /F
```

**Solution - Opsi 2: Use different port**
Edit `cmd/server/main.go`:
```go
const port = ":50052"  // Change from 50051
```

Edit `cmd/client/main.go`:
```go
const serverAddress = "localhost:50052"  // Match server port
```

---

### ❌ "connection refused" or "cannot connect to server"

**Problem:** Server tidak running atau port salah

**Solution:**
1. **Ensure server is running:**
   ```bash
   # Terminal 1
   cd cmd/server
   go run main.go
   # Should print: ✅ Server listening at [::]:50051
   ```

2. **Check from another terminal:**
   ```bash
   # Terminal 2
   telnet localhost 50051
   # Should connect (Ctrl+C to exit)
   ```

3. **If connecting from different machine:**
   ```bash
   # Replace localhost with server IP
   const serverAddress = "192.168.1.100:50051"
   ```

4. **Check firewall:**
   ```bash
   # macOS
   sudo lsof -i :50051
   
   # Ubuntu
   sudo ufw allow 50051
   
   # Windows
   netsh advfirewall firewall add rule name="gRPC" dir=in action=allow protocol=tcp localport=50051
   ```

---

### ❌ "rpc error: code = Unauthenticated desc = token tidak valid"

**Problem:** Invalid atau missing token

**Solutions:**

1. **Not logged in yet:**
   ```bash
   # Must login first
   > login alice password123
   ```

2. **Token expired (after 24 hours):**
   ```bash
   # Login again to get new token
   > logout
   > login alice password123
   ```

3. **Wrong username/password:**
   ```bash
   # Check credentials
   > login alice password123  # Correct
   
   # Default users:
   # alice, bob, charlie (all with password123)
   ```

4. **Token modified/corrupted:**
   - Logout and login again
   - Or restart client

---

### ❌ "rpc error: code = NotFound desc = room tidak ditemukan"

**Problem:** Room ID salah atau room sudah dihapus

**Solutions:**

1. **Verify room exists:**
   ```bash
   > list_rooms
   # See all available rooms with their IDs
   ```

2. **Use correct room ID:**
   ```bash
   > join_room 550e8400-e29b-41d4-a716-446655440000
   # Use exact ID from list_rooms
   ```

3. **Room was deleted:**
   - Last member left room
   - Room auto-deleted
   - Create new room:
     ```bash
     > create_room "New Room" "" true
     ```

---

### ❌ "rpc error: code = AlreadyExists desc = username sudah terdaftar"

**Problem:** Username sudah dipakai

**Solutions:**

1. **Use different username:**
   ```bash
   > register john password123 "John Doe"
   ```

2. **Login with existing username:**
   ```bash
   > login john password123
   ```

---

### ❌ AI responses not working

**Problem:** Grok API bukan returning responses

**Solutions:**

1. **Check API key:**
   ```bash
   # Set in .env or environment
   export GROK_API_KEY=your-key-here
   
   # Verify
   echo $GROK_API_KEY
   ```

2. **Verify room has AI enabled:**
   ```bash
   > list_rooms
   # Look for 🤖 emoji next to room name
   ```

3. **Check API connectivity:**
   ```bash
   # Test with curl
   curl -X POST https://api.x.ai/v1/chat/completions \
     -H "Authorization: Bearer YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"model":"grok-2","messages":[{"role":"user","content":"Hi"}]}'
   ```

4. **Check logs:**
   - Server logs akan show API errors
   - Look for "Grok API error" messages

---

## Chat Issues

### ❌ "Can't see messages from other users"

**Problem:** Not in same room or stream not working

**Solutions:**

1. **Ensure same room:**
   ```bash
   # User A
   > create_room "Team" "" false
   > list_rooms  # Note room ID
   > join_room <id>
   
   # User B
   > list_rooms
   > join_room <same-id>
   ```

2. **Check connection:**
   ```bash
   > chat
   # Should print: 📢 Starting chat...
   # If nothing, connection issue
   ```

3. **Restart chat stream:**
   ```bash
   # Exit chat
   exit
   
   # Rejoin
   > join_room <room-id>
   > chat
   ```

---

### ❌ "Messages not broadcasting to all users"

**Problem:** Other users not receiving messages

**Solutions:**

1. **Check if all users joined same room:**
   - Get room ID from list_rooms
   - All users must use same room ID

2. **Verify stream is active:**
   - Each user should see "📢 Starting chat..."
   - If not, join/rejoin room

3. **Check for errors:**
   - Look at server terminal for error messages
   - Look at client for connection errors

4. **Network issues:**
   ```bash
   # Test connectivity
   ping <server-ip>
   
   # Test TCP connection
   telnet localhost 50051
   ```

---

### ❌ "Chat stream closes unexpectedly"

**Problem:** Disconnect from stream without warning

**Solutions:**

1. **Check for timeouts:**
   - Long idle periods might timeout
   - Send message to keep alive

2. **Check network:**
   - WiFi disconnect
   - Network interruption
   - Run on stable network

3. **Server crash:**
   ```bash
   # Check server terminal
   # Look for panic or error messages
   # Restart server
   cd cmd/server && go run main.go
   ```

4. **Too many concurrent users:**
   - Each stream = goroutine
   - OS limit on file descriptors
   - Run on machine with resources

---

## Performance Issues

### ❌ "Chat lag or slow messages"

**Solutions:**

1. **Network latency:**
   ```bash
   # Check ping time
   ping <server-ip>
   # Should be < 50ms
   ```

2. **Server overloaded:**
   ```bash
   # Check server logs
   # Look for processing time messages
   ```

3. **Too many users:**
   - Current: ~100 concurrent users per server
   - For more, need database + load balancer

4. **AI response slow:**
   - Grok API calls can take 2-5 seconds
   - This is normal - processing continues

---

### ❌ "Memory usage high"

**Problem:** Application using lots of memory

**Solutions:**

1. **Check number of active streams:**
   ```bash
   # Each stream = ~1MB memory
   # More streams = more memory
   ```

2. **Long-running instance:**
   - No automatic cleanup of expired sessions
   - Restart server periodically:
     ```bash
     # Stop server (Ctrl+C)
     # Start new instance
     go run cmd/server/main.go
     ```

3. **For production:**
   - Implement periodic cleanup
   - Use database instead of memory
   - Add memory limits

---

## Compilation Issues

### ❌ "cannot find package"

**Problem:** Import path wrong atau package missing

**Solutions:**

1. **Check import paths:**
   ```go
   // Should match go.mod module path
   import "grpc-ai-chat/internal/models"
   ```

2. **Download dependencies:**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Regenerate proto files:**
   ```bash
   protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/chat.proto
   ```

---

### ❌ "syntax error" or "undefined" in generated files

**Problem:** Proto files not regenerated after changes

**Solution:**

1. **Modify proto file**
2. **Regenerate:**
   ```bash
   protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/chat.proto
   ```
3. **Rebuild:**
   ```bash
   go run cmd/server/main.go
   ```

---

### ❌ "version incompatibility" warning

**Problem:** Go module version conflicts

**Solution:**
```bash
go mod tidy
go get -u ./...
```

---

## Debugging Tips

### Enable Debug Output

**Add logging to code:**
```go
log.Printf("[DEBUG] User login: %s", username)
log.Printf("[DEBUG] Room created: %s", roomID)
log.Printf("[DEBUG] Stream message: %v", msg)
```

**Or compile with debug flag:**
```bash
go run -v cmd/server/main.go
# Shows build details
```

### Use grpcurl

**Install:**
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Test services:**
```bash
# List services
grpcurl -plaintext localhost:50051 list

# Call RPC
grpcurl -plaintext -d '{"username":"alice","password":"password123"}' \
  localhost:50051 chat.v1.UserService.Login
```

### Monitor Network Traffic

**macOS/Linux:**
```bash
# Monitor gRPC traffic
tcpdump -i lo0 tcp port 50051

# Or use wireshark
# Filter: tcp.port == 50051
```

### Check Resource Usage

```bash
# Monitor memory & CPU
top  # macOS/Linux
Get-Process -ProcessName go  # Windows

# Show open files
lsof -p <PID>
```

---

## Database/State Issues

### ❌ "User data lost after restart"

**This is expected!** Current implementation uses in-memory storage.

**Solution for production:**
- Implement persistent database
- Use PostgreSQL, MongoDB, or similar
- Store users, sessions, and messages

**Workaround for testing:**
- Don't restart server
- Or recreate users after restart

---

### ❌ "Duplicate users sometimes created"

**Problem:** Race condition on registration

**Solution - Current:** (Not a real issue with current implementation)

**Solution for production:**
- Use database with unique constraints
- Add validation at database level

---

## Environment Variable Issues

### ❌ "GROK_API_KEY not set"

**Problem:** AI responses not working

**Solutions:**

1. **Set in .env file:**
   ```
   # .env
   GROK_API_KEY=your-key-here
   ```

2. **Set in environment:**
   ```bash
   export GROK_API_KEY=your-key-here
   ```

3. **Verify:**
   ```bash
   echo $GROK_API_KEY
   # Should print your key
   ```

4. **For Windows PowerShell:**
   ```powershell
   $env:GROK_API_KEY="your-key-here"
   ```

---

## Getting Help

### 1. Check Documentation
- **QUICKSTART.md** - Common use cases
- **API_REFERENCE.md** - API details
- **ARCHITECTURE.md** - System design

### 2. Check Logs
- **Server logs:** Look for error messages
- **Client logs:** Check output messages
- **System logs:** Check OS event viewer

### 3. Test in Isolation
```bash
# Test connectivity
telnet localhost 50051

# Test with grpcurl
grpcurl -plaintext localhost:50051 list

# Test single RPC
grpcurl -plaintext -d '{"username":"alice","password":"password123"}' \
  localhost:50051 chat.v1.UserService.Login
```

### 4. Reduce Complexity
- Test with single client first
- Test with simple messages
- Remove AI integration if not needed

### 5. Restart Everything
```bash
# Kill processes
pkill -f "go run"

# Clear any stuck connections
# (Usually not necessary)

# Restart server
cd cmd/server && go run main.go

# Restart client
cd cmd/client && go run main.go
```

---

## Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| "bind: address already in use" | Port in use | Kill process or use different port |
| "connection refused" | Server not running | Start server first |
| "token tidak valid" | Missing/expired token | Login again |
| "room tidak ditemukan" | Wrong room ID | Use ID from list_rooms |
| "username sudah terdaftar" | Duplicate username | Use different name |
| "context deadline exceeded" | Operation timeout | Check network, increase timeout |
| "rpc error: code = Internal" | Server error | Check server logs, restart |

---

## Still Having Issues?

1. **Check documentation:** README.md, QUICKSTART.md, ARCHITECTURE.md
2. **Review code:** Look at implementation, understand flow
3. **Add debugging:** Log important operations
4. **Test components:** Test each service separately
5. **Simplify:** Reduce to minimal reproduction case
6. **Restart:** Fresh start often solves issues

---

Happy debugging! 🐛
