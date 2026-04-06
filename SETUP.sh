#!/bin/bash

# gRPC AI Chat - Setup Script
# Run this script to setup the development environment

set -e

echo "🚀 gRPC AI Chat - Setup Script"
echo "=============================="
echo ""

# Check Go installation
echo "✓ Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later"
    echo "   Visit: https://golang.org/dl"
    exit 1
fi
go version

echo ""
echo "✓ Checking protoc installation..."
if ! command -v protoc &> /dev/null; then
    echo "⚠️  protoc is not installed. Installing..."
    # Instructions for different OS
    echo "   macOS: brew install protobuf"
    echo "   Ubuntu: sudo apt-get install protobuf-compiler"
    echo "   Windows: Download from https://github.com/protocolbuffers/protobuf/releases"
    echo "   Then continue with this script"
    exit 1
fi
protoc --version

echo ""
echo "✓ Installing Go protoc plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

echo ""
echo "✓ Downloading dependencies..."
go mod download
go mod tidy

echo ""
echo "✓ Generating proto files..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/chat.proto

echo ""
echo "✓ Creating .env file..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "   Created .env file"
    echo "   ⚠️  Remember to add your GROK_API_KEY to .env"
else
    echo "   .env already exists"
fi

echo ""
echo "✓ Creating bin directory..."
mkdir -p bin

echo ""
echo "=============================="
echo "✅ Setup Complete!"
echo "=============================="
echo ""
echo "📝 Next steps:"
echo ""
echo "1. Edit .env and add your Grok API key (optional)"
echo "   GROK_API_KEY=your-key-here"
echo ""
echo "2. Start the server:"
echo "   cd cmd/server && go run main.go"
echo ""
echo "3. In another terminal, start the client:"
echo "   cd cmd/client && go run main.go"
echo ""
echo "4. Login with default credentials:"
echo "   > login alice password123"
echo ""
echo "5. Create a room and chat:"
echo "   > create_room \"Hello World\" \"\" true"
echo "   > list_rooms"
echo "   > join_room <room-id>"
echo "   > chat"
echo ""
echo "📚 Documentation:"
echo "   - QUICKSTART.md - Get started in 5 minutes"
echo "   - README.md - Complete guide"
echo "   - ARCHITECTURE.md - System design"
echo ""
