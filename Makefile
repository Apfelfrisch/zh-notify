it: compile-sql build-native

arm-build:
	@echo "Building for Arm..."
	@env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build .

build-native:
	@echo "Building native..."
	@go build .

compile-sql:
	@sqlc generate

test:
	@go test ./...
