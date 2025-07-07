# gRPC Sample Project

This project demonstrates a clean, maintainable gRPC setup in Go using idiomatic project structure, code generation, and modular design.

## 📁 Project Structure

```text
.
├── api/                # Protobuf definitions and generated code
│   ├── proto/          # .proto files
│   │   └── greeter.proto
│   └── gen/            # Generated Go code from .proto files
│       └── v1/
│           ├── greeter.pb.go
│           └── greeter_grpc.pb.go
├── cmd/                # Main entry points for server and client
│   ├── server/         
│   │   └── main.go     # gRPC server
│   └── client/
│       └── main.go     # gRPC client
├── internal/           # Business logic / core implementation
│   └── greeter/
│       └── service.go
├── go.mod              # Go module definition
├── go.sum              # Go dependencies lock file
├── Makefile            # CLI commands (code generation, etc.)
└── README.md           # Project documentation
```

## ⚙️ Setup Instructions

### 1. Install Dependencies

Ensure you have Go and Protocol Buffers installed.

#### Install Go

[https://go.dev/doc/install](https://go.dev/doc/install)

#### Install Protocol Buffers Compiler (`protoc`)

```bash
# Ubuntu
sudo apt install -y protobuf-compiler

# Mac
brew install protobuf
```

#### Install gRPC Plugins for Go

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure `$GOPATH/bin` is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 2. Generate gRPC Code

To generate the gRPC code from the `.proto` file, use the `Makefile`:

```bash
make gen
```

This will generate the code under the `api/gen/v1/` directory.

### 3. Install Dependencies

Run the following command to install any required dependencies:

```bash
make deps
```

### 4. Build the Project

To build the server and client binaries:

```bash
make build
```

### 5. Run the Server

Start the gRPC server:

```bash
make start-server
```

Or run the server manually:

```bash
go run cmd/server/main.go
```

Expected output:

```
2025/06/15 09:03:31 INFO Starting gRPC server on port=50051
```

### 6. Run the Client

Start the gRPC client:

```bash
make start-client
```

Or run the client manually:

```bash
go run cmd/client/main.go
```

Expected output:

```
Response from server: Hello, Krushnal!
```

## 🔐 Features Included

* ✅ Protobuf-based service definition
* ✅ gRPC server and client using modern Go APIs
* ✅ Clean layered structure (`cmd`, `internal`, `api`)
* ✅ Easy code generation via `Makefile`
* ✅ Logging with `slog`
* ✅ Dependency management with Go modules

## 📄 License

This project is open-sourced under the [MIT License](../../../LICENSE).