#!/bin/bash

echo "Downloading dependencies..."
go mod download && go mod verify

echo "Running tests..."
go test -v ./...

echo "Building binaries..."
go build -v -o babilema ./cmd/babilema/main.go

echo "Successfully built the project, run it by executing ./babilema"
