APP_NAME := smart-card-reader
BIN_DIR := bin

.PHONY: run build build-mac-arm64 build-win-x64 test clean

run:
	go run ./cmd/card-reader

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/card-reader

build-mac-arm64:
	mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(APP_NAME)-macos-arm64 ./cmd/card-reader

build-win-x64:
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-win-x64.exe ./cmd/card-reader

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
