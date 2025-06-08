
BINARY_CLIENT=build/gophkeeper-client
BINARY_SERVER=build/gophkeeper-server

PROTO_FILES=proto/keeper.proto
GEN_DIR=gen

all: build

build: generate
	mkdir -p build
	go build -o ${BINARY_CLIENT} ./client/cmd
	go build -o ${BINARY_SERVER} ./server/cmd

clean:
	rm -rf build/*

test-unit:
	go test ./... -coverprofile=coverage.out

test-integration:
	go test ./tests/integration/...

run-server:
	./${BINARY_SERVER}

run-client:
	./${BINARY_CLIENT} --help

generate:
	@if ! [ -x "$$(command -v protoc)" ]; then \
		echo "protoc not found. Please install protobuf compiler."; \
		exit 1; \
	fi
	protoc \
        --go_out=${GEN_DIR} \
        --go_opt=module=github.com/dvkhr/gophkeeper \
        --go-grpc_out=${GEN_DIR} \
        --go-grpc_opt=module=github.com/dvkhr/gophkeeper \
        ${PROTO_FILES}