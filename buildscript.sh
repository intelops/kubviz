#!/bin/bash

set -e

# Install dependencies
echo "Installing dependencies..."
go mod download

# Build the project
echo "Building the project..."
go build ./...

# Run tests
echo "Running tests..."
go test ./...

echo "Build and test completed successfully."

