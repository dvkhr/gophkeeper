
BINARY_CLIENT=build/gophkeeper-client
BINARY_SERVER=build/gophkeeper-server

PROTO_FILES=proto/keeper.proto
GEN_DIR=.

# Настройки тестов
TEST_PKG=./...
TEST_FLAGS=-v -race
COVER_PROFILE=coverage.out
COVER_HTML=coverage.html

# Документация
DOC_DIR=docs
DOC_PORT=6060

# Версия
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILDDATE=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')


all: build

build: generate
	mkdir -p build
	go build -ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILDDATE}" \
		-o ${BINARY_CLIENT} ./client/cmd
	go build -o ${BINARY_SERVER} ./server/cmd
	@echo "Сборка завершена:"
	@echo "  Клиент: ${BINARY_CLIENT}"
	@echo "  Сервер: ${BINARY_SERVER}"
	@echo "  Версия: ${VERSION}, Сборка: ${BUILDDATE}"

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
	go test -p 1 -race -coverprofile=${COVER_PROFILE} -covermode=atomic ${TEST_PKG}
	go tool cover -html=${COVER_PROFILE} -o ${COVER_HTML}
	@echo "Отчёт о покрытии сохранён в ${COVER_HTML}"

# Сборка клиента для всех платформ
build-client-all:
	@echo "Сборка клиента для всех платформ..."
	GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILDDATE}" -o build/gophkeeper-client-linux-amd64 ./client/cmd
	GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILDDATE}" -o build/gophkeeper-client-darwin-amd64 ./client/cmd
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILDDATE}" -o build/gophkeeper-client-windows-amd64.exe ./client/cmd
	@echo "Клиент собран для Linux, macOS, Windows (amd64)."

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

# Генерация документации через локальный сервер
doc:
	@mkdir -p ${DOC_DIR}
	@echo "Запуск локального godoc сервера для генерации HTML..."
	@godoc -http=:${DOC_PORT} &
	@sleep 2
	@echo "Скачиваем HTML-страницы..."
	@curl -s -o ${DOC_DIR}/auth.html http://localhost:${DOC_PORT}/pkg/github.com/dvkhr/gophkeeper/server/internal/auth/
	@curl -s -o ${DOC_DIR}/repository.html http://localhost:${DOC_PORT}/pkg/github.com/dvkhr/gophkeeper/server/internal/repository/
	@curl -s -o ${DOC_DIR}/db.html http://localhost:${DOC_PORT}/pkg/github.com/dvkhr/gophkeeper/server/internal/db/
	@echo "Останавливаем godoc сервер..."
	@PID=$(lsof -t -i:${DOC_PORT}); \
	if [ -n "$$PID" ]; then \
		kill $$PID || echo "Не удалось остановить godoc"; \
	else \
		echo "godoc не запущен на порту ${DOC_PORT}"; \
	fi
	@echo "HTML-документация сохранена в ${DOC_DIR}/"