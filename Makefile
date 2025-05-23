WORK_DIR := $(shell pwd)

GO_OUT_DIR := $(WORK_DIR)/internal/common/genproto/ticket/
PROTO_DIR := $(WORK_DIR)/proto/
PROTO_FILES := $(PROTO_DIR)/*.proto 

run-example: 
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

test:
	@echo "Running tests with coverage..."
	@go test -timeout 30s -run ^TestUnit_*  -coverprofile coverage.out ./...
	@echo "Coverage report generated in coverage.out."

clean :
	@echo "Cleaning up generated files..."
	@rm -rf $(GO_OUT_DIR)/*
	@rm -rf $(WORK_DIR)/mock/*
	@echo "Clean up completed."

mocks:
	@echo "Generating mocks for TicketService..."
	@mockgen -destination=./mock/ticket_service_mock.go -package=mock github.com/talk2sohail/train-ticket-api/internal/ticket/types TicketService
