.PHONY: all build clean update test

all: clean update build test

build:
	@echo "Building the project..."
	go build -o ./bin/server-beta ./cmd/server-beta

clean:
	@echo "Cleaning..."
	rm -f ./bin/server-beta

update:
	@echo "Updating dependencies..."
	go get -u all
	go mod tidy

test:
	@echo "Running tests..."
	CGO_ENABLED=1 go test ./...go test ./...