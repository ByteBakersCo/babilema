#!/bin/bash

echo "Building..."
go build -v -o babilema ./cmd/babilema/main.go

echo "Successfully built the project, run it by executing ./babilema"
