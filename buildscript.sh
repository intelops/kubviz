#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Setting up Go environment"

# Set the Go version. Adjust this to match the version you need.
GO_VERSION="1.19"

# Download and install the specified Go version
echo "Installing Go ${GO_VERSION}"
curl -sSL https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz -o go${GO_VERSION}.linux-amd64.tar.gz
tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
export PATH="/usr/local/go/bin:$PATH"

# Verify Go installation
go version

# Navigate to the repository root. Adjust if your project structure is different.
cd "$(dirname "$0")/.."

echo "Installing dependencies"
# Install Go modules
go mod download

echo "Building the project"
# Build the project. Adjust the build command if needed.
go build ./...

echo "Running tests"
# Run tests
go test ./...

echo "Build and tests completed successfully"

