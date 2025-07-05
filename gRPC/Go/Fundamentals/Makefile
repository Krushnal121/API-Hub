PROTO_DIR=api/proto
OUT_DIR=api/gen/v1

.PHONY: gen clean start-server start-client test deps

# Generate gRPC code from proto files
gen:
	@echo "Generating gRPC code..."
	@mkdir -p $(OUT_DIR)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "Code generation complete!"

# Install dependencies
deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Start the gRPC server
start-server:
	@echo "Starting gRPC server..."
	go run cmd/server/main.go

# Run the gRPC client
start-client:
	@echo "Running gRPC client..."
	go run cmd/client/main.go

# Clean generated files
clean:
	rm -rf api/gen/
	go clean

# Build binaries
build:
	@echo "Building server..."
	go build -o bin/server cmd/server/main.go
	@echo "Building client..."
	go build -o bin/client cmd/client/main.go

# Generate self-signed certificates for TLS (optional)
gen-certs:
	@mkdir -p certs
	openssl req -newkey rsa:4096 -nodes -sha256 -keyout certs/server.key \
		-x509 -days 365 -out certs/server.crt \
		-subj "/C=US/ST=CA/L=San Francisco/O=gRPC Demo/OU=IT/CN=localhost" \
		-addext "subjectAltName = DNS:localhost,IP:127.0.0.1"
	@echo "Self-signed certificates generated in certs/"