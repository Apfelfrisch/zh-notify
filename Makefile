arm-build:
	@echo "Building for Arm..."
	@env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build .

test:
	@go test ./...
