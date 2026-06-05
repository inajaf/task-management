SHELL := /bin/bash

run:
	go run ./cmd/task-management/main.go

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -v ./...