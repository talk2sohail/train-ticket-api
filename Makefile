WORK_DIR := $(shell pwd)

GO_OUT_DIR := $(WORK_DIR)/internal/common/genproto/ticket/
PROTO_DIR := $(WORK_DIR)/proto/
PROTO_FILES := $(PROTO_DIR)/*.proto 

run: 
	@go run cmd/*.go

run-ticket:
	@go run internal/ticket/main.go

gen:
	@protoc -I=$(PROTO_DIR) \
		--go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

tidy:
	@echo "Updating Go module..."
	@go mod tidy
	@echo "Go module updated."

format:
	@echo "Formatting Go files..."
	@go fmt ./...
	@echo "Go files formatted."

download:
	@echo "Downloading Go dependencies..."
	@go mod download
	@echo "Go dependencies downloaded."

clean :
	@echo "Cleaning up generated files..."
	@rm -rf $(GO_OUT_DIR)/*
	@echo "Clean up completed."
