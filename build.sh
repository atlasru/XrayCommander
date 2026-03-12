#!/bin/bash

# XrayCommander Build Script

set -e

echo "==================================="
echo "  XrayCommander Build Script"
echo "==================================="
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go 1.21 or higher: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# Download dependencies
echo ""
echo "Downloading dependencies..."
go mod download
go mod tidy
echo "Dependencies downloaded"

# Build
echo ""
echo "Building XrayCommander..."
go build -o xraycommander ./cmd/xraycommander
echo "Build successful"

echo ""
echo "==================================="
echo "Build complete!"
echo ""
echo "Run the application:"
echo "  ./xraycommander"
echo "==================================="
