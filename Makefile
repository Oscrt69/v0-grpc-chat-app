.PHONY: help proto build-server build-client run-server run-client run-both clean test

help:
	@echo "gRPC AI Chat - Available Commands"
	@echo "=================================="
	@echo "make proto          - Generate proto files"
	@echo "make build-server   - Build server binary"
	@echo "make build-client   - Build client binary"
	@echo "make run-server     - Run server"
	@echo "make run-client     - Run client"
	@echo "make run-both       - Run server and client (requires tmux)"
	@echo "make clean          - Clean build artifacts"
	@echo "make test           - Run tests"
	@echo "make deps           - Download dependencies"

# Generate proto files
proto:
	@echo "Generating proto files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/chat.proto
	@echo "✅ Proto files generated"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies downloaded"

# Build server
build-server: proto
	@echo "Building server..."
	mkdir -p bin
	cd cmd/server && go build -o ../../bin/server
	@echo "✅ Server built: bin/server"

# Build client
build-client: proto
	@echo "Building client..."
	mkdir -p bin
	cd cmd/client && go build -o ../../bin/client
	@echo "✅ Client built: bin/client"

# Build both
build: build-server build-client
	@echo "✅ All binaries built"

# Run server
run-server: proto
	@echo "🚀 Starting server..."
	cd cmd/server && go run main.go

# Run client
run-client: proto
	@echo "🚀 Starting client..."
	cd cmd/client && go run main.go

# Run server and client (requires tmux)
run-both: proto
	@echo "Starting server and clients..."
	tmux new-session -d -s grpc-chat
	tmux send-keys -t grpc-chat "cd cmd/server && go run main.go" Enter
	tmux split-window -t grpc-chat -h
	tmux send-keys -t grpc-chat "sleep 1 && cd cmd/client && go run main.go" Enter
	tmux attach -t grpc-chat

# Clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f cmd/server/server cmd/client/client
	@echo "✅ Cleaned"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "✅ Tests completed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...
	@echo "✅ Linting completed"

# Install tools
install-tools:
	@echo "Installing development tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Tools installed"

# Setup development environment
setup: deps install-tools proto
	@echo "✅ Development environment setup complete"
