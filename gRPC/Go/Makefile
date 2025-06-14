PROTO_DIR=api/proto
OUT_DIR=api/gen/v1

gen:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

start-client:
	go run ./cmd/client/main.go

start-server:
	go run ./cmd/server/main.go