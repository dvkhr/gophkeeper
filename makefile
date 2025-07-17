
BINARY_CLIENT=build/gophkeeper-client
BINARY_SERVER=build/gophkeeper-server

PROTO_FILES=proto/keeper.proto
GEN_DIR=.

# Настройки тестов
TEST_PKG=./...
TEST_FLAGS=-v -race
COVER_PROFILE=coverage.out
COVER_HTML=coverage.html

all: build

build: generate
	mkdir -p build
	go build -o ${BINARY_CLIENT} ./client/cmd
	go build -o ${BINARY_SERVER} ./server/cmd

clean:
	rm -rf build/*
	rm -f ${COVER_PROFILE} ${COVER_HTML}


test: test-unit test-integration

test-unit:
	go test ${TEST_FLAGS} ${TEST_PKG}

test-integration:
	@echo "Запуск интеграционных тестов..."
	go test ${TEST_FLAGS} ./tests/integration/...

cover:
	go test -race -coverprofile=${COVER_PROFILE} -covermode=atomic ${TEST_PKG}
	go tool cover -html=${COVER_PROFILE} -o ${COVER_HTML}
	@echo "Отчёт о покрытии сохранён в ${COVER_HTML}"

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

		