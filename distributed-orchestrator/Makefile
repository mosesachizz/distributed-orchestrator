.PHONY: build test deploy clean

# Build targets
build:
	@echo "Building orchestrator..."
	@go build -o bin/master ./cmd/master
	@go build -o bin/worker ./cmd/worker
	@go build -o bin/cli ./cmd/cli

build-docker:
	@echo "Building Docker images..."
	@docker build -f Dockerfile.master -t orchestrator-master:latest .
	@docker build -f Dockerfile.worker -t orchestrator-worker:latest .

test:
	@echo "Running tests..."
	@go test ./... -v

# Deployment targets
deploy-dev: build-docker
	@echo "Deploying to development environment..."
	@docker-compose up -d

deploy-prod:
	@echo "Deploying to production environment..."
	@kubectl apply -f deployments/kubernetes/

# Code generation
proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go-grpc_out=. pkg/rpc/proto/service.proto

# Cleanup
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@docker-compose down
	@docker rmi orchestrator-master orchestrator-worker 2>/dev/null || true

# Development
dev:
	@echo "Starting development environment..."
	@docker-compose up -d etcd
	@sleep 2
	@go run ./cmd/master &
	@go run ./cmd/worker &

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build binaries"
	@echo "  build-docker - Build Docker images"
	@echo "  test         - Run tests"
	@echo "  deploy-dev   - Deploy to development"
	@echo "  deploy-prod  - Deploy to production"
	@echo "  proto        - Generate protobuf code"
	@echo "  clean        - Clean up"
	@echo "  dev          - Start development environment"