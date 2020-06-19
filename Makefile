.PHONY: dependency unittest test
BINARY=goruda

build: dependency
	go build -o ${BINARY} github.com/golangid/goruda/cmd/goruda

dependency:
	@echo "Installing dependency"
	@go mod vendor

test:
	@go test -v ./...

unittest: dependency
	@go test -v -short ./...
